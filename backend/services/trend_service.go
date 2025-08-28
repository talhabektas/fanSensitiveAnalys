package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"taraftar-analizi/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TrendService struct {
	commentCollection   *mongo.Collection
	sentimentCollection *mongo.Collection
	teamService         *TeamService
}

type TrendData struct {
	Date     string  `json:"date"`
	Positive int     `json:"positive"`
	Negative int     `json:"negative"`
	Neutral  int     `json:"neutral"`
	Total    int     `json:"total"`
	Score    float64 `json:"score"` // Ortalama sentiment puanı (-1 ile 1 arası)
}

type TeamTrend struct {
	TeamID   string      `json:"team_id"`
	TeamName string      `json:"team_name"`
	Data     []TrendData `json:"data"`
	Overall  struct {
		TotalComments    int     `json:"total_comments"`
		PositivePercent  float64 `json:"positive_percent"`
		NegativePercent  float64 `json:"negative_percent"`
		NeutralPercent   float64 `json:"neutral_percent"`
		TrendDirection   string  `json:"trend_direction"` // "up", "down", "stable"
		WeeklyChange     float64 `json:"weekly_change"`   // Yüzde değişim
	} `json:"overall"`
}

type TrendAnalysisResponse struct {
	Period    string      `json:"period"`    // "7d", "30d", "90d"
	StartDate string      `json:"start_date"`
	EndDate   string      `json:"end_date"`
	Teams     []TeamTrend `json:"teams"`
	Summary   struct {
		TotalComments       int     `json:"total_comments"`
		MostPositiveTeam    string  `json:"most_positive_team"`
		MostNegativeTeam    string  `json:"most_negative_team"`
		BiggestImprovement  string  `json:"biggest_improvement"`
		BiggestDecline      string  `json:"biggest_decline"`
		AverageDaily        float64 `json:"average_daily"`
	} `json:"summary"`
}

type InsightData struct {
	Type        string `json:"type"`        // "improvement", "decline", "spike", "drop"
	TeamName    string `json:"team_name"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Severity    string `json:"severity"` // "high", "medium", "low"
}

func NewTrendService() *TrendService {
	return &TrendService{
		commentCollection:   config.GetCollection("comments"),
		sentimentCollection: config.GetCollection("sentiments"),
		teamService:         NewTeamService(),
	}
}

// GetTrendAnalysis - Belirtilen dönem için trend analizi
func (ts *TrendService) GetTrendAnalysis(period string) (*TrendAnalysisResponse, error) {
	days := 7
	switch period {
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		period = "7d"
		days = 7
	}

	// Force current time calculation - no caching
	// Include today's data by extending endDate to end of day
	endDate := time.Now().AddDate(0, 0, 1) // Add 1 day to include today
	startDate := endDate.AddDate(0, 0, -days-1) // Adjust start accordingly
	
	// Debug logging
	log.Printf("[TrendAnalysis] Period: %s, Days: %d, StartDate: %s, EndDate: %s", 
		period, days, startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))

	// Tüm takımları al
	teams, err := ts.teamService.GetAllTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %v", err)
	}

	response := &TrendAnalysisResponse{
		Period:    period,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Teams:     make([]TeamTrend, 0),
	}
	
	// Debug: Log teams count
	log.Printf("[TrendAnalysis] Found %d teams to analyze", len(teams))

	var totalComments int
	var mostPositiveTeam, mostNegativeTeam string
	var highestPositive, lowestPositive float64 = -1, 101

	// Her takım için trend analizi yap
	for _, team := range teams {
		teamTrend, err := ts.getTeamTrend(team.ID, startDate, endDate, days)
		if err != nil {
			log.Printf("Failed to get trend for team %s: %v", team.Name, err)
			continue
		}

		teamTrend.TeamName = team.Name
		response.Teams = append(response.Teams, *teamTrend)

		// Summary için istatistikler topla
		totalComments += teamTrend.Overall.TotalComments
		log.Printf("[TrendAnalysis] Team %s - Comments: %d, Positive: %.2f%%", 
			team.Name, teamTrend.Overall.TotalComments, teamTrend.Overall.PositivePercent)
		
		if teamTrend.Overall.PositivePercent > highestPositive {
			highestPositive = teamTrend.Overall.PositivePercent
			mostPositiveTeam = team.Name
		}
		if teamTrend.Overall.PositivePercent < lowestPositive && teamTrend.Overall.TotalComments > 10 {
			lowestPositive = teamTrend.Overall.PositivePercent
			mostNegativeTeam = team.Name
		}
	}

	// Summary bilgilerini doldur
	response.Summary.TotalComments = totalComments
	response.Summary.MostPositiveTeam = mostPositiveTeam
	response.Summary.MostNegativeTeam = mostNegativeTeam
	response.Summary.AverageDaily = float64(totalComments) / float64(days)
	
	// Debug: Log final summary
	log.Printf("[TrendAnalysis] Final Summary - Total: %d, Daily Avg: %.2f, Most Positive: %s", 
		totalComments, response.Summary.AverageDaily, mostPositiveTeam)

	// En çok gelişen ve düşen takımları bul
	ts.calculateImprovements(response)

	return response, nil
}

// getTeamTrend - Tek takım için trend verileri
func (ts *TrendService) getTeamTrend(teamID primitive.ObjectID, startDate, endDate time.Time, days int) (*TeamTrend, error) {
	trend := &TeamTrend{
		TeamID: teamID.Hex(),
		Data:   make([]TrendData, 0),
	}
	
	log.Printf("[getTeamTrend] Processing team %s for %d days (%s to %s)", 
		teamID.Hex(), days, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Her gün için veri topla (bugün dahil)
	for i := 0; i <= days; i++ {
		date := startDate.AddDate(0, 0, i)
		nextDate := date.AddDate(0, 0, 1)

		// O günkü comment ve sentiment verilerini al
		log.Printf("[getTeamTrend] Querying day %s for team %s", date.Format("2006-01-02"), teamID.Hex())
		
		var dayData TrendData
		dayData.Date = date.Format("2006-01-02")
		
		// 1. Comment sayısını al (sentiment olsun olmasın tüm comment'lar)
		commentCount, err := ts.commentCollection.CountDocuments(context.TODO(), bson.M{
			"team_id": teamID,
			"created_at": bson.M{
				"$gte": date,
				"$lt":  nextDate,
			},
		})
		if err == nil {
			// Total'a comment sayısını ata (sentiment olmasa da göstermek için)
			dayData.Total = int(commentCount)
		}
		
		// 2. O günkü sentiment verilerini al - doğrudan sentiment collection'dan
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"team_id": teamID,
					"created_at": bson.M{
						"$gte": date,
						"$lt":  nextDate,
					},
				},
			},
			{
				"$group": bson.M{
					"_id": "$sentiment",
					"count": bson.M{"$sum": 1},
					"avg_confidence": bson.M{"$avg": "$confidence"},
				},
			},
		}

		cursor, err := ts.sentimentCollection.Aggregate(context.TODO(), pipeline)
		if err != nil {
			// Sentiment hatası olsa da comment sayısını göster
			trend.Data = append(trend.Data, dayData)
			continue
		}

		var results []struct {
			ID            string  `bson:"_id"`
			Count         int     `bson:"count"`
			AvgConfidence float64 `bson:"avg_confidence"`
		}

		if err := cursor.All(context.TODO(), &results); err == nil {
			log.Printf("[getTeamTrend] Found %d sentiment records for %s on %s (total comments: %d)", len(results), teamID.Hex(), date.Format("2006-01-02"), dayData.Total)
			totalScore := 0.0
			sentimentTotal := 0
			for _, result := range results {
				switch result.ID {
				case "POSITIVE":
					dayData.Positive = result.Count
					totalScore += float64(result.Count) * 1.0
				case "NEGATIVE":
					dayData.Negative = result.Count
					totalScore += float64(result.Count) * -1.0
				case "NEUTRAL":
					dayData.Neutral = result.Count
				}
				sentimentTotal += result.Count
			}

			if sentimentTotal > 0 {
				dayData.Score = totalScore / float64(sentimentTotal)
			}
		}

		trend.Data = append(trend.Data, dayData)
		cursor.Close(context.TODO())
	}

	// Overall istatistikleri hesapla
	ts.calculateOverallStats(trend)

	return trend, nil
}

// calculateOverallStats - Genel istatistikleri hesapla
func (ts *TrendService) calculateOverallStats(trend *TeamTrend) {
	var totalPositive, totalNegative, totalNeutral, totalComments int
	var weeklyScores []float64

	for i, data := range trend.Data {
		totalPositive += data.Positive
		totalNegative += data.Negative
		totalNeutral += data.Neutral
		totalComments += data.Total

		// Son 7 günün skorlarını topla trend için
		if i >= len(trend.Data)-7 {
			weeklyScores = append(weeklyScores, data.Score)
		}
	}

	trend.Overall.TotalComments = totalComments

	if totalComments > 0 {
		trend.Overall.PositivePercent = float64(totalPositive) / float64(totalComments) * 100
		trend.Overall.NegativePercent = float64(totalNegative) / float64(totalComments) * 100
		trend.Overall.NeutralPercent = float64(totalNeutral) / float64(totalComments) * 100
	}

	// Trend yönünü hesapla
	if len(weeklyScores) >= 4 {
		firstHalf := (weeklyScores[0] + weeklyScores[1]) / 2
		secondHalf := (weeklyScores[len(weeklyScores)-2] + weeklyScores[len(weeklyScores)-1]) / 2
		
		change := secondHalf - firstHalf
		trend.Overall.WeeklyChange = change * 100 // Yüzde olarak

		if change > 0.1 {
			trend.Overall.TrendDirection = "up"
		} else if change < -0.1 {
			trend.Overall.TrendDirection = "down"
		} else {
			trend.Overall.TrendDirection = "stable"
		}
	}
}

// calculateImprovements - En çok gelişen/düşen takımları bul
func (ts *TrendService) calculateImprovements(response *TrendAnalysisResponse) {
	var maxImprovement, maxDecline float64 = -100, 100
	var improvedTeam, declinedTeam string

	for _, team := range response.Teams {
		change := team.Overall.WeeklyChange
		
		if change > maxImprovement {
			maxImprovement = change
			improvedTeam = team.TeamName
		}
		
		if change < maxDecline {
			maxDecline = change
			declinedTeam = team.TeamName
		}
	}

	response.Summary.BiggestImprovement = improvedTeam
	response.Summary.BiggestDecline = declinedTeam
}

// GenerateInsights - AI benzeri insight'lar üret
func (ts *TrendService) GenerateInsights(period string) ([]InsightData, error) {
	analysis, err := ts.GetTrendAnalysis(period)
	if err != nil {
		return nil, err
	}

	var insights []InsightData

	// Her takım için insight'lar üret
	for _, team := range analysis.Teams {
		// Gelişme varsa (threshold azaltıldı)
		if team.Overall.WeeklyChange > 5 {
			insights = append(insights, InsightData{
				Type:        "improvement",
				TeamName:    team.TeamName,
				Description: fmt.Sprintf("%s taraftar duyguları %s döneminde %.1f%% iyileşme gösterdi", team.TeamName, period, team.Overall.WeeklyChange),
				Value:       fmt.Sprintf("+%.1f%%", team.Overall.WeeklyChange),
				Severity:    "high",
			})
		}

		// Düşüş varsa (threshold azaltıldı)
		if team.Overall.WeeklyChange < -5 {
			insights = append(insights, InsightData{
				Type:        "decline",
				TeamName:    team.TeamName,
				Description: fmt.Sprintf("%s taraftar duyguları %s döneminde %.1f%% düştü", team.TeamName, period, -team.Overall.WeeklyChange),
				Value:       fmt.Sprintf("%.1f%%", team.Overall.WeeklyChange),
				Severity:    "high",
			})
		}

		// Pozitif takım (threshold azaltıldı)
		if team.Overall.PositivePercent > 60 && team.Overall.TotalComments > 10 {
			insights = append(insights, InsightData{
				Type:        "spike",
				TeamName:    team.TeamName,
				Description: fmt.Sprintf("%s taraftarlarının %.1f%%'i pozitif duygular taşıyor", team.TeamName, team.Overall.PositivePercent),
				Value:       fmt.Sprintf("%.1f%%", team.Overall.PositivePercent),
				Severity:    "medium",
			})
		}

		// Negatif yoğunluk
		if team.Overall.NegativePercent > 40 && team.Overall.TotalComments > 10 {
			insights = append(insights, InsightData{
				Type:        "warning",
				TeamName:    team.TeamName,
				Description: fmt.Sprintf("%s taraftarlarının %.1f%%'i negatif duygular ifade ediyor", team.TeamName, team.Overall.NegativePercent),
				Value:       fmt.Sprintf("%.1f%%", team.Overall.NegativePercent),
				Severity:    "medium",
			})
		}

		// Yüksek aktivite
		if team.Overall.TotalComments > 100 {
			insights = append(insights, InsightData{
				Type:        "activity",
				TeamName:    team.TeamName,
				Description: fmt.Sprintf("%s için %s döneminde %d yorum analiz edildi - yüksek taraftar aktivitesi", team.TeamName, period, team.Overall.TotalComments),
				Value:       fmt.Sprintf("%d yorum", team.Overall.TotalComments),
				Severity:    "low",
			})
		}
	}

	// Genel insight'lar (threshold azaltıldı)
	if analysis.Summary.AverageDaily > 10 {
		insights = append(insights, InsightData{
			Type:        "spike",
			TeamName:    "Genel",
			Description: fmt.Sprintf("Günlük ortalama %.0f yorum - düzenli taraftar aktivitesi gözlemleniyor", analysis.Summary.AverageDaily),
			Value:       fmt.Sprintf("%.0f/gün", analysis.Summary.AverageDaily),
			Severity:    "medium",
		})
	}

	// Platform insight'ı
	if analysis.Summary.TotalComments > 0 {
		insights = append(insights, InsightData{
			Type:        "trend",
			TeamName:    "Platform",
			Description: fmt.Sprintf("Toplam %d yorum analiz edildi. Günlük ortalama %.1f yorum işleniyor", analysis.Summary.TotalComments, analysis.Summary.AverageDaily),
			Value:       fmt.Sprintf("%d yorum", analysis.Summary.TotalComments),
			Severity:    "low",
		})
	}

	// Insight'ları önem sırasına göre sırala
	sort.Slice(insights, func(i, j int) bool {
		severityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return severityOrder[insights[i].Severity] > severityOrder[insights[j].Severity]
	})

	return insights, nil
}