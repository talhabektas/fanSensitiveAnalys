package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/middleware"
	"taraftar-analizi/models"
	"taraftar-analizi/routes"
	"taraftar-analizi/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// UTF-8 encoding ayarları
	os.Setenv("LANG", "tr_TR.UTF-8")
	os.Setenv("LC_ALL", "tr_TR.UTF-8")
	
	config.LoadConfig()
	
	gin.SetMode(config.AppConfig.GinMode)

	config.ConnectDatabase()
	defer config.DisconnectDatabase()

	router := setupRouter()

	server := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", config.AppConfig.Port)
		log.Printf("Frontend URL: %s", config.AppConfig.FrontendURL)
		log.Printf("MongoDB Database: %s", config.AppConfig.MongoDatabase)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.APIKeyMiddleware())
	router.Use(middleware.ErrorMiddleware())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Taraftar Duygu Analizi API",
			"version":    "1.0.0",
			"status":     "running",
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	})

	router.GET("/health", healthCheck)

	api := router.Group("/api/v1")
	
	commentRoutes := routes.NewCommentRoutes()
	commentRoutes.RegisterRoutes(api)

	sentimentRoutes := routes.NewSentimentRoutes()
	sentimentRoutes.RegisterRoutes(api)

	teamRoutes := routes.NewTeamRoutes()
	teamRoutes.RegisterRoutes(api)

	redditRoutes := setupRedditRoutes()
	redditRoutes(api)

	dashboardRoutes := setupDashboardRoutes()
	dashboardRoutes(api)

	webhookRoutes := setupWebhookRoutes()
	webhookRoutes(api)

	liveRoutes := setupLiveRoutes()
	liveRoutes(api)

	youtubeRoutes := setupYouTubeRoutes()
	youtubeRoutes(api)

	trendRoutes := setupTrendRoutes()
	trendRoutes(api)

	reportRoutes := setupReportRoutes()
	reportRoutes(api)

	// N8N Webhook Routes
	n8nRouter := routes.NewN8NWebhookRouter()
	n8nRouter.SetupRoutes(router)

	return router
}

func setupLiveRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		live := router.Group("/live")
		{
			live.POST("/reddit/start", func(c *gin.Context) {
				var req struct {
					Subreddits []string `json:"subreddits" binding:"required"`
					Interval   int      `json:"interval_minutes"`
					Limit      int      `json:"limit"`
				}
				
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid request body",
						"message": err.Error(),
					})
					return
				}

				// Default values
				if req.Interval == 0 {
					req.Interval = 5 // 5 dakika
				}
				if req.Limit == 0 {
					req.Limit = 25
				}

				// Takım ID'lerini al
				teamService := services.NewTeamService()
				teams, err := teamService.GetAllTeams()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get teams",
						"message": err.Error(),
					})
					return
				}

				// Team mapping oluştur
				teamMapping := make(map[string]string)
				for _, team := range teams {
					teamMapping[strings.ToLower(team.Name)] = team.ID.Hex()
				}

				liveService := services.NewRedditLiveService()
				if liveService == nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to initialize Reddit service",
						"message": "Check Reddit API credentials",
					})
					return
				}

				config := services.LiveStreamConfig{
					Subreddits: req.Subreddits,
					Teams:      teamMapping,
					Interval:   time.Duration(req.Interval) * time.Minute,
					Limit:      req.Limit,
				}

				err = liveService.StartLiveStream(config)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to start live stream",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":     "Reddit live stream started",
					"subreddits":  req.Subreddits,
					"interval":    req.Interval,
					"limit":       req.Limit,
					"teams_count": len(teamMapping),
				})
			})

			live.POST("/reddit/stop", func(c *gin.Context) {
				liveService := services.NewRedditLiveService()
				if liveService == nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Reddit service not available",
					})
					return
				}

				liveService.StopLiveStream()

				c.JSON(http.StatusOK, gin.H{
					"message": "Reddit live stream stopped",
				})
			})

			live.GET("/reddit/status", func(c *gin.Context) {
				liveService := services.NewRedditLiveService()
				if liveService == nil {
					c.JSON(http.StatusOK, gin.H{
						"is_running":   false,
						"client_ready": false,
						"error":        "Service not initialized",
					})
					return
				}

				status := liveService.GetStreamStatus()
				c.JSON(http.StatusOK, status)
			})
		}
	}
}

func healthCheck(c *gin.Context) {
	dbHealth := "healthy"
	if err := config.HealthCheck(); err != nil {
		dbHealth = "unhealthy: " + err.Error()
	}

	huggingfaceHealth := "not configured"
	if config.AppConfig.HuggingFaceToken != "" {
		huggingfaceHealth = "configured"
	}

	redditHealth := "not configured"
	if config.AppConfig.RedditClientID != "" && config.AppConfig.RedditClientSecret != "" {
		redditHealth = "configured"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": gin.H{
			"database":     dbHealth,
			"huggingface":  huggingfaceHealth,
			"reddit":       redditHealth,
		},
		"version": "1.0.0",
	})
}

func setupRedditRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		reddit := router.Group("/reddit")
		{
			reddit.POST("/collect", func(c *gin.Context) {
				redditService := services.NewRedditService()
				
				comments, err := redditService.CollectFromAllSubreddits()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to collect from Reddit",
						"message": err.Error(),
					})
					return
				}

				commentService := services.NewCommentService()
				savedCount := 0
				
				for _, comment := range comments {
					_, err := commentService.CreateComment(comment)
					if err == nil {
						savedCount++
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"message":      "Reddit collection completed",
					"collected":    len(comments),
					"saved":        savedCount,
					"duplicates":   len(comments) - savedCount,
				})
			})

			reddit.POST("/subreddit/:name", func(c *gin.Context) {
				subredditName := c.Param("name")
				if subredditName == "" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid subreddit",
						"message": "Subreddit name is required",
					})
					return
				}

				redditService := services.NewRedditService()
				
				comments, err := redditService.GetSubredditPosts(subredditName, 50)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to collect from subreddit",
						"message": err.Error(),
					})
					return
				}

				commentService := services.NewCommentService()
				savedCount := 0
				
				for _, comment := range comments {
					_, err := commentService.CreateComment(comment)
					if err == nil {
						savedCount++
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"message":    "Subreddit collection completed",
					"subreddit":  subredditName,
					"collected":  len(comments),
					"saved":      savedCount,
					"duplicates": len(comments) - savedCount,
				})
			})
		}
	}
}

func setupDashboardRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		dashboard := router.Group("/dashboard")
		{
			dashboard.GET("/data", func(c *gin.Context) {
				reportService := services.NewReportService()
				
				dashboardData, err := reportService.GetDashboardData()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get dashboard data",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, dashboardData)
			})

			dashboard.GET("/stats", func(c *gin.Context) {
				reportService := services.NewReportService()
				
				stats, err := reportService.GenerateOverallStats()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get stats",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, stats)
			})

			dashboard.GET("/comparison", func(c *gin.Context) {
				reportService := services.NewReportService()
				
				startDate := time.Now().AddDate(0, 0, -30)
				endDate := time.Now()

				comparison, err := reportService.GenerateTeamComparison(startDate, endDate)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get team comparison",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"teams":      comparison,
					"period":     "last_30_days",
					"start_date": startDate.Format("2006-01-02"),
					"end_date":   endDate.Format("2006-01-02"),
				})
			})
		}
	}
}

func setupWebhookRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		webhook := router.Group("/webhook")
		{
			webhook.POST("/comment", func(c *gin.Context) {
				var req models.CommentCreateRequest
				
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid request body",
						"message": err.Error(),
					})
					return
				}

				commentService := services.NewCommentService()
				
				comment, err := commentService.CreateComment(req)
				if err != nil {
					if err.Error() == "comment already exists" {
						c.JSON(http.StatusOK, gin.H{
							"message": "Comment already exists, skipping",
							"duplicate": true,
						})
						return
					}
					
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create comment",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"message":    "Comment created successfully",
					"comment_id": comment.ID.Hex(),
				})
			})

			webhook.POST("/sentiment", func(c *gin.Context) {
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

				sentimentService := services.NewSentimentService()
				
				sentiment, err := sentimentService.SaveSentiment(commentID, teamID, req.Result)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to save sentiment",
						"message": err.Error(),
					})
					return
				}

				commentService := services.NewCommentService()
				updateReq := models.CommentUpdateRequest{
					HasSentiment: &[]bool{true}[0],
					Sentiment:    req.Result,
				}
				
				if err := commentService.UpdateComment(req.CommentID, updateReq); err != nil {
					log.Printf("Warning: Failed to update comment with sentiment: %v", err)
				}

				c.JSON(http.StatusCreated, gin.H{
					"message":      "Sentiment saved successfully",
					"sentiment_id": sentiment.ID.Hex(),
				})
			})

			webhook.GET("/unprocessed", func(c *gin.Context) {
				limitStr := c.DefaultQuery("limit", "50")
				limit, err := strconv.Atoi(limitStr)
				if err != nil {
					limit = 50
				}

				commentService := services.NewCommentService()
				
				comments, err := commentService.GetUnprocessedComments(limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get unprocessed comments",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"comments": comments,
					"count":    len(comments),
					"limit":    limit,
				})
			})
		}
	}
}

func setupYouTubeRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		youtube := router.Group("/youtube")
		{
			youtube.POST("/collect", func(c *gin.Context) {
				youtubeService := services.NewYouTubeService()
				
				err := youtubeService.CollectTurkishFootballComments()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to collect YouTube comments",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "YouTube comments collection started",
					"status":  "success",
				})
			})
		}
	}
}

func setupTrendRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		trends := router.Group("/trends")
		{
			trends.GET("/analysis", func(c *gin.Context) {
				period := c.DefaultQuery("period", "7d")
				
				// Add no-cache headers to prevent caching of trend data
				c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.Header("X-Timestamp", time.Now().Format(time.RFC3339))
				
				if period != "7d" && period != "30d" && period != "90d" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid period. Must be 7d, 30d, or 90d",
					})
					return
				}

				trendService := services.NewTrendService()
				analysis, err := trendService.GetTrendAnalysis(period)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get trend analysis",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, analysis)
			})

			trends.GET("/insights", func(c *gin.Context) {
				period := c.DefaultQuery("period", "7d")
				
				// Add no-cache headers to prevent caching of insight data
				c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.Header("X-Timestamp", time.Now().Format(time.RFC3339))
				
				trendService := services.NewTrendService()
				insights, err := trendService.GenerateInsights(period)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to generate insights",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"insights": insights,
					"period":   period,
					"count":    len(insights),
				})
			})
			
			// Debug endpoint for data verification
			trends.GET("/debug", func(c *gin.Context) {
				// Add no-cache headers
				c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.Header("X-Timestamp", time.Now().Format(time.RFC3339))
				
				commentService := services.NewCommentService()
				
				// Get recent comments count
				recentComments, _ := commentService.GetComments(models.CommentQuery{
					StartDate: time.Now().AddDate(0, 0, -7),
					EndDate:   time.Now(),
					Limit:     10,
				})
				
				// Get recent sentiments - basit bir count
				sentimentCount, _ := config.GetCollection("sentiments").CountDocuments(
					context.TODO(),
					primitive.M{
						"created_at": primitive.M{
							"$gte": time.Now().AddDate(0, 0, -7),
						},
					},
				)
				
				// Count comments by date for last 7 days
				dateRanges := make(map[string]int)
				for _, comment := range recentComments.Comments {
					dateKey := comment.CreatedAt.Format("2006-01-02")
					dateRanges[dateKey]++
				}
				
				c.JSON(http.StatusOK, gin.H{
					"current_time":        time.Now().Format(time.RFC3339),
					"server_timezone":     time.Now().Location().String(),
					"recent_comments":     len(recentComments.Comments),
					"recent_sentiments":   sentimentCount,
					"last_7_days_start":   time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
					"last_7_days_end":     time.Now().Format(time.RFC3339),
					"comments_by_date":    dateRanges,
					"sample_recent_comments": recentComments.Comments,
					"data_freshness": gin.H{
						"has_today_data": dateRanges[time.Now().Format("2006-01-02")] > 0,
						"has_yesterday_data": dateRanges[time.Now().AddDate(0, 0, -1).Format("2006-01-02")] > 0,
						"total_days_with_data": len(dateRanges),
					},
				})
			})
		}
	}
}

func setupReportRoutes() func(*gin.RouterGroup) {
	return func(router *gin.RouterGroup) {
		reports := router.Group("/reports")
		{
			reports.GET("/executive", func(c *gin.Context) {
				period := c.DefaultQuery("period", "30d")
				
				if period != "7d" && period != "30d" && period != "90d" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid period. Must be 7d, 30d, or 90d",
					})
					return
				}

				reportService := services.NewReportGeneratorService()
				report, err := reportService.GenerateExecutiveReport(period)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to generate executive report",
						"message": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, report)
			})

			reports.GET("/executive/html", func(c *gin.Context) {
				period := c.DefaultQuery("period", "30d")
				
				if period != "7d" && period != "30d" && period != "90d" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid period. Must be 7d, 30d, or 90d",
					})
					return
				}

				reportService := services.NewReportGeneratorService()
				htmlReport, err := reportService.GenerateHTMLReport(period)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to generate HTML report",
						"message": err.Error(),
					})
					return
				}

				c.Header("Content-Type", "text/html; charset=utf-8")
				c.Header("Content-Disposition", "attachment; filename=executive-report.html")
				c.String(http.StatusOK, htmlReport)
			})

			reports.GET("/executive/download", func(c *gin.Context) {
				period := c.DefaultQuery("period", "30d")
				format := c.DefaultQuery("format", "html")
				
				if period != "7d" && period != "30d" && period != "90d" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid period. Must be 7d, 30d, or 90d",
					})
					return
				}

				reportService := services.NewReportGeneratorService()
				
				if format == "json" {
					report, err := reportService.GenerateExecutiveReport(period)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to generate report",
							"message": err.Error(),
						})
						return
					}

					c.Header("Content-Type", "application/json")
					c.Header("Content-Disposition", "attachment; filename=executive-report.json")
					c.JSON(http.StatusOK, report)
				} else {
					htmlReport, err := reportService.GenerateHTMLReport(period)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to generate HTML report",
							"message": err.Error(),
						})
						return
					}

					c.Header("Content-Type", "text/html; charset=utf-8")
					c.Header("Content-Disposition", "attachment; filename=executive-report.html")
					c.String(http.StatusOK, htmlReport)
				}
			})
		}
	}
}

