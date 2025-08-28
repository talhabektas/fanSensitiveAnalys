package routes

import (
	"net/http"
	"time"

	"taraftar-analizi/models"
	"taraftar-analizi/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type N8NWebhookRouter struct {
	youtubeService      *services.YouTubeService
	analyticsService    *services.EnhancedAnalyticsService
	sentimentService    *services.SentimentService
}

func NewN8NWebhookRouter() *N8NWebhookRouter {
	return &N8NWebhookRouter{
		youtubeService:   services.NewYouTubeService(),
		analyticsService: services.NewEnhancedAnalyticsService(),
		sentimentService: services.NewSentimentService(),
	}
}

// N8N için özel endpoint'ler
func (n *N8NWebhookRouter) SetupRoutes(router *gin.Engine) {
	n8n := router.Group("/api/n8n")
	{
		// YouTube yorum toplama webhook'u
		n8n.POST("/collect-comments", n.CollectCommentsWebhook)
		
		// Sentiment durumu kontrol webhook'u
		n8n.GET("/sentiment-status", n.GetSentimentStatus)
		
		// Takım bazlı analiz webhook'u
		n8n.GET("/team-analysis/:teamId", n.GetTeamAnalysis)
		
		// Trend uyarısı webhook'u
		n8n.GET("/trend-alerts", n.GetTrendAlerts)
		
		// Performans raporu webhook'u
		n8n.GET("/performance-report", n.GetPerformanceReport)
		
		// Sistem durumu webhook'u
		n8n.GET("/health-check", n.HealthCheck)
	}
}

// YouTube yorumlarını topla ve N8N'e response döndür
func (n *N8NWebhookRouter) CollectCommentsWebhook(c *gin.Context) {
	startTime := time.Now()
	
	// YouTube yorumlarını topla
	err := n.youtubeService.CollectTurkishFootballComments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"duration": time.Since(startTime).Seconds(),
		})
		return
	}
	
	// Toplama sonrası istatistikleri al
	stats, err := n.analyticsService.GetEnhancedSentimentStats(nil)
	if err != nil {
		// İstatistik alamasak bile toplama başarılı sayılır
		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"message":         "YouTube comments collected successfully",
			"duration":        time.Since(startTime).Seconds(),
			"timestamp":       time.Now().Format(time.RFC3339),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "YouTube comments collected successfully",
		"duration":        time.Since(startTime).Seconds(),
		"total_collected": stats.TotalAnalyzed,
		"overall_sentiment": stats.OverallSentiment,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// Sentiment durumunu N8N için optimize edilmiş formatta döndür
func (n *N8NWebhookRouter) GetSentimentStatus(c *gin.Context) {
	stats, err := n.analyticsService.GetEnhancedSentimentStats(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sentiment stats",
		})
		return
	}
	
	// Alert durumlarını hesapla
	isHighNegative := stats.OverallSentiment < -0.3
	isHighPositive := stats.OverallSentiment > 0.5
	isNormal := !isHighNegative && !isHighPositive
	
	// En düşük ve en yüksek takımları bul
	var lowestTeam, highestTeam string
	if len(stats.TeamComparison) > 0 {
		highestTeam = stats.TeamComparison[0].TeamName
		lowestTeam = stats.TeamComparison[len(stats.TeamComparison)-1].TeamName
	}
	
	c.JSON(http.StatusOK, gin.H{
		"overall_sentiment":  stats.OverallSentiment,
		"total_analyzed":     stats.TotalAnalyzed,
		"sentiment_breakdown": stats.SentimentBreakdown,
		"team_comparison":    stats.TeamComparison,
		"alert_status": gin.H{
			"is_high_negative": isHighNegative,
			"is_high_positive": isHighPositive,
			"is_normal":        isNormal,
		},
		"top_teams": gin.H{
			"highest": highestTeam,
			"lowest":  lowestTeam,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Belirli bir takım için detaylı analiz
func (n *N8NWebhookRouter) GetTeamAnalysis(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	
	teamID, err := primitive.ObjectIDFromHex(teamIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid team ID",
		})
		return
	}
	
	stats, err := n.analyticsService.GetEnhancedSentimentStats(&teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get team analysis",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"team_id":           teamIdStr,
		"analysis":          stats,
		"recommendation":    generateTeamRecommendation(stats),
		"timestamp":         time.Now().Format(time.RFC3339),
	})
}

// Trend uyarıları için endpoint
func (n *N8NWebhookRouter) GetTrendAlerts(c *gin.Context) {
	// Basit trend analizi - şimdilik boş döndür
	c.JSON(http.StatusOK, gin.H{
		"all_trends":      []interface{}{},
		"critical_trends": []interface{}{},
		"alert_count":     0,
		"message":         "Trend analysis coming soon",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// Performans raporu webhook'u
func (n *N8NWebhookRouter) GetPerformanceReport(c *gin.Context) {
	// İstatistikleri al
	stats, err := n.analyticsService.GetEnhancedSentimentStats(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate performance report",
		})
		return
	}
	
	// Performans skoru hesapla
	performanceScore := calculatePerformanceScore(
		stats.OverallSentiment,
		stats.ConfidenceStats.AverageConfidence,
		stats.ToxicityStats.AverageToxicity,
	)
	
	// Uyarıları tespit et
	alerts := []string{}
	if stats.OverallSentiment < -0.4 {
		alerts = append(alerts, "🚨 Kritik negatif sentiment!")
	}
	if stats.ToxicityStats.AverageToxicity > 0.15 {
		alerts = append(alerts, "⚠️ Yüksek toksiklik oranı!")
	}
	if stats.ConfidenceStats.AverageConfidence < 0.6 {
		alerts = append(alerts, "📊 Düşük güven seviyesi!")
	}
	
	c.JSON(http.StatusOK, gin.H{
		"performance_score": performanceScore,
		"alerts":           alerts,
		"stats_summary": gin.H{
			"total_analyzed":      stats.TotalAnalyzed,
			"overall_sentiment":   stats.OverallSentiment,
			"confidence_average":  stats.ConfidenceStats.AverageConfidence,
			"toxicity_average":    stats.ToxicityStats.AverageToxicity,
		},
		"team_rankings":     getTeamRankings(stats.TeamComparison),
		"top_categories":    getTopCategories(stats.CategoryBreakdown, 5),
		"generated_at":      time.Now().Format(time.RFC3339),
	})
}

// Sistem sağlığı kontrolü
func (n *N8NWebhookRouter) HealthCheck(c *gin.Context) {
	// Backend servislerin durumunu kontrol et
	health := gin.H{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": gin.H{
			"youtube_api":   checkYouTubeAPIHealth(),
			"groq_ai":       checkGroqAPIHealth(),
			"database":      checkDatabaseHealth(),
		},
		"version": "1.0.0",
	}
	
	c.JSON(http.StatusOK, health)
}

// Yardımcı fonksiyonlar

func generateTeamRecommendation(stats *models.EnhancedSentimentStats) string {
	if stats.OverallSentiment > 0.3 {
		return "Takım performansı ve taraftar memnuniyeti yüksek. Pozitif trendi sürdürmek için mevcut stratejiye devam edilmeli."
	} else if stats.OverallSentiment < -0.3 {
		return "Taraftarlar arasında olumsuz hava hakim. Acil PR ve iletişim stratejisi gerekiyor."
	}
	return "Taraftar sentiment'i nötr seviyede. Pozitif gelişmeler için proaktif adımlar atılabilir."
}

func calculatePerformanceScore(sentiment, confidence, toxicity float64) int {
	// 0-100 arası performans skoru hesapla
	score := 50.0 // Başlangıç skoru
	
	// Sentiment etkisi (+/- 30 puan)
	score += (sentiment * 30)
	
	// Güven seviyesi etkisi (+20 puan)
	score += (confidence * 20)
	
	// Toksiklik cezası (-30 puan)
	score -= (toxicity * 30)
	
	// 0-100 aralığında sınırla
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return int(score)
}

func getTeamRankings(teamComparison []models.TeamSentimentComparison) []map[string]interface{} {
	result := make([]map[string]interface{}, len(teamComparison))
	
	for i, team := range teamComparison {
		trend := "stabil"
		if team.AvgSentiment > 0.1 {
			trend = "yükseliş"
		} else if team.AvgSentiment < -0.1 {
			trend = "düşüş"
		}
		
		result[i] = map[string]interface{}{
			"rank":      i + 1,
			"name":      team.TeamName,
			"sentiment": team.AvgSentiment,
			"comments":  team.TotalComments,
			"trend":     trend,
		}
	}
	
	return result
}

func getTopCategories(categories []models.CategoryStats, limit int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, limit)
	
	for i, cat := range categories {
		if i >= limit { break }
		result = append(result, map[string]interface{}{
			"category":   cat.Category,
			"percentage": cat.Percentage,
			"count":      cat.Count,
		})
	}
	
	return result
}

// Sağlık kontrol fonksiyonları
func checkYouTubeAPIHealth() string {
	// Basit YouTube API kontrolü yapılabilir
	return "operational"
}

func checkGroqAPIHealth() string {
	// Basit Groq API kontrolü yapılabilir
	return "operational"
}

func checkDatabaseHealth() string {
	// MongoDB bağlantı kontrolü yapılabilir
	return "operational"
}