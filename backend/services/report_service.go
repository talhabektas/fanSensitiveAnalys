package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportService struct {
	commentsCollection   *mongo.Collection
	sentimentsCollection *mongo.Collection
	teamsCollection      *mongo.Collection
}

func NewReportService() *ReportService {
	return &ReportService{
		commentsCollection:   config.GetCollection("comments"),
		sentimentsCollection: config.GetCollection("sentiments"),
		teamsCollection:      config.GetCollection("teams"),
	}
}

func (rs *ReportService) GenerateTeamReport(teamID primitive.ObjectID, startDate, endDate time.Time) (*models.SentimentReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	team, err := rs.getTeamInfo(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("error getting team info: %w", err)
	}

	report := &models.SentimentReport{
		TeamID:   teamID,
		TeamName: team.Name,
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

	basicStats, err := rs.getBasicSentimentStats(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error getting basic stats: %w", err)
	}

	if basicStats != nil {
		report.TotalAnalyzed = basicStats.TotalAnalyzed
		report.AverageSentiment = basicStats.AverageSentiment
		report.SentimentCounts["POSITIVE"] = basicStats.PositiveCount
		report.SentimentCounts["NEGATIVE"] = basicStats.NegativeCount
		report.SentimentCounts["NEUTRAL"] = basicStats.NeutralCount
	}

	trendAnalysis, err := rs.getTrendAnalysis(ctx, teamID, startDate, endDate)
	if err != nil {
		log.Printf("Error getting trend analysis: %v", err)
	} else {
		report.TrendAnalysis = *trendAnalysis
	}

	topKeywords, err := rs.getTopKeywords(ctx, teamID, startDate, endDate)
	if err != nil {
		log.Printf("Error getting top keywords: %v", err)
	} else {
		report.TopKeywords = topKeywords
	}

	hourlyDist, err := rs.getHourlyDistribution(ctx, teamID, startDate, endDate)
	if err != nil {
		log.Printf("Error getting hourly distribution: %v", err)
	} else {
		report.HourlyDistribution = hourlyDist
	}

	sourceBreakdown, err := rs.getSourceBreakdown(ctx, teamID, startDate, endDate)
	if err != nil {
		log.Printf("Error getting source breakdown: %v", err)
	} else {
		report.SourceBreakdown = sourceBreakdown
	}

	return report, nil
}

func (rs *ReportService) GenerateOverallStats() (*models.SentimentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Sentiment stats from sentiments collection
	sentimentPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":                nil,
				"total_analyzed":     bson.M{"$sum": 1},
				"overall_sentiment":  bson.M{"$avg": "$score"},
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

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, sentimentPipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating sentiment stats: %w", err)
	}
	defer cursor.Close(ctx)

	var sentimentResult []struct {
		TotalAnalyzed     int64   `bson:"total_analyzed"`
		OverallSentiment  float64 `bson:"overall_sentiment"`
		AvgConfidence     float64 `bson:"avg_confidence"`
		PositiveCount     int64   `bson:"positive_count"`
		NegativeCount     int64   `bson:"negative_count"`
		NeutralCount      int64   `bson:"neutral_count"`
		HighConfidence    int64   `bson:"high_confidence"`
		MediumConfidence  int64   `bson:"medium_confidence"`
		LowConfidence     int64   `bson:"low_confidence"`
	}

	if err = cursor.All(ctx, &sentimentResult); err != nil {
		return nil, fmt.Errorf("error decoding sentiment stats: %w", err)
	}

	// Get total comments count from comments collection
	totalComments, err := rs.commentsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Error getting total comments count: %v", err)
		totalComments = 0
	}

	// Get comment counts per team
	teamCommentsPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$team_id",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "teams",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "team",
			},
		},
		{
			"$unwind": "$team",
		},
		{
			"$project": bson.M{
				"_id":       0,
				"team_name": "$team.name",
				"count":     1,
			},
		},
	}

	teamCursor, err := rs.commentsCollection.Aggregate(ctx, teamCommentsPipeline)
	if err != nil {
		log.Printf("Error getting team comment counts: %v", err)
	}
	
	var teamComments []struct {
		TeamName string `bson:"team_name"`
		Count    int64  `bson:"count"`
	}
	
	if teamCursor != nil {
		defer teamCursor.Close(ctx)
		if err = teamCursor.All(ctx, &teamComments); err != nil {
			log.Printf("Error decoding team comment counts: %v", err)
		}
	}

	stats := &models.SentimentStats{
		SentimentBreakdown: make(map[string]int64),
		ModelPerformance:   make(map[string]models.ModelPerformance),
		TeamComments:       make(map[string]int64), // Yeni field
	}

	// Set total comments count (from comments collection)
	stats.TotalComments = totalComments

	// Set team comment counts
	for _, teamComment := range teamComments {
		stats.TeamComments[teamComment.TeamName] = teamComment.Count
	}

	if len(sentimentResult) > 0 {
		r := sentimentResult[0]
		stats.TotalAnalyzed = r.TotalAnalyzed
		stats.OverallSentiment = r.OverallSentiment
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

	teamComparison, err := rs.getTeamComparison(ctx)
	if err == nil {
		stats.TeamComparison = teamComparison
	}

	recentTrends, err := rs.getRecentTrends(ctx)
	if err == nil {
		stats.RecentTrends = recentTrends
	}

	modelPerformance, err := rs.getModelPerformance(ctx)
	if err == nil {
		stats.ModelPerformance = modelPerformance
	}

	return stats, nil
}

func (rs *ReportService) GenerateTeamComparison(startDate, endDate time.Time) ([]models.TeamSentimentComparison, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
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

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error generating team comparison: %w", err)
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

func (rs *ReportService) GetDashboardData() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dashboardData := make(map[string]interface{})

	overallStats, err := rs.GenerateOverallStats()
	if err != nil {
		log.Printf("Error getting overall stats: %v", err)
		dashboardData["overall_stats"] = nil
	} else {
		dashboardData["overall_stats"] = overallStats
	}

	teamComparison, err := rs.GenerateTeamComparison(time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		log.Printf("Error getting team comparison: %v", err)
		dashboardData["team_comparison"] = nil
	} else {
		dashboardData["team_comparison"] = teamComparison
	}

	recentComments, err := rs.getRecentComments(ctx, 10)
	if err != nil {
		log.Printf("Error getting recent comments: %v", err)
		dashboardData["recent_comments"] = nil
	} else {
		dashboardData["recent_comments"] = recentComments
	}

	dailyTrends, err := rs.getDailyTrends(ctx, 30)
	if err != nil {
		log.Printf("Error getting daily trends: %v", err)
		dashboardData["daily_trends"] = nil
	} else {
		dashboardData["daily_trends"] = dailyTrends
	}

	return dashboardData, nil
}

func (rs *ReportService) getTeamInfo(ctx context.Context, teamID primitive.ObjectID) (*models.Team, error) {
	var team models.Team
	err := rs.teamsCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (rs *ReportService) getBasicSentimentStats(ctx context.Context, teamID primitive.ObjectID, startDate, endDate time.Time) (*struct {
	TotalAnalyzed    int64   `bson:"total_analyzed"`
	AverageSentiment float64 `bson:"avg_sentiment"`
	PositiveCount    int64   `bson:"positive_count"`
	NegativeCount    int64   `bson:"negative_count"`
	NeutralCount     int64   `bson:"neutral_count"`
}, error) {
	matchFilter := bson.M{
		"team_id": teamID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
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

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalAnalyzed    int64   `bson:"total_analyzed"`
		AverageSentiment float64 `bson:"avg_sentiment"`
		PositiveCount    int64   `bson:"positive_count"`
		NegativeCount    int64   `bson:"negative_count"`
		NeutralCount     int64   `bson:"neutral_count"`
	}
	
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	return &results[0], nil
}

func (rs *ReportService) getTrendAnalysis(ctx context.Context, teamID primitive.ObjectID, startDate, endDate time.Time) (*models.TrendAnalysis, error) {
	matchFilter := bson.M{
		"team_id": teamID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$created_at",
					},
				},
				"avg_score": bson.M{"$avg": "$score"},
				"count":     bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dailyScores []struct {
		Date     string  `bson:"_id"`
		AvgScore float64 `bson:"avg_score"`
		Count    int64   `bson:"count"`
	}

	if err = cursor.All(ctx, &dailyScores); err != nil {
		return nil, err
	}

	if len(dailyScores) < 2 {
		return &models.TrendAnalysis{
			Direction:     "stable",
			ChangePercent: 0,
		}, nil
	}

	firstScore := dailyScores[0].AvgScore
	lastScore := dailyScores[len(dailyScores)-1].AvgScore
	changePercent := ((lastScore - firstScore) / firstScore) * 100

	direction := "stable"
	if changePercent > 5 {
		direction = "improving"
	} else if changePercent < -5 {
		direction = "declining"
	}

	return &models.TrendAnalysis{
		Direction:     direction,
		ChangePercent: changePercent,
	}, nil
}

func (rs *ReportService) getTopKeywords(ctx context.Context, teamID primitive.ObjectID, startDate, endDate time.Time) ([]models.KeywordAnalysis, error) {
	return []models.KeywordAnalysis{}, nil
}

func (rs *ReportService) getHourlyDistribution(ctx context.Context, teamID primitive.ObjectID, startDate, endDate time.Time) (map[int]models.SentimentHourly, error) {
	matchFilter := bson.M{
		"team_id": teamID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{"$match": matchFilter},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$hour": "$created_at",
				},
				"count":     bson.M{"$sum": 1},
				"positive":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "POSITIVE"}}, 1, 0}}},
				"negative":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEGATIVE"}}, 1, 0}}},
				"neutral":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEUTRAL"}}, 1, 0}}},
				"avg_score": bson.M{"$avg": "$score"},
			},
		},
	}

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	hourlyDist := make(map[int]models.SentimentHourly)
	for cursor.Next(ctx) {
		var doc struct {
			Hour     int     `bson:"_id"`
			Count    int64   `bson:"count"`
			Positive int64   `bson:"positive"`
			Negative int64   `bson:"negative"`
			Neutral  int64   `bson:"neutral"`
			AvgScore float64 `bson:"avg_score"`
		}
		if err := cursor.Decode(&doc); err == nil {
			hourlyDist[doc.Hour] = models.SentimentHourly{
				Hour:     doc.Hour,
				Count:    doc.Count,
				Positive: doc.Positive,
				Negative: doc.Negative,
				Neutral:  doc.Neutral,
				AvgScore: doc.AvgScore,
			}
		}
	}

	return hourlyDist, nil
}

func (rs *ReportService) getSourceBreakdown(ctx context.Context, teamID primitive.ObjectID, startDate, endDate time.Time) (map[string]models.SentimentSourceStats, error) {
	matchFilter := bson.M{
		"team_id": teamID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "comments",
				"localField":   "comment_id",
				"foreignField": "_id",
				"as":           "comment",
			},
		},
		{
			"$unwind": "$comment",
		},
		{"$match": matchFilter},
		{
			"$group": bson.M{
				"_id":          "$comment.source",
				"count":        bson.M{"$sum": 1},
				"positive":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "POSITIVE"}}, 1, 0}}},
				"negative":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEGATIVE"}}, 1, 0}}},
				"neutral":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$label", "NEUTRAL"}}, 1, 0}}},
				"avg_sentiment": bson.M{"$avg": "$score"},
			},
		},
	}

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	sourceBreakdown := make(map[string]models.SentimentSourceStats)
	for cursor.Next(ctx) {
		var doc struct {
			Source       string  `bson:"_id"`
			Count        int64   `bson:"count"`
			Positive     int64   `bson:"positive"`
			Negative     int64   `bson:"negative"`
			Neutral      int64   `bson:"neutral"`
			AvgSentiment float64 `bson:"avg_sentiment"`
		}
		if err := cursor.Decode(&doc); err == nil {
			sourceBreakdown[doc.Source] = models.SentimentSourceStats{
				Source:       doc.Source,
				Count:        doc.Count,
				Positive:     doc.Positive,
				Negative:     doc.Negative,
				Neutral:      doc.Neutral,
				AvgSentiment: doc.AvgSentiment,
			}
		}
	}

	return sourceBreakdown, nil
}

func (rs *ReportService) getTeamComparison(ctx context.Context) ([]models.TeamSentimentComparison, error) {
	return rs.GenerateTeamComparison(time.Now().AddDate(0, 0, -30), time.Now())
}

func (rs *ReportService) getRecentTrends(ctx context.Context) ([]models.SentimentTrend, error) {
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

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
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

func (rs *ReportService) getModelPerformance(ctx context.Context) (map[string]models.ModelPerformance, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":               "$model_used",
				"total_processed":   bson.M{"$sum": 1},
				"avg_confidence":    bson.M{"$avg": "$confidence"},
				"avg_processing_time": bson.M{"$avg": "$analysis_details.processing_time"},
			},
		},
	}

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	performance := make(map[string]models.ModelPerformance)
	for cursor.Next(ctx) {
		var doc struct {
			Model              string  `bson:"_id"`
			TotalProcessed     int64   `bson:"total_processed"`
			AvgConfidence      float64 `bson:"avg_confidence"`
			AvgProcessingTime  float64 `bson:"avg_processing_time"`
		}
		if err := cursor.Decode(&doc); err == nil {
			performance[doc.Model] = models.ModelPerformance{
				Model:             doc.Model,
				TotalProcessed:    doc.TotalProcessed,
				AverageTime:       doc.AvgProcessingTime,
				SuccessRate:       0.95,
				AverageConfidence: doc.AvgConfidence,
			}
		}
	}

	return performance, nil
}

func (rs *ReportService) getRecentComments(ctx context.Context, limit int) ([]models.Comment, error) {
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := rs.commentsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (rs *ReportService) getDailyTrends(ctx context.Context, days int) ([]models.SentimentTrend, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": time.Now().AddDate(0, 0, -days),
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

	cursor, err := rs.sentimentsCollection.Aggregate(ctx, pipeline)
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