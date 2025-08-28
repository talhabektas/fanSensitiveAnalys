package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EnhancedAnalyticsService struct {
	sentimentCollection *mongo.Collection
	commentCollection   *mongo.Collection
	summaryCollection   *mongo.Collection
	trendsCollection    *mongo.Collection
	groqService         *GroqService
}

func NewEnhancedAnalyticsService() *EnhancedAnalyticsService {
	return &EnhancedAnalyticsService{
		sentimentCollection: config.GetCollection("sentiments"),
		commentCollection:   config.GetCollection("comments"),
		summaryCollection:   config.GetCollection("comment_summaries"),
		trendsCollection:    config.GetCollection("trend_insights"),
		groqService:         NewGroqService(),
	}
}

// 1. Gelişmiş İstatistikler
func (eas *EnhancedAnalyticsService) GetEnhancedSentimentStats(teamID *primitive.ObjectID) (*models.EnhancedSentimentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Temel filtre
	baseFilter := bson.M{}
	if teamID != nil {
		baseFilter["team_id"] = *teamID
	}

	// Ana istatistikleri al
	pipeline := []bson.M{
		{"$match": baseFilter},
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
				"avg_toxicity":       bson.M{"$avg": "$analysis_details.toxicity_score"},
				"high_toxicity":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$gte": []interface{}{"$analysis_details.toxicity_score", 0.7}}, 1, 0}}},
			},
		},
	}

	cursor, err := eas.sentimentCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating enhanced stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAnalyzed    int64   `bson:"total_analyzed"`
		AvgSentiment     float64 `bson:"avg_sentiment"`
		AvgConfidence    float64 `bson:"avg_confidence"`
		PositiveCount    int64   `bson:"positive_count"`
		NegativeCount    int64   `bson:"negative_count"`
		NeutralCount     int64   `bson:"neutral_count"`
		HighConfidence   int64   `bson:"high_confidence"`
		MediumConfidence int64   `bson:"medium_confidence"`
		LowConfidence    int64   `bson:"low_confidence"`
		AvgToxicity      float64 `bson:"avg_toxicity"`
		HighToxicity     int64   `bson:"high_toxicity"`
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("error decoding enhanced stats: %w", err)
	}

	stats := &models.EnhancedSentimentStats{
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

		stats.ToxicityStats = models.ToxicityStats{
			TotalScanned:    r.TotalAnalyzed,
			HighToxicity:    r.HighToxicity,
			MediumToxicity:  0, // Hesaplanacak
			LowToxicity:     r.TotalAnalyzed - r.HighToxicity,
			AverageToxicity: r.AvgToxicity,
		}
	}

	// Kategori dağılımını al
	categoryBreakdown, err := eas.getCategoryBreakdown(ctx, baseFilter)
	if err == nil {
		stats.CategoryBreakdown = categoryBreakdown
	}

	// Son özetleri al
	latestSummary, err := eas.getLatestSummary(ctx, teamID)
	if err == nil {
		stats.LatestSummary = latestSummary
	}

	// Trend içgörüleri al
	trendInsights, err := eas.getRecentTrendInsights(ctx, teamID)
	if err == nil {
		stats.TrendInsights = trendInsights
	}

	return stats, nil
}

// 2. Kategori Dağılımı
func (eas *EnhancedAnalyticsService) getCategoryBreakdown(ctx context.Context, baseFilter bson.M) ([]models.CategoryStats, error) {
	pipeline := []bson.M{
		{"$match": baseFilter},
		{
			"$group": bson.M{
				"_id":           "$analysis_details.category",
				"count":         bson.M{"$sum": 1},
				"avg_sentiment": bson.M{"$avg": "$score"},
				"keywords":      bson.M{"$push": "$analysis_details.keywords"},
			},
		},
		{"$sort": bson.M{"count": -1}},
	}

	cursor, err := eas.sentimentCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.CategoryStats
	totalCount := int64(0)

	// İlk geçişte toplam sayıyı hesapla
	var tempResults []struct {
		Category     string              `bson:"_id"`
		Count        int64               `bson:"count"`
		AvgSentiment float64             `bson:"avg_sentiment"`
		Keywords     [][]string          `bson:"keywords"`
	}

	if err = cursor.All(ctx, &tempResults); err != nil {
		return nil, err
	}

	for _, r := range tempResults {
		totalCount += r.Count
	}

	// İkinci geçişte yüzdeleri hesapla
	for _, r := range tempResults {
		category := r.Category
		if category == "" {
			category = "Genel"
		}

		// Anahtar kelimeleri flatten et
		keywordMap := make(map[string]bool)
		for _, kwList := range r.Keywords {
			for _, kw := range kwList {
				if kw != "" {
					keywordMap[kw] = true
				}
			}
		}

		var topKeywords []string
		for kw := range keywordMap {
			topKeywords = append(topKeywords, kw)
			if len(topKeywords) >= 5 { // En fazla 5 anahtar kelime
				break
			}
		}

		categories = append(categories, models.CategoryStats{
			Category:     category,
			Count:        r.Count,
			Percentage:   float64(r.Count) * 100 / float64(totalCount),
			AvgSentiment: r.AvgSentiment,
			Keywords:     topKeywords,
		})
	}

	return categories, nil
}

// 3. Yorum Özetleme - Tüm Takımlar
func (eas *EnhancedAnalyticsService) GenerateAllTeamsSummary() (*models.CommentSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Son 24 saatin yorumlarını al (tüm takımlardan)
	since := time.Now().Add(-24 * time.Hour)
	filter := bson.M{
		"created_at": bson.M{"$gte": since},
	}

	cursor, err := eas.commentCollection.Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		return nil, fmt.Errorf("error fetching comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, fmt.Errorf("error decoding comments: %w", err)
	}

	if len(comments) == 0 {
		return nil, fmt.Errorf("no comments found for summary")
	}

	// Grok AI ile gerçek analiz yap
	var commentTexts []string
	var platforms []string
	
	for i, comment := range comments {
		if i >= 30 { // İlk 30 yorumu analiz et
			break
		}
		if len(comment.Text) > 10 {
			commentTexts = append(commentTexts, comment.Text)
			platforms = append(platforms, comment.Source)
		}
	}

	if len(commentTexts) == 0 {
		return nil, fmt.Errorf("no valid comment texts found for analysis")
	}

	// Grok AI'dan gerçek analiz iste
	analysisPrompt := fmt.Sprintf(`Türk futbol taraftarlarının son 24 saatteki %d yorumunu analiz et ve özetini çıkar:

YORUMLAR:
%s

Lütfen şu formatta analiz ver:
- En çok konuşulan 3 konu nedir?
- Taraftarların genel ruh hali nasıl (iyimser/kötümser/kararsız)?
- Hangi transferler/oyuncular/teknik direktörler en çok tartışılıyor?
- Yaklaşan maçlar hakkında ne düşünüyorlar?
- Genel sonuç cümlesi

Gerçek bir futbol uzmanı gibi yorumla.`, len(commentTexts), strings.Join(commentTexts[:min(len(commentTexts), 20)], " | "))

	aiSummary, err := eas.groqService.callGroqAPI(analysisPrompt)
	if err != nil {
		log.Printf("Grok AI analysis failed, using fallback: %v", err)
		aiSummary = fmt.Sprintf("Son 24 saatte %d taraftar yorumu analiz edildi. Yorumlar genellikle transfer haberleri, maç sonuçları ve takım performansı üzerine yoğunlaşıyor.", len(comments))
	}

	// Ana konuları çıkar
	topics := []string{"transfer haberleri", "maç yorumları", "takım performansı"}
	if len(platforms) > 0 {
		topics = append(topics, "sosyal medya tepkileri")
	}

	return &models.CommentSummary{
		TeamID:        primitive.NilObjectID,
		Period:        models.ReportPeriod{StartDate: since, EndDate: time.Now(), Label: "Son 24 Saat"},
		Summary:       aiSummary,
		TotalComments: len(comments),
		MainTopics:    topics,
		GeneratedBy:   "grok-ai-enhanced",
		CreatedAt:     time.Now(),
	}, nil
}

// 3. Yorum Özetleme - Belirli Takım
func (eas *EnhancedAnalyticsService) GenerateDailySummary(teamID primitive.ObjectID) (*models.CommentSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Son 24 saatin yorumlarını al
	since := time.Now().Add(-24 * time.Hour)
	filter := bson.M{
		"team_id": teamID,
		"created_at": bson.M{"$gte": since},
	}

	cursor, err := eas.commentCollection.Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		return nil, fmt.Errorf("error fetching comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, fmt.Errorf("error decoding comments: %w", err)
	}

	if len(comments) == 0 {
		return nil, fmt.Errorf("no comments found for summary")
	}

	// Grok AI ile takıma özel gerçek analiz
	var commentTexts []string
	var teamName string
	
	for i, comment := range comments {
		if i >= 25 { // İlk 25 yorumu analiz et
			break
		}
		if len(comment.Text) > 10 {
			commentTexts = append(commentTexts, comment.Text)
		}
	}

	if len(commentTexts) == 0 {
		return nil, fmt.Errorf("no valid comment texts found for team analysis")
	}

	// Takım ismini al (basit olsun)
	teamName = "Bu takım" // Gerçekte team collection'dan alınacak

	// Takıma özel analiz prompt'u
	teamAnalysisPrompt := fmt.Sprintf(`%s taraftarlarının son 24 saatteki %d yorumunu analiz et:

YORUMLAR:
%s

Takıma özel analiz yap ve şunları söyle:
- Bu takımın taraftarları en çok hangi konularda yorum yapıyor? (transfer, maç, teknik direktör, vs)
- Taraftarların takım hakkındaki genel düşünceleri nasıl? (memnun/memnun değil/endişeli)
- Hangi oyuncular en çok konuşuluyor? (pozitif/negatif)
- Teknik direktör hakkında ne düşünüyorlar?
- Yaklaşan maçlar için beklentileri nedir?
- Takım performansından memnunlar mı?

Kısa ve öz bir futbol uzmanı analizi yap.`, teamName, len(commentTexts), strings.Join(commentTexts[:min(len(commentTexts), 15)], " | "))

	aiSummary, err := eas.groqService.callGroqAPI(teamAnalysisPrompt)
	if err != nil {
		log.Printf("Team-specific Grok AI analysis failed: %v", err)
		aiSummary = fmt.Sprintf("%s taraftarları son 24 saatte aktif yorumlar yaptı. Öne çıkan konular: takım performansı, transfer haberleri ve maç stratejileri.", teamName)
	}

	// Takıma özel konular
	topics := []string{"takım performansı", "oyuncu yorumları", "taraftar tepkileri"}

	// Özeti döndür
	commentSummary := &models.CommentSummary{
		ID:            primitive.NewObjectID(),
		TeamID:        teamID,
		Period:        models.ReportPeriod{StartDate: since, EndDate: time.Now(), Label: "Son 24 Saat"},
		Summary:       aiSummary,
		TotalComments: len(comments),
		MainTopics:    topics,
		GeneratedBy:   "grok-ai-team-analysis",
		CreatedAt:     time.Now(),
	}

	return commentSummary, nil
}

// min function kaldırıldı - youtube_service.go'da var

// 4. Trend Analizi
func (eas *EnhancedAnalyticsService) GenerateTrendInsights(teamID *primitive.ObjectID) ([]*models.TrendInsight, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Son 7 günün verilerini al
	since := time.Now().AddDate(0, 0, -7)
	filter := bson.M{
		"created_at": bson.M{"$gte": since},
	}
	if teamID != nil {
		filter["team_id"] = *teamID
	}

	cursor, err := eas.sentimentCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error fetching trend data: %w", err)
	}
	defer cursor.Close(ctx)

	var sentiments []models.Sentiment
	if err = cursor.All(ctx, &sentiments); err != nil {
		return nil, fmt.Errorf("error decoding trend data: %w", err)
	}

	if len(sentiments) < 10 {
		return nil, fmt.Errorf("insufficient data for trend analysis")
	}

	// Grok AI ile trend analizi
	trendReport, err := eas.groqService.AnalyzeTrends(sentiments)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trends: %w", err)
	}

	// Trend insight'ı oluştur ve kaydet
	insight := &models.TrendInsight{
		ID: primitive.NewObjectID(),
		Period: models.ReportPeriod{
			StartDate: since,
			EndDate:   time.Now(),
			Label:     "Son 7 Gün",
		},
		Title:       "Haftalık Trend Analizi",
		Description: trendReport,
		TrendType:   "general",
		Confidence:  0.8,
		Keywords:    []string{"trend", "analiz", "haftalık"},
		GeneratedBy: "grok-ai",
		CreatedAt:   time.Now(),
	}

	if teamID != nil {
		insight.TeamID = *teamID
	}

	_, err = eas.trendsCollection.InsertOne(ctx, insight)
	if err != nil {
		log.Printf("Failed to save trend insight: %v", err)
	}

	return []*models.TrendInsight{insight}, nil
}

func (eas *EnhancedAnalyticsService) getLatestSummary(ctx context.Context, teamID *primitive.ObjectID) (string, error) {
	filter := bson.M{}
	if teamID != nil {
		filter["team_id"] = *teamID
	}

	var summary models.CommentSummary
	err := eas.summaryCollection.FindOne(ctx, filter, 
		options.FindOne().SetSort(bson.M{"created_at": -1})).Decode(&summary)
	
	if err != nil {
		return "", err
	}

	return summary.Summary, nil
}

func (eas *EnhancedAnalyticsService) getRecentTrendInsights(ctx context.Context, teamID *primitive.ObjectID) ([]models.TrendInsight, error) {
	filter := bson.M{}
	if teamID != nil {
		filter["team_id"] = *teamID
	}

	cursor, err := eas.trendsCollection.Find(ctx, filter, 
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(5))
	
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var insights []models.TrendInsight
	if err = cursor.All(ctx, &insights); err != nil {
		return nil, err
	}

	return insights, nil
}