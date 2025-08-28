package routes

import (
	"context"
	"net/http"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"
	"taraftar-analizi/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TeamRoutes struct {
	reportService *services.ReportService
}

func NewTeamRoutes() *TeamRoutes {
	return &TeamRoutes{
		reportService: services.NewReportService(),
	}
}

func (tr *TeamRoutes) RegisterRoutes(router *gin.RouterGroup) {
	teams := router.Group("/teams")
	{
		teams.GET("", tr.GetTeams)
		teams.POST("", tr.CreateTeam)
		teams.GET("/:id", tr.GetTeam)
		teams.PUT("/:id", tr.UpdateTeam)
		teams.DELETE("/:id", tr.DeleteTeam)
		teams.GET("/:id/sentiment", tr.GetTeamSentiment)
		teams.GET("/:id/stats", tr.GetTeamStats)
		teams.POST("/seed", tr.SeedTeams)
	}
}

func (tr *TeamRoutes) GetTeams(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")
	
	findOptions := options.Find().SetSort(bson.D{{"name", 1}})
	
	cursor, err := collection.Find(ctx, bson.M{"is_active": true}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get teams",
			"message": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err = cursor.All(ctx, &teams); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to decode teams",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
		"count": len(teams),
	})
}

func (tr *TeamRoutes) CreateTeam(c *gin.Context) {
	var req models.TeamCreateRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")

	existingCount, err := collection.CountDocuments(ctx, bson.M{"slug": req.Slug})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check existing teams",
			"message": err.Error(),
		})
		return
	}

	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Team already exists",
			"message": "A team with this slug already exists",
		})
		return
	}

	team := &models.Team{
		Name:       req.Name,
		Slug:       req.Slug,
		League:     req.League,
		Country:    req.Country,
		Logo:       req.Logo,
		Colors:     req.Colors,
		Keywords:   req.Keywords,
		Subreddits: req.Subreddits,
	}

	team.BeforeCreate()

	_, err = collection.InsertOne(ctx, team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create team",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"team":    team,
	})
}

func (tr *TeamRoutes) GetTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID is required",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")
	
	var team models.Team
	err = collection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Team not found",
			"message": "Team with specified ID does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (tr *TeamRoutes) UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID is required",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	var req models.TeamUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.Name != nil {
		update["$set"].(bson.M)["name"] = *req.Name
	}
	if req.League != nil {
		update["$set"].(bson.M)["league"] = *req.League
	}
	if req.Country != nil {
		update["$set"].(bson.M)["country"] = *req.Country
	}
	if req.Logo != nil {
		update["$set"].(bson.M)["logo"] = *req.Logo
	}
	if req.Colors != nil {
		update["$set"].(bson.M)["colors"] = *req.Colors
	}
	if req.Keywords != nil {
		update["$set"].(bson.M)["keywords"] = *req.Keywords
	}
	if req.Subreddits != nil {
		update["$set"].(bson.M)["subreddits"] = *req.Subreddits
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["is_active"] = *req.IsActive
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": teamID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update team",
			"message": err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Team not found",
			"message": "Team with specified ID does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team updated successfully",
	})
}

func (tr *TeamRoutes) DeleteTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID is required",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": teamID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete team",
			"message": err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Team not found",
			"message": "Team with specified ID does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team deleted successfully",
	})
}

func (tr *TeamRoutes) GetTeamSentiment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID is required",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	report, err := tr.reportService.GenerateTeamReport(teamID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get team sentiment",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (tr *TeamRoutes) GetTeamStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID is required",
		})
		return
	}

	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Team ID must be a valid ObjectID",
		})
		return
	}

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	report, err := tr.reportService.GenerateTeamReport(teamID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get team stats",
			"message": err.Error(),
		})
		return
	}

	stats := models.TeamStats{
		TeamID:        teamID,
		TeamName:      report.TeamName,
		TotalComments: report.TotalAnalyzed,
		AvgSentiment:  report.AverageSentiment,
	}

	if positive, exists := report.SentimentCounts["POSITIVE"]; exists {
		stats.PositiveCount = positive
	}
	if negative, exists := report.SentimentCounts["NEGATIVE"]; exists {
		stats.NegativeCount = negative
	}
	if neutral, exists := report.SentimentCounts["NEUTRAL"]; exists {
		stats.NeutralCount = neutral
	}

	stats.SentimentTrend = report.TrendAnalysis.Direction

	c.JSON(http.StatusOK, stats)
}

func (tr *TeamRoutes) SeedTeams(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := config.GetCollection("teams")

	var insertedCount int
	for _, team := range models.TurkishTeams {
		existingCount, err := collection.CountDocuments(ctx, bson.M{"slug": team.Slug})
		if err != nil {
			continue
		}
		
		if existingCount > 0 {
			continue
		}

		team.BeforeCreate()
		
		_, err = collection.InsertOne(ctx, team)
		if err != nil {
			continue
		}
		
		insertedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Teams seeded successfully",
		"inserted_count": insertedCount,
		"total_teams":    len(models.TurkishTeams),
	})
}