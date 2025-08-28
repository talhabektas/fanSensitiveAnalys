package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Slug        string             `json:"slug" bson:"slug" validate:"required,min=2,max=50"`
	League      string             `json:"league" bson:"league"`
	Country     string             `json:"country" bson:"country"`
	Logo        string             `json:"logo" bson:"logo"`
	Colors      []string           `json:"colors" bson:"colors"`
	Keywords    []string           `json:"keywords" bson:"keywords"`
	Subreddits  []string           `json:"subreddits" bson:"subreddits"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type TeamStats struct {
	TeamID           primitive.ObjectID `json:"team_id" bson:"team_id"`
	TeamName         string             `json:"team_name" bson:"team_name"`
	TotalComments    int64              `json:"total_comments" bson:"total_comments"`
	PositiveCount    int64              `json:"positive_count" bson:"positive_count"`
	NegativeCount    int64              `json:"negative_count" bson:"negative_count"`
	NeutralCount     int64              `json:"neutral_count" bson:"neutral_count"`
	AvgSentiment     float64            `json:"avg_sentiment" bson:"avg_sentiment"`
	LastCommentAt    time.Time          `json:"last_comment_at" bson:"last_comment_at"`
	SentimentTrend   string             `json:"sentiment_trend" bson:"sentiment_trend"`
}

type TeamCreateRequest struct {
	Name       string   `json:"name" validate:"required,min=2,max=100"`
	Slug       string   `json:"slug" validate:"required,min=2,max=50"`
	League     string   `json:"league"`
	Country    string   `json:"country"`
	Logo       string   `json:"logo"`
	Colors     []string `json:"colors"`
	Keywords   []string `json:"keywords" validate:"required,min=1"`
	Subreddits []string `json:"subreddits"`
}

type TeamUpdateRequest struct {
	Name       *string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	League     *string   `json:"league,omitempty"`
	Country    *string   `json:"country,omitempty"`
	Logo       *string   `json:"logo,omitempty"`
	Colors     *[]string `json:"colors,omitempty"`
	Keywords   *[]string `json:"keywords,omitempty"`
	Subreddits *[]string `json:"subreddits,omitempty"`
	IsActive   *bool     `json:"is_active,omitempty"`
}

func (t *Team) BeforeCreate() {
	t.ID = primitive.NewObjectID()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	if t.IsActive == false {
		t.IsActive = true
	}
}

func (t *Team) BeforeUpdate() {
	t.UpdatedAt = time.Now()
}

var TurkishTeams = []Team{
	{
		Name:       "Galatasaray",
		Slug:       "galatasaray",
		League:     "Süper Lig",
		Country:    "Turkey",
		Colors:     []string{"#FFA500", "#8B0000"},
		Keywords:   []string{"galatasaray", "gala", "gs", "aslan", "sarı-kırmızı"},
		Subreddits: []string{"galatasaray"},
		IsActive:   true,
	},
	{
		Name:       "Fenerbahçe",
		Slug:       "fenerbahce",
		League:     "Süper Lig",
		Country:    "Turkey",
		Colors:     []string{"#FFFF00", "#000080"},
		Keywords:   []string{"fenerbahçe", "fenerbahce", "fener", "fb", "kanarya", "sarı-lacivert"},
		Subreddits: []string{"fenerbahce"},
		IsActive:   true,
	},
	{
		Name:       "Beşiktaş",
		Slug:       "besiktas",
		League:     "Süper Lig",
		Country:    "Turkey",
		Colors:     []string{"#000000", "#FFFFFF"},
		Keywords:   []string{"beşiktaş", "besiktas", "bjk", "kartal", "siyah-beyaz"},
		Subreddits: []string{"besiktas"},
		IsActive:   true,
	},
	{
		Name:       "Trabzonspor",
		Slug:       "trabzonspor",
		League:     "Süper Lig",
		Country:    "Turkey",
		Colors:     []string{"#800080", "#000080"},
		Keywords:   []string{"trabzonspor", "trabzon", "ts", "bordo-mavi"},
		Subreddits: []string{"trabzonspor"},
		IsActive:   true,
	},
}