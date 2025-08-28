package routes

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"taraftar-analizi/models"
	"taraftar-analizi/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SentimentRoutes struct {
	sentimentService         *services.SentimentService
	commentService           *services.CommentService
	enhancedAnalyticsService *services.EnhancedAnalyticsService
}

func NewSentimentRoutes() *SentimentRoutes {
	return &SentimentRoutes{
		sentimentService:         services.NewSentimentService(),
		commentService:           services.NewCommentService(),
		enhancedAnalyticsService: services.NewEnhancedAnalyticsService(),
	}
}

func (sr *SentimentRoutes) RegisterRoutes(router *gin.RouterGroup) {
	sentiments := router.Group("/sentiments")
	{
		sentiments.POST("/analyze", sr.AnalyzeText)
		sentiments.POST("/analyze/batch", sr.AnalyzeBatch)
		sentiments.GET("/stats", sr.GetSentimentStats)
		sentiments.GET("/report/:teamId", sr.GetTeamReport)
		sentiments.POST("", sr.SaveSentiment)
		sentiments.DELETE("/cleanup", sr.CleanupDuplicates)
		
		// 🚀 Yeni Grok AI Özellikleri
		sentiments.GET("/enhanced-stats", sr.GetEnhancedSentimentStats)
		sentiments.GET("/enhanced-stats/:teamId", sr.GetTeamEnhancedStats)
		sentiments.POST("/summary/generate", sr.GenerateDailySummary) // Tüm takımlar için
		sentiments.POST("/summary/generate/:teamId", sr.GenerateDailySummary)
		sentiments.GET("/trends/insights", sr.GetTrendInsights)
		sentiments.GET("/trends/insights/:teamId", sr.GetTeamTrendInsights)
		sentiments.GET("/categories/stats", sr.GetCategoryStats)
		sentiments.POST("/test-grok", sr.TestGrokAI)
	}
}

func (sr *SentimentRoutes) AnalyzeText(c *gin.Context) {
	var req models.SentimentAnalysisRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Text is required for analysis",
		})
		return
	}

	result, err := sr.sentimentService.AnalyzeText(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Analysis completed successfully",
		"result":  result,
	})
}

func (sr *SentimentRoutes) AnalyzeBatch(c *gin.Context) {
	// Frontend'in ne gönderdiğini log'layalım
	body, _ := c.GetRawData()
	log.Printf("Received batch request: %s", string(body))
	
	// Body'yi restore et
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	
	// Basitçe unprocessed comment'ları çek ve analiz et
	// Frontend formatına bakmaksızın çalışsın
	sr.handleOldFormatBatch(c, struct {
		Texts []string `json:"texts"`
	}{Texts: []string{}})
}

// Eski format handler (sadece text'ler var)
func (sr *SentimentRoutes) handleOldFormatBatch(c *gin.Context, req struct {
	Texts []string `json:"texts"`
}) {
	// Unprocessed comment'ları çek
	comments, err := sr.getUnprocessedComments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get unprocessed comments",
			"message": err.Error(),
		})
		return
	}

	if len(comments) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":       "No unprocessed comments found",
			"total_texts":   0,
			"success_count": 0,
			"failed_count":  0,
		})
		return
	}

	// Comment text'lerini analiz et
	texts := make([]string, len(comments))
	for i, comment := range comments {
		texts[i] = comment.Text
	}

	results, err := sr.sentimentService.AnalyzeBatch(texts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Batch analysis failed",
			"message": err.Error(),
		})
		return
	}

	// Sonuçları database'e kaydet
	successCount := 0
	for i, result := range results {
		if result == nil {
			continue
		}

		// Sentiment'i database'e kaydet
		_, err = sr.sentimentService.SaveSentiment(comments[i].ID, comments[i].TeamID, result)
		if err != nil {
			log.Printf("Failed to save sentiment for comment %s: %v", comments[i].ID.Hex(), err)
			continue
		}

		// Comment'ı processed olarak işaretle
		err = sr.commentService.MarkAsProcessed(comments[i].ID)
		if err != nil {
			log.Printf("Failed to mark comment as processed %s: %v", comments[i].ID.Hex(), err)
		}

		successCount++
		log.Printf("Saved sentiment %s (%.3f confidence) for comment %s", result.Label, result.Confidence, comments[i].ID.Hex())
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Batch analysis completed",
		"total_texts":   len(comments),
		"success_count": successCount,
		"failed_count":  len(comments) - successCount,
	})
}

// Yeni format handler (comment ID'leri var)
func (sr *SentimentRoutes) handleNewFormatBatch(c *gin.Context, req struct {
	Comments []struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		TeamID string `json:"team_id"`
	} `json:"comments"`
}) {
	if len(req.Comments) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Maximum 50 comments can be analyzed at once",
		})
		return
	}

	// Texts çıkar ve analiz et
	texts := make([]string, len(req.Comments))
	for i, comment := range req.Comments {
		texts[i] = comment.Text
	}

	results, err := sr.sentimentService.AnalyzeBatch(texts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Batch analysis failed",
			"message": err.Error(),
		})
		return
	}

	// Sonuçları database'e kaydet
	successCount := 0
	
	for i, result := range results {
		if result == nil {
			continue
		}

		// Comment ID'yi parse et
		commentID, err := primitive.ObjectIDFromHex(req.Comments[i].ID)
		if err != nil {
			log.Printf("Invalid comment ID %s: %v", req.Comments[i].ID, err)
			continue
		}

		// Team ID'yi parse et  
		teamID, err := primitive.ObjectIDFromHex(req.Comments[i].TeamID)
		if err != nil {
			log.Printf("Invalid team ID %s: %v", req.Comments[i].TeamID, err)
			continue
		}

		// Sentiment'i database'e kaydet
		_, err = sr.sentimentService.SaveSentiment(commentID, teamID, result)
		if err != nil {
			log.Printf("Failed to save sentiment for comment %s: %v", req.Comments[i].ID, err)
			continue
		}

		successCount++
		log.Printf("Saved sentiment %s (%.3f confidence) for comment %s", result.Label, result.Confidence, req.Comments[i].ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Batch analysis completed",
		"total_texts":   len(req.Comments),
		"success_count": successCount,
		"failed_count":  len(req.Comments) - successCount,
	})
}

// Unprocessed comment'ları çeken helper fonksiyon
func (sr *SentimentRoutes) getUnprocessedComments() ([]models.Comment, error) {
	// Comment service'den sentiment'i olmayan comment'ları çek
	return sr.commentService.GetUnprocessedComments(50)
}

func (sr *SentimentRoutes) GetSentimentStats(c *gin.Context) {
	stats, err := sr.sentimentService.GetSentimentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get sentiment statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (sr *SentimentRoutes) GetTeamReport(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	if teamIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID is required",
		})
		return
	}

	teamId, err := primitive.ObjectIDFromHex(teamIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	var startDate, endDate time.Time
	
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid start date",
				"message": "Start date must be in YYYY-MM-DD format",
			})
			return
		}
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid end date",
				"message": "End date must be in YYYY-MM-DD format",
			})
			return
		}
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid date range",
			"message": "End date must be after start date",
		})
		return
	}

	report, err := sr.sentimentService.GenerateReport(teamId, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate report",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (sr *SentimentRoutes) SaveSentiment(c *gin.Context) {
	var req struct {
		CommentID string                  `json:"comment_id" binding:"required"`
		TeamID    string                  `json:"team_id" binding:"required"`
		Result    *models.SentimentResult `json:"result" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	commentID, err := primitive.ObjectIDFromHex(req.CommentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid comment ID",
			"message": "Comment ID must be a valid ObjectID",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(req.TeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	sentiment, err := sr.sentimentService.SaveSentiment(commentID, teamID, req.Result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save sentiment",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Sentiment saved successfully",
		"sentiment": sentiment,
	})
}

func (sr *SentimentRoutes) CleanupDuplicates(c *gin.Context) {
	log.Println("Starting cleanup of duplicate sentiments...")
	
	result, err := sr.sentimentService.CleanupDuplicates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup duplicates",
			"message": err.Error(),
		})
		return
	}

	if len(result.Errors) > 0 {
		log.Printf("Cleanup completed with %d errors", len(result.Errors))
		for _, errMsg := range result.Errors {
			log.Printf("Cleanup error: %s", errMsg)
		}
	}

	log.Printf("Cleanup successful: removed %d of %d duplicates", result.DuplicatesRemoved, result.TotalDuplicatesFound)

	c.JSON(http.StatusOK, gin.H{
		"message": "Duplicate cleanup completed successfully",
		"result":  result,
	})
}

// 🚀 YENİ GROK AI ENDPOİNT'LERİ

// Gelişmiş İstatistikler (Tüm Takımlar)
func (sr *SentimentRoutes) GetEnhancedSentimentStats(c *gin.Context) {
	stats, err := sr.enhancedAnalyticsService.GetEnhancedSentimentStats(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get enhanced statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Enhanced sentiment statistics retrieved successfully",
		"data":    stats,
	})
}

// Gelişmiş İstatistikler (Belirli Takım)
func (sr *SentimentRoutes) GetTeamEnhancedStats(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	teamId, err := primitive.ObjectIDFromHex(teamIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	stats, err := sr.enhancedAnalyticsService.GetEnhancedSentimentStats(&teamId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get enhanced team statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Enhanced team sentiment statistics retrieved successfully",
		"data":    stats,
	})
}

// Günlük Özet Oluştur
func (sr *SentimentRoutes) GenerateDailySummary(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	
	// Eğer teamId yoksa (Tüm Takımlar seçilmişse)
	if teamIdStr == "" {
		// Tüm takımlar için özet oluştur
		summary, err := sr.enhancedAnalyticsService.GenerateAllTeamsSummary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to generate daily summary",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Daily summary generated successfully",
			"data":    summary,
		})
		return
	}

	// Belirli takım için
	teamId, err := primitive.ObjectIDFromHex(teamIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	summary, err := sr.enhancedAnalyticsService.GenerateDailySummary(teamId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate daily summary",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Daily summary generated successfully",
		"data":    summary,
	})
}

// Trend İçgörüleri (Tüm Takımlar)
func (sr *SentimentRoutes) GetTrendInsights(c *gin.Context) {
	insights, err := sr.enhancedAnalyticsService.GenerateTrendInsights(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get trend insights",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trend insights retrieved successfully",
		"data":    insights,
	})
}

// Trend İçgörüleri (Belirli Takım)
func (sr *SentimentRoutes) GetTeamTrendInsights(c *gin.Context) {
	teamIdStr := c.Param("teamId")
	teamId, err := primitive.ObjectIDFromHex(teamIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid team ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	insights, err := sr.enhancedAnalyticsService.GenerateTrendInsights(&teamId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get team trend insights",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team trend insights retrieved successfully",
		"data":    insights,
	})
}

// Kategori İstatistikleri
func (sr *SentimentRoutes) GetCategoryStats(c *gin.Context) {
	teamIdStr := c.Query("team_id")
	var teamId *primitive.ObjectID
	
	if teamIdStr != "" {
		id, err := primitive.ObjectIDFromHex(teamIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid team ID",
				"message": "Team ID must be a valid ObjectID",
			})
			return
		}
		teamId = &id
	}

	stats, err := sr.enhancedAnalyticsService.GetEnhancedSentimentStats(teamId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get category statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Category statistics retrieved successfully",
		"categories": stats.CategoryBreakdown,
		"toxicity":   stats.ToxicityStats,
	})
}

// Grok AI Test Endpoint
func (sr *SentimentRoutes) TestGrokAI(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": "Text is required",
		})
		return
	}

	// Grok AI servisini test et
	groqService := services.NewGroqService()
	
	// 1. Enhanced sentiment analysis test
	groqResult, err := groqService.EnhancedSentimentAnalysis(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Grok AI test failed",
			"message": err.Error(),
		})
		return
	}

	// 2. Category analysis test
	category, keywords, catErr := groqService.CategorizeComment(req.Text)
	if catErr != nil {
		log.Printf("Category analysis failed: %v", catErr)
		category = "Test Failed"
		keywords = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grok AI test successful!",
		"results": gin.H{
			"sentiment_analysis": gin.H{
				"label":         groqResult.Enhanced.Label,
				"confidence":    groqResult.Enhanced.Confidence,
				"model":         groqResult.Enhanced.ModelUsed,
				"toxicity":      groqResult.ToxicityScore,
				"summary":       groqResult.Summary,
				"keywords":      groqResult.Keywords,
			},
			"categorization": gin.H{
				"category": category,
				"keywords": keywords,
			},
			"original_text": req.Text,
		},
	})
}