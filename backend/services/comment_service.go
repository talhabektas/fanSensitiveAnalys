package services

import (
	"context"
	"errors"
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

type CommentService struct {
	collection *mongo.Collection
}

func NewCommentService() *CommentService {
	return &CommentService{
		collection: config.GetCollection("comments"),
	}
}

func (cs *CommentService) CreateComment(req models.CommentCreateRequest) (*models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := cs.CheckDuplicate(req.SourceID, req.Source)
	if err != nil {
		return nil, fmt.Errorf("error checking duplicate: %w", err)
	}
	if exists {
		return nil, errors.New("comment already exists")
	}

	comment := &models.Comment{
		SourceID:    req.SourceID,
		Source:      req.Source,
		Author:      req.Author,
		Text:        strings.TrimSpace(req.Text),
		URL:         req.URL,
		Score:       req.Score,
		ParentID:    req.ParentID,
		Subreddit:   req.Subreddit,
		Language:    req.Language,
		Metadata:    req.Metadata,
		IsProcessed: false,
		HasSentiment: false,
	}

	if req.TeamID != "" {
		teamID, err := primitive.ObjectIDFromHex(req.TeamID)
		if err != nil {
			return nil, fmt.Errorf("invalid team ID: %w", err)
		}
		comment.TeamID = teamID
	} else {
		comment.TeamID = cs.detectTeamFromText(comment.Text)
	}

	comment.BeforeCreate()

	_, err = cs.collection.InsertOne(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("error creating comment: %w", err)
	}

	log.Printf("Created comment: %s from %s", comment.ID.Hex(), comment.Source)
	return comment, nil
}

func (cs *CommentService) GetComments(query models.CommentQuery) (*models.CommentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := cs.buildFilter(query)
	
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	skip := (query.Page - 1) * query.Limit

	sortField := "created_at"
	sortOrder := -1
	if query.SortBy != "" {
		sortField = query.SortBy
	}
	if query.SortOrder == "asc" {
		sortOrder = 1
	}

	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{sortField, sortOrder}})

	cursor, err := cs.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error finding comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, fmt.Errorf("error decoding comments: %w", err)
	}

	total, err := cs.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error counting comments: %w", err)
	}

	totalPages := int((total + int64(query.Limit) - 1) / int64(query.Limit))

	return &models.CommentResponse{
		Comments:   comments,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}

func (cs *CommentService) GetUnprocessedComments(limit int) ([]models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	filter := bson.M{
		"has_sentiment": false,
		"language":      "tr",
		"text":          bson.M{"$ne": ""},
	}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at", 1}})

	cursor, err := cs.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error finding unprocessed comments: %w", err)
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, fmt.Errorf("error decoding comments: %w", err)
	}

	return comments, nil
}

func (cs *CommentService) UpdateComment(id string, req models.CommentUpdateRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	commentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.IsProcessed != nil {
		update["$set"].(bson.M)["is_processed"] = *req.IsProcessed
	}
	if req.HasSentiment != nil {
		update["$set"].(bson.M)["has_sentiment"] = *req.HasSentiment
	}
	if req.Sentiment != nil {
		update["$set"].(bson.M)["sentiment"] = *req.Sentiment
		update["$set"].(bson.M)["has_sentiment"] = true
	}
	if req.TeamID != nil {
		teamID, err := primitive.ObjectIDFromHex(*req.TeamID)
		if err != nil {
			return fmt.Errorf("invalid team ID: %w", err)
		}
		update["$set"].(bson.M)["team_id"] = teamID
	}
	if req.Language != nil {
		update["$set"].(bson.M)["language"] = *req.Language
	}

	result, err := cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	if err != nil {
		return fmt.Errorf("error updating comment: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (cs *CommentService) BulkUpdateProcessed(commentIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, idStr := range commentIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, id)
	}

	if len(objectIDs) == 0 {
		return errors.New("no valid comment IDs provided")
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	update := bson.M{
		"$set": bson.M{
			"is_processed": true,
			"updated_at":   time.Now(),
		},
	}

	result, err := cs.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error bulk updating comments: %w", err)
	}

	log.Printf("Bulk updated %d comments as processed", result.ModifiedCount)
	return nil
}

func (cs *CommentService) MarkAsProcessed(commentID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_processed": true,
			"has_sentiment": true,
			"updated_at": time.Now(),
		},
	}

	result, err := cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	if err != nil {
		return fmt.Errorf("error marking comment as processed: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("comment not found")
	}

	log.Printf("Marked comment %s as processed", commentID.Hex())
	return nil
}

func (cs *CommentService) GetCommentStats() (*models.CommentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":                nil,
				"total_comments":     bson.M{"$sum": 1},
				"processed_comments": bson.M{"$sum": bson.M{"$cond": []interface{}{"$is_processed", 1, 0}}},
			},
		},
	}

	cursor, err := cs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating comment stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalComments     int64 `bson:"total_comments"`
		ProcessedComments int64 `bson:"processed_comments"`
	}

	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("error decoding stats: %w", err)
	}

	stats := &models.CommentStats{
		SentimentBreakdown: make(map[string]int64),
		SourceBreakdown:    make(map[string]int64),
		LanguageBreakdown:  make(map[string]int64),
	}

	if len(result) > 0 {
		stats.TotalComments = result[0].TotalComments
		stats.ProcessedComments = result[0].ProcessedComments
		stats.UnprocessedComments = stats.TotalComments - stats.ProcessedComments
	}

	sentimentStats, err := cs.getSentimentBreakdown(ctx)
	if err == nil {
		stats.SentimentBreakdown = sentimentStats
	}

	sourceStats, err := cs.getSourceBreakdown(ctx)
	if err == nil {
		stats.SourceBreakdown = sourceStats
	}

	languageStats, err := cs.getLanguageBreakdown(ctx)
	if err == nil {
		stats.LanguageBreakdown = languageStats
	}

	dailyStats, err := cs.getDailyStats(ctx)
	if err == nil {
		stats.DailyStats = dailyStats
	}

	return stats, nil
}

func (cs *CommentService) CheckDuplicate(sourceID, source string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"source_id": sourceID,
		"source":    source,
	}

	count, err := cs.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("error checking duplicate: %w", err)
	}

	return count > 0, nil
}

func (cs *CommentService) buildFilter(query models.CommentQuery) bson.M {
	filter := bson.M{}

	if query.TeamID != "" {
		teamID, err := primitive.ObjectIDFromHex(query.TeamID)
		if err == nil {
			filter["team_id"] = teamID
		}
	}

	if query.Source != "" {
		filter["source"] = query.Source
	}

	if query.Author != "" {
		filter["author"] = bson.M{"$regex": query.Author, "$options": "i"}
	}

	if query.Language != "" {
		filter["language"] = query.Language
	}

	if query.IsProcessed != nil {
		filter["is_processed"] = *query.IsProcessed
	}

	if query.HasSentiment != nil {
		filter["has_sentiment"] = *query.HasSentiment
	}

	if query.Sentiment != "" {
		filter["sentiment.label"] = strings.ToUpper(query.Sentiment)
	}

	if !query.StartDate.IsZero() || !query.EndDate.IsZero() {
		dateFilter := bson.M{}
		if !query.StartDate.IsZero() {
			dateFilter["$gte"] = query.StartDate
		}
		if !query.EndDate.IsZero() {
			dateFilter["$lte"] = query.EndDate.Add(24 * time.Hour)
		}
		filter["created_at"] = dateFilter
	}

	if query.Search != "" {
		filter["$text"] = bson.M{"$search": query.Search}
	}

	return filter
}

func (cs *CommentService) detectTeamFromText(text string) primitive.ObjectID {
	text = strings.ToLower(text)
	
	for _, team := range models.TurkishTeams {
		for _, keyword := range team.Keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return team.ID
			}
		}
	}
	
	return primitive.NilObjectID
}

func (cs *CommentService) getSentimentBreakdown(ctx context.Context) (map[string]int64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{"sentiment.label": bson.M{"$exists": true}},
		},
		{
			"$group": bson.M{
				"_id":   "$sentiment.label",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := cs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			result[doc.ID] = doc.Count
		}
	}

	return result, nil
}

func (cs *CommentService) getSourceBreakdown(ctx context.Context) (map[string]int64, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$source",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := cs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			result[doc.ID] = doc.Count
		}
	}

	return result, nil
}

func (cs *CommentService) getLanguageBreakdown(ctx context.Context) (map[string]int64, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$language",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := cs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			result[doc.ID] = doc.Count
		}
	}

	return result, nil
}

func (cs *CommentService) getDailyStats(ctx context.Context) ([]models.DailyStat, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{
					"$gte": time.Now().AddDate(0, 0, -30),
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
				"count":    bson.M{"$sum": 1},
				"positive": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$sentiment.label", "POSITIVE"}}, 1, 0}}},
				"negative": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$sentiment.label", "NEGATIVE"}}, 1, 0}}},
				"neutral":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$sentiment.label", "NEUTRAL"}}, 1, 0}}},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := cs.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []models.DailyStat
	for cursor.Next(ctx) {
		var doc struct {
			Date     string `bson:"_id"`
			Count    int64  `bson:"count"`
			Positive int64  `bson:"positive"`
			Negative int64  `bson:"negative"`
			Neutral  int64  `bson:"neutral"`
		}
		if err := cursor.Decode(&doc); err == nil {
			stats = append(stats, models.DailyStat{
				Date:     doc.Date,
				Count:    doc.Count,
				Positive: doc.Positive,
				Negative: doc.Negative,
				Neutral:  doc.Neutral,
			})
		}
	}

	return stats, nil
}

// GetCommentByTextAndAuthor - YouTube yorumları için duplicate kontrolü
func (cs *CommentService) GetCommentByTextAndAuthor(text, author string) (*models.Comment, error) {
	var comment models.Comment
	
	filter := bson.M{
		"text":   text,
		"author": author,
	}

	err := cs.collection.FindOne(context.TODO(), filter).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Bulunamadı, duplicate değil
		}
		return nil, fmt.Errorf("failed to find comment: %v", err)
	}

	return &comment, nil
}