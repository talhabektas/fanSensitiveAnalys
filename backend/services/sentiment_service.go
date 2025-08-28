package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"github.com/go-resty/resty/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SentimentService struct {
	collection  *mongo.Collection
	client      *resty.Client
	groqService *GroqService
}

type HuggingFaceResponse struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type HuggingFaceRequest struct {
	Inputs []string `json:"inputs"`
}

const (
	HuggingFaceURL = "https://api-inference.huggingface.co/models/savasy/bert-base-turkish-sentiment-cased"
	MaxRetries     = 3
	RetryDelay     = time.Second * 2
)

func NewSentimentService() *SentimentService {
	client := resty.New().
		SetTimeout(30*time.Second).
		SetRetryCount(MaxRetries).
		SetRetryWaitTime(RetryDelay)

	if config.AppConfig.HuggingFaceToken != "" {
		client.SetAuthToken(config.AppConfig.HuggingFaceToken)
	}

	return &SentimentService{
		collection:  config.GetCollection("sentiments"),
		client:      client,
		groqService: NewGroqService(),
	}
}

func (ss *SentimentService) AnalyzeText(text string) (*models.SentimentResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if len(text) < 5 {
		return nil, fmt.Errorf("text too short for analysis")
	}

	if len(text) > 5000 {
		text = text[:5000]
	}

	// Hibrit analiz: HuggingFace + Grok AI
	return ss.HybridAnalyze(text)
}

func (ss *SentimentService) HybridAnalyze(text string) (*models.SentimentResult, error) {
	cleanText := ss.preprocessText(text)
	startTime := time.Now()

	// HuggingFace analizi
	hfResult, hfErr := ss.analyzeWithHuggingFace(cleanText)
	
	// Grok AI analizi (paralel olarak)
	groqResult, groqErr := ss.groqService.EnhancedSentimentAnalysis(cleanText)

	processingTime := time.Since(startTime).Seconds()

	// Her iki analiz de başarısız olduysa
	if hfErr != nil && groqErr != nil {
		return nil, fmt.Errorf("both analyses failed - HF: %v, Groq: %v", hfErr, groqErr)
	}

	// Sadece HuggingFace başarılı
	if hfResult != nil && groqErr != nil {
		log.Printf("Using HuggingFace only (Groq failed): %s (%.3f) in %.2fs", hfResult.Label, hfResult.Confidence, processingTime)
		hfResult.ModelUsed = "hf-only"
		return hfResult, nil
	}

	// Sadece Grok başarılı  
	if groqResult != nil && hfErr != nil {
		log.Printf("Using Groq only (HF failed): %s (%.3f) in %.2fs", groqResult.Enhanced.Label, groqResult.Enhanced.Confidence, processingTime)
		groqResult.Enhanced.ModelUsed = "groq-only"
		return groqResult.Enhanced, nil
	}

	// Her ikisi de başarılı - hibrit karar
	if hfResult != nil && groqResult != nil {
		hybridResult := ss.combineResults(hfResult, groqResult)
		log.Printf("Hybrid analysis: %s (%.3f confidence, HF: %.3f, Groq: %.3f) in %.2fs", 
			hybridResult.Label, hybridResult.Confidence, hfResult.Confidence, groqResult.Enhanced.Confidence, processingTime)
		return hybridResult, nil
	}

	return nil, fmt.Errorf("unexpected error in hybrid analysis")
}

func (ss *SentimentService) analyzeWithHuggingFace(text string) (*models.SentimentResult, error) {
	requestBody := HuggingFaceRequest{
		Inputs: []string{text},
	}

	var responses [][]HuggingFaceResponse
	resp, err := ss.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(&responses).
		Post(HuggingFaceURL)

	if err != nil {
		return nil, fmt.Errorf("huggingface API error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("huggingface API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if len(responses) == 0 || len(responses[0]) == 0 {
		return nil, fmt.Errorf("empty response from huggingface API")
	}

	results := responses[0]
	bestResult := ss.findBestResult(results)

	return &models.SentimentResult{
		Label:       ss.normalizeLabel(bestResult.Label),
		Score:       bestResult.Score,
		Confidence:  bestResult.Score,
		ModelUsed:   "savasy/bert-base-turkish-sentiment-cased",
		ProcessedAt: time.Now(),
	}, nil
}

func (ss *SentimentService) combineResults(hf *models.SentimentResult, groq *GroqAnalysisResult) *models.SentimentResult {
	// Güven skorlarına göre ağırlıklı karar
	hfWeight := hf.Confidence
	groqWeight := groq.Enhanced.Confidence
	
	// Eğer her iki model de aynı sentiment'i veriyorsa, güven skorunu artır
	if hf.Label == groq.Enhanced.Label {
		combinedConfidence := (hfWeight + groqWeight) / 2.0
		if combinedConfidence > 0.95 {
			combinedConfidence = 0.95 // Max limit
		}
		
		return &models.SentimentResult{
			Label:       hf.Label,
			Score:       (hf.Score + groq.Enhanced.Score) / 2.0,
			Confidence:  combinedConfidence + 0.1, // Consensus bonus
			ModelUsed:   "hybrid-consensus",
			ProcessedAt: time.Now(),
		}
	}

	// Farklı sentiment'ler - daha güvenilir olanı seç
	if hfWeight > groqWeight {
		return &models.SentimentResult{
			Label:       hf.Label,
			Score:       hf.Score,
			Confidence:  hf.Confidence,
			ModelUsed:   "hybrid-hf-primary",
			ProcessedAt: time.Now(),
		}
	} else {
		return &models.SentimentResult{
			Label:       groq.Enhanced.Label,
			Score:       groq.Enhanced.Score,
			Confidence:  groq.Enhanced.Confidence,
			ModelUsed:   "hybrid-groq-primary",
			ProcessedAt: time.Now(),
		}
	}
}

func (ss *SentimentService) AnalyzeBatch(texts []string) ([]*models.SentimentResult, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	batchSize := 10
	var allResults []*models.SentimentResult

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		results, err := ss.analyzeBatch(batch)
		if err != nil {
			log.Printf("Error analyzing batch %d-%d: %v", i, end, err)
			
			for range batch {
				allResults = append(allResults, nil)
			}
			continue
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func (ss *SentimentService) analyzeBatch(texts []string) ([]*models.SentimentResult, error) {
	var results []*models.SentimentResult
	
	// HuggingFace API batch desteği güvenilir değil, tekil istekler gönderelim
	for _, text := range texts {
		result, err := ss.analyzeSingle(text)
		if err != nil {
			log.Printf("Error analyzing text: %v", err)
			results = append(results, nil)
			continue
		}
		results = append(results, result)
		
		// Rate limiting için kısa bir bekleme
		time.Sleep(100 * time.Millisecond)
	}

	return results, nil
}

func (ss *SentimentService) analyzeSingle(text string) (*models.SentimentResult, error) {
	cleanText := ss.preprocessText(text)
	
	requestBody := HuggingFaceRequest{
		Inputs: []string{cleanText},
	}

	var responses [][]HuggingFaceResponse
	resp, err := ss.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(&responses).
		Post(HuggingFaceURL)

	if err != nil {
		return nil, fmt.Errorf("huggingface API error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("huggingface API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if len(responses) == 0 || len(responses[0]) == 0 {
		return nil, fmt.Errorf("empty response from huggingface API")
	}

	response := responses[0]
	bestResult := ss.findBestResult(response)
	
	return &models.SentimentResult{
		Label:       ss.normalizeLabel(bestResult.Label),
		Score:       bestResult.Score,
		Confidence:  bestResult.Score,
		ModelUsed:   "savasy/bert-base-turkish-sentiment-cased",
		ProcessedAt: time.Now(),
	}, nil
}

func (ss *SentimentService) SaveSentimentEnhanced(commentID primitive.ObjectID, teamID primitive.ObjectID, result *models.SentimentResult, text string) (*models.Sentiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Zaten sentiment var mı kontrol et
	var existingSentiment models.Sentiment
	err := ss.collection.FindOne(ctx, bson.M{"comment_id": commentID}).Decode(&existingSentiment)
	if err == nil {
		log.Printf("Sentiment already exists for comment %s, skipping", commentID.Hex())
		return &existingSentiment, nil
	}

	// Grok AI ile kategori ve anahtar kelime analizi
	category, keywords, categoryErr := ss.groqService.CategorizeComment(text)
	if categoryErr != nil {
		log.Printf("Category analysis failed: %v", categoryErr)
		category = "Genel"
		keywords = []string{}
	}

	// Gelişmiş analiz detayları
	analysisDetails := models.AnalysisDetails{
		Scores: map[string]float64{
			result.Label: result.Score,
		},
		Keywords:       keywords,
		Language:       "tr",
		ProcessingTime: 0,
		Category:       category,
		ToxicityScore:  0.0, // Groq servisinden gelecek
	}

	// Grok AI'dan detayları al
	if groqResult, err := ss.groqService.EnhancedSentimentAnalysis(text); err == nil && groqResult != nil {
		analysisDetails.ToxicityScore = groqResult.ToxicityScore
		analysisDetails.GroqSummary = groqResult.Summary
		if len(groqResult.Keywords) > 0 {
			analysisDetails.Keywords = groqResult.Keywords
		}
	}

	sentiment := &models.Sentiment{
		CommentID:       commentID,
		TeamID:          teamID,
		Label:           result.Label,
		Score:           result.Score,
		Confidence:      result.Confidence,
		ModelUsed:       result.ModelUsed,
		AnalysisDetails: analysisDetails,
		Metadata: models.SentimentMetadata{
			ProcessedBy: "hybrid-analysis",
			ProcessedAt: result.ProcessedAt,
			APIVersion:  "v2",
			RetryCount:  0,
		},
	}

	sentiment.BeforeCreate()

	_, err = ss.collection.InsertOne(ctx, sentiment)
	if err != nil {
		return nil, fmt.Errorf("error saving enhanced sentiment: %w", err)
	}

	log.Printf("Enhanced sentiment saved: %s (%s, %.3f confidence, category: %s)", 
		result.Label, result.ModelUsed, result.Confidence, category)
	return sentiment, nil
}

// Geriye uyumluluk için eski metod
func (ss *SentimentService) SaveSentiment(commentID primitive.ObjectID, teamID primitive.ObjectID, result *models.SentimentResult) (*models.Sentiment, error) {
	return ss.SaveSentimentEnhanced(commentID, teamID, result, "")
}

func (ss *SentimentService) GetSentimentStats() (*models.SentimentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":                nil,
				"total_analyzed":     bson.M{"$sum": 1},
				"avg_sentiment":      bson.M{"$avg": "$score"},
				"avg_confidence":     bson.M{"$avg": "$confidence"},
				"positive_count":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "POSITIVE"}}, 1, 0}}},
				"negative_count":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEGATIVE"}}, 1, 0}}},
				"neutral_count":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEUTRAL"}}, 1, 0}}},
				"high_confidence":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$gte": []interface{}{"$confidence", 0.8}}, 1, 0}}},
				"medium_confidence":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$and": []interface{}{bson.M{"$gte": []interface{}{"$confidence", 0.6}}, bson.M{"$lt": []interface{}{"$confidence", 0.8}}}}, 1, 0}}},
				"low_confidence":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$lt": []interface{}{"$confidence", 0.6}}, 1, 0}}},
			},
		},
	}

	cursor, err := ss.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating sentiment stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAnalyzed     int64   `bson:"total_analyzed"`
		AvgSentiment      float64 `bson:"avg_sentiment"`
		AvgConfidence     float64 `bson:"avg_confidence"`
		PositiveCount     int64   `bson:"positive_count"`
		NegativeCount     int64   `bson:"negative_count"`
		NeutralCount      int64   `bson:"neutral_count"`
		HighConfidence    int64   `bson:"high_confidence"`
		MediumConfidence  int64   `bson:"medium_confidence"`
		LowConfidence     int64   `bson:"low_confidence"`
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("error decoding stats: %w", err)
	}

	stats := &models.SentimentStats{
		SentimentBreakdown: make(map[string]int64),
		ModelPerformance:   make(map[string]models.ModelPerformance),
	}

	if len(result) > 0 {
		r := result[0]
		stats.TotalAnalyzed = r.TotalAnalyzed
		stats.OverallSentiment = r.AvgSentiment
		stats.SentimentBreakdown["POSITIVE"] = r.PositiveCount
		stats.SentimentBreakdown["NEGATIVE"] = r.NegativeCount
		stats.SentimentBreakdown["NEUTRAL"] = r.NeutralCount
		stats.ConfidenceStats = models.ConfidenceStats{
			AverageConfidence: r.AvgConfidence,
			HighConfidence:    r.HighConfidence,
			MediumConfidence:  r.MediumConfidence,
			LowConfidence:     r.LowConfidence,
		}
	}

	teamComparison, err := ss.getTeamComparison(ctx)
	if err == nil {
		stats.TeamComparison = teamComparison
	}

	recentTrends, err := ss.getRecentTrends(ctx)
	if err == nil {
		stats.RecentTrends = recentTrends
	}

	return stats, nil
}

func (ss *SentimentService) GenerateReport(teamID primitive.ObjectID, startDate, endDate time.Time) (*models.SentimentReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	matchStage := bson.M{
		"team_id": teamID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{
			"$group": bson.M{
				"_id":             nil,
				"total_analyzed":  bson.M{"$sum": 1},
				"avg_sentiment":   bson.M{"$avg": "$score"},
				"positive_count":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "POSITIVE"}}, 1, 0}}},
				"negative_count":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEGATIVE"}}, 1, 0}}},
				"neutral_count":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEUTRAL"}}, 1, 0}}},
			},
		},
	}

	cursor, err := ss.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error generating report: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAnalyzed int64   `bson:"total_analyzed"`
		AvgSentiment  float64 `bson:"avg_sentiment"`
		PositiveCount int64   `bson:"positive_count"`
		NegativeCount int64   `bson:"negative_count"`
		NeutralCount  int64   `bson:"neutral_count"`
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("error decoding report: %w", err)
	}

	report := &models.SentimentReport{
		TeamID: teamID,
		Period: models.ReportPeriod{
			StartDate: startDate,
			EndDate:   endDate,
			Label:     fmt.Sprintf("%s - %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		},
		SentimentCounts:    make(map[string]int64),
		HourlyDistribution: make(map[int]models.SentimentHourly),
		SourceBreakdown:    make(map[string]models.SentimentSourceStats),
		GeneratedAt:        time.Now(),
	}

	if len(result) > 0 {
		r := result[0]
		report.TotalAnalyzed = r.TotalAnalyzed
		report.AverageSentiment = r.AvgSentiment
		report.SentimentCounts["POSITIVE"] = r.PositiveCount
		report.SentimentCounts["NEGATIVE"] = r.NegativeCount
		report.SentimentCounts["NEUTRAL"] = r.NeutralCount
	}

	return report, nil
}

func (ss *SentimentService) preprocessText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	
	// BERT modelinin 512 token limitini aşmamak için text'i kısalt
	// Türkçe için ortalama 1 token ≈ 4-5 karakter, güvenlik için 400 token ≈ 1600 karakter limiti
	maxLength := 1600
	if len(text) > maxLength {
		// Kelime ortasından kesmemek için son tam kelimenin sonunda kes
		truncated := text[:maxLength]
		lastSpace := strings.LastIndex(truncated, " ")
		if lastSpace > maxLength/2 { // Çok kısa olmasın diye kontrol
			text = truncated[:lastSpace] + "..."
		} else {
			text = truncated + "..."
		}
	}
	
	return text
}

func (ss *SentimentService) normalizeLabel(label string) string {
	label = strings.ToUpper(strings.TrimSpace(label))
	
	switch label {
	case "POSITIVE", "POZITIF", "POS":
		return "POSITIVE"
	case "NEGATIVE", "NEGATIF", "NEG":
		return "NEGATIVE"
	case "NEUTRAL", "NÖTR", "NEU":
		return "NEUTRAL"
	default:
		return "NEUTRAL"
	}
}

func (ss *SentimentService) findBestResult(results []HuggingFaceResponse) HuggingFaceResponse {
	if len(results) == 0 {
		return HuggingFaceResponse{Label: "NEUTRAL", Score: 0.5}
	}
	
	if len(results) == 1 {
		return results[0]
	}
	
	best := results[0]
	for _, result := range results[1:] {
		if result.Score > best.Score {
			best = result
		}
	}
	
	return best
}

func (ss *SentimentService) getTeamComparison(ctx context.Context) ([]models.TeamSentimentComparison, error) {
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "teams",
				"localField":   "team_id",
				"foreignField": "_id",
				"as":           "team",
			},
		},
		{
			"$unwind": "$team",
		},
		{
			"$group": bson.M{
				"_id":            "$team_id",
				"team_name":      bson.M{"$first": "$team.name"},
				"avg_sentiment":  bson.M{"$avg": "$score"},
				"total_comments": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"avg_sentiment": -1},
		},
	}

	cursor, err := ss.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comparisons []models.TeamSentimentComparison
	ranking := 1
	for cursor.Next(ctx) {
		var doc struct {
			ID            primitive.ObjectID `bson:"_id"`
			TeamName      string             `bson:"team_name"`
			AvgSentiment  float64            `bson:"avg_sentiment"`
			TotalComments int64              `bson:"total_comments"`
		}
		if err := cursor.Decode(&doc); err == nil {
			comparisons = append(comparisons, models.TeamSentimentComparison{
				TeamID:        doc.ID,
				TeamName:      doc.TeamName,
				AvgSentiment:  doc.AvgSentiment,
				TotalComments: doc.TotalComments,
				Ranking:       ranking,
			})
			ranking++
		}
	}

	return comparisons, nil
}

func (ss *SentimentService) getRecentTrends(ctx context.Context) ([]models.SentimentTrend, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": time.Now().AddDate(0, 0, -7),
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$created_at",
					},
				},
				"positive": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "POSITIVE"}}, 1, 0}}},
				"negative": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEGATIVE"}}, 1, 0}}},
				"neutral":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEUTRAL"}}, 1, 0}}},
				"avg_score": bson.M{"$avg": "$score"},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := ss.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trends []models.SentimentTrend
	for cursor.Next(ctx) {
		var doc struct {
			Date     string  `bson:"_id"`
			Positive int64   `bson:"positive"`
			Negative int64   `bson:"negative"`
			Neutral  int64   `bson:"neutral"`
			Score    float64 `bson:"avg_score"`
		}
		if err := cursor.Decode(&doc); err == nil {
			trends = append(trends, models.SentimentTrend{
				Date:     doc.Date,
				Positive: doc.Positive,
				Negative: doc.Negative,
				Neutral:  doc.Neutral,
				Score:    doc.Score,
			})
		}
	}

	return trends, nil
}

func (ss *SentimentService) CleanupDuplicates() (*models.CleanupResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Starting duplicate sentiment cleanup...")

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$comment_id",
				"ids": bson.M{"$push": "$_id"},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gt": 1},
			},
		},
	}

	cursor, err := ss.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error finding duplicates: %w", err)
	}
	defer cursor.Close(ctx)

	var totalDuplicates int64
	var removedCount int64
	var errors []string

	for cursor.Next(ctx) {
		var duplicate struct {
			CommentID primitive.ObjectID   `bson:"_id"`
			IDs       []primitive.ObjectID `bson:"ids"`
			Count     int64                `bson:"count"`
		}
		
		if err := cursor.Decode(&duplicate); err != nil {
			log.Printf("Error decoding duplicate: %v", err)
			continue
		}

		totalDuplicates += duplicate.Count - 1

		duplicateIDs := duplicate.IDs[1:]
		
		filter := bson.M{
			"_id": bson.M{"$in": duplicateIDs},
		}
		
		result, err := ss.collection.DeleteMany(ctx, filter)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to delete duplicates for comment %s: %v", duplicate.CommentID.Hex(), err)
			errors = append(errors, errMsg)
			log.Printf(errMsg)
			continue
		}
		
		removedCount += result.DeletedCount
		log.Printf("Removed %d duplicate sentiments for comment %s", result.DeletedCount, duplicate.CommentID.Hex())
	}

	log.Printf("Cleanup completed. Removed %d of %d duplicate sentiments", removedCount, totalDuplicates)

	return &models.CleanupResult{
		TotalDuplicatesFound: totalDuplicates,
		DuplicatesRemoved:    removedCount,
		Errors:              errors,
		CompletedAt:          time.Now(),
	}, nil
}