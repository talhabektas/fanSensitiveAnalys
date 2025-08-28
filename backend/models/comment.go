package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SourceID     string             `json:"source_id" bson:"source_id" validate:"required"`
	Source       string             `json:"source" bson:"source" validate:"required,oneof=reddit youtube twitter instagram"`
	TeamID       primitive.ObjectID `json:"team_id" bson:"team_id"`
	Author       string             `json:"author" bson:"author"`
	Text         string             `json:"text" bson:"text" validate:"required,min=5,max=5000"`
	URL          string             `json:"url" bson:"url"`
	Score        int64              `json:"score" bson:"score"`
	ParentID     string             `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Subreddit    string             `json:"subreddit,omitempty" bson:"subreddit,omitempty"`
	Language     string             `json:"language" bson:"language"`
	IsProcessed  bool               `json:"is_processed" bson:"is_processed"`
	HasSentiment bool               `json:"has_sentiment" bson:"has_sentiment"`
	Sentiment    *SentimentResult   `json:"sentiment,omitempty" bson:"sentiment,omitempty"`
	Metadata     CommentMetadata    `json:"metadata" bson:"metadata"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

type CommentMetadata struct {
	Platform      string            `json:"platform" bson:"platform"`
	PostID        string            `json:"post_id,omitempty" bson:"post_id,omitempty"`
	ThreadID      string            `json:"thread_id,omitempty" bson:"thread_id,omitempty"`
	IsReply       bool              `json:"is_reply" bson:"is_reply"`
	ReplyCount    int64             `json:"reply_count" bson:"reply_count"`
	LikeCount     int64             `json:"like_count" bson:"like_count"`
	Tags          []string          `json:"tags" bson:"tags"`
	ProcessedBy   string            `json:"processed_by" bson:"processed_by"`
	Quality       string            `json:"quality" bson:"quality"`
	DetectedTeams []string          `json:"detected_teams" bson:"detected_teams"`
	CustomFields  map[string]string `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
}

type SentimentResult struct {
	Label      string    `json:"label" bson:"label" validate:"required,oneof=POSITIVE NEGATIVE NEUTRAL"`
	Score      float64   `json:"score" bson:"score" validate:"required,gte=0,lte=1"`
	Confidence float64   `json:"confidence" bson:"confidence" validate:"required,gte=0,lte=1"`
	ModelUsed  string    `json:"model_used" bson:"model_used"`
	ProcessedAt time.Time `json:"processed_at" bson:"processed_at"`
}

type CommentCreateRequest struct {
	SourceID    string            `json:"source_id" validate:"required"`
	Source      string            `json:"source" validate:"required,oneof=reddit youtube twitter instagram"`
	TeamID      string            `json:"team_id,omitempty"`
	Author      string            `json:"author"`
	Text        string            `json:"text" validate:"required,min=5,max=5000"`
	URL         string            `json:"url"`
	Score       int64             `json:"score"`
	ParentID    string            `json:"parent_id,omitempty"`
	Subreddit   string            `json:"subreddit,omitempty"`
	Language    string            `json:"language"`
	Metadata    CommentMetadata   `json:"metadata"`
}

type CommentUpdateRequest struct {
	IsProcessed  *bool            `json:"is_processed,omitempty"`
	HasSentiment *bool            `json:"has_sentiment,omitempty"`
	Sentiment    *SentimentResult `json:"sentiment,omitempty"`
	TeamID       *string          `json:"team_id,omitempty"`
	Language     *string          `json:"language,omitempty"`
}

type CommentQuery struct {
	TeamID       string    `json:"team_id,omitempty" form:"team_id"`
	Source       string    `json:"source,omitempty" form:"source"`
	Author       string    `json:"author,omitempty" form:"author"`
	Language     string    `json:"language,omitempty" form:"language"`
	IsProcessed  *bool     `json:"is_processed,omitempty" form:"is_processed"`
	HasSentiment *bool     `json:"has_sentiment,omitempty" form:"has_sentiment"`
	Sentiment    string    `json:"sentiment,omitempty" form:"sentiment"`
	StartDate    time.Time `json:"start_date,omitempty" form:"start_date" time_format:"2006-01-02"`
	EndDate      time.Time `json:"end_date,omitempty" form:"end_date" time_format:"2006-01-02"`
	Search       string    `json:"search,omitempty" form:"search"`
	Page         int       `json:"page,omitempty" form:"page"`
	Limit        int       `json:"limit,omitempty" form:"limit"`
	SortBy       string    `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder    string    `json:"sort_order,omitempty" form:"sort_order"`
}

type CommentResponse struct {
	Comments   []Comment `json:"comments"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CommentStats struct {
	TotalComments     int64              `json:"total_comments"`
	ProcessedComments int64              `json:"processed_comments"`
	UnprocessedComments int64            `json:"unprocessed_comments"`
	SentimentBreakdown map[string]int64 `json:"sentiment_breakdown"`
	SourceBreakdown   map[string]int64   `json:"source_breakdown"`
	LanguageBreakdown map[string]int64   `json:"language_breakdown"`
	DailyStats        []DailyStat        `json:"daily_stats"`
}

type DailyStat struct {
	Date     string `json:"date"`
	Count    int64  `json:"count"`
	Positive int64  `json:"positive"`
	Negative int64  `json:"negative"`
	Neutral  int64  `json:"neutral"`
}

func (c *Comment) BeforeCreate() {
	c.ID = primitive.NewObjectID()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	c.IsProcessed = false
	c.HasSentiment = false
	
	if c.Language == "" {
		c.Language = "tr"
	}
}

func (c *Comment) BeforeUpdate() {
	c.UpdatedAt = time.Now()
}

func (c *Comment) IsValidSentiment() bool {
	if c.Sentiment == nil {
		return false
	}
	
	validLabels := []string{"POSITIVE", "NEGATIVE", "NEUTRAL"}
	for _, label := range validLabels {
		if c.Sentiment.Label == label {
			return true
		}
	}
	return false
}

func (c *Comment) GetSentimentColor() string {
	if c.Sentiment == nil {
		return "#gray-500"
	}
	
	switch c.Sentiment.Label {
	case "POSITIVE":
		return "#10B981"
	case "NEGATIVE":
		return "#EF4444"
	case "NEUTRAL":
		return "#6B7280"
	default:
		return "#6B7280"
	}
}