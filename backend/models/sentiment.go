package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sentiment struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CommentID   primitive.ObjectID `json:"comment_id" bson:"comment_id" validate:"required"`
	TeamID      primitive.ObjectID `json:"team_id" bson:"team_id"`
	Label       string             `json:"label" bson:"label" validate:"required,oneof=POSITIVE NEGATIVE NEUTRAL"`
	Score       float64            `json:"score" bson:"score" validate:"required,gte=0,lte=1"`
	Confidence  float64            `json:"confidence" bson:"confidence" validate:"required,gte=0,lte=1"`
	ModelUsed   string             `json:"model_used" bson:"model_used"`
	AnalysisDetails AnalysisDetails `json:"analysis_details" bson:"analysis_details"`
	Metadata    SentimentMetadata  `json:"metadata" bson:"metadata"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type AnalysisDetails struct {
	Scores         map[string]float64 `json:"scores" bson:"scores"`
	Keywords       []string           `json:"keywords" bson:"keywords"`
	Emotions       map[string]float64 `json:"emotions,omitempty" bson:"emotions,omitempty"`
	Topics         []string           `json:"topics,omitempty" bson:"topics,omitempty"`
	Language       string             `json:"language" bson:"language"`
	TextLength     int                `json:"text_length" bson:"text_length"`
	ProcessingTime float64            `json:"processing_time" bson:"processing_time"`
	Category       string             `json:"category,omitempty" bson:"category,omitempty"`
	ToxicityScore  float64           `json:"toxicity_score,omitempty" bson:"toxicity_score,omitempty"`
	GroqSummary    string             `json:"groq_summary,omitempty" bson:"groq_summary,omitempty"`
}

type SentimentMetadata struct {
	ProcessedBy   string            `json:"processed_by" bson:"processed_by"`
	ProcessedAt   time.Time         `json:"processed_at" bson:"processed_at"`
	APIVersion    string            `json:"api_version" bson:"api_version"`
	RetryCount    int               `json:"retry_count" bson:"retry_count"`
	CustomFields  map[string]string `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
}

type SentimentAnalysisRequest struct {
	Text     string `json:"text" validate:"required,min=5,max=5000"`
	Language string `json:"language"`
	TeamID   string `json:"team_id,omitempty"`
}

type SentimentReport struct {
	TeamID           primitive.ObjectID       `json:"team_id" bson:"team_id"`
	TeamName         string                   `json:"team_name" bson:"team_name"`
	Period           ReportPeriod             `json:"period" bson:"period"`
	TotalAnalyzed    int64                    `json:"total_analyzed" bson:"total_analyzed"`
	SentimentCounts  map[string]int64         `json:"sentiment_counts" bson:"sentiment_counts"`
	AverageSentiment float64                  `json:"average_sentiment" bson:"average_sentiment"`
	TrendAnalysis    TrendAnalysis            `json:"trend_analysis" bson:"trend_analysis"`
	TopKeywords      []KeywordAnalysis        `json:"top_keywords" bson:"top_keywords"`
	HourlyDistribution map[int]SentimentHourly `json:"hourly_distribution" bson:"hourly_distribution"`
	SourceBreakdown  map[string]SentimentSourceStats `json:"source_breakdown" bson:"source_breakdown"`
	GeneratedAt      time.Time                `json:"generated_at" bson:"generated_at"`
}

type ReportPeriod struct {
	StartDate time.Time `json:"start_date" bson:"start_date"`
	EndDate   time.Time `json:"end_date" bson:"end_date"`
	Label     string    `json:"label" bson:"label"`
}

type TrendAnalysis struct {
	Direction     string    `json:"direction" bson:"direction"`
	ChangePercent float64   `json:"change_percent" bson:"change_percent"`
	PeakDate      time.Time `json:"peak_date" bson:"peak_date"`
	LowestDate    time.Time `json:"lowest_date" bson:"lowest_date"`
	Volatility    float64   `json:"volatility" bson:"volatility"`
}

type KeywordAnalysis struct {
	Keyword        string  `json:"keyword" bson:"keyword"`
	Count          int64   `json:"count" bson:"count"`
	AvgSentiment   float64 `json:"avg_sentiment" bson:"avg_sentiment"`
	SentimentRange string  `json:"sentiment_range" bson:"sentiment_range"`
}

type SentimentHourly struct {
	Hour     int     `json:"hour" bson:"hour"`
	Count    int64   `json:"count" bson:"count"`
	Positive int64   `json:"positive" bson:"positive"`
	Negative int64   `json:"negative" bson:"negative"`
	Neutral  int64   `json:"neutral" bson:"neutral"`
	AvgScore float64 `json:"avg_score" bson:"avg_score"`
}

type SentimentSourceStats struct {
	Source       string  `json:"source" bson:"source"`
	Count        int64   `json:"count" bson:"count"`
	Positive     int64   `json:"positive" bson:"positive"`
	Negative     int64   `json:"negative" bson:"negative"`
	Neutral      int64   `json:"neutral" bson:"neutral"`
	AvgSentiment float64 `json:"avg_sentiment" bson:"avg_sentiment"`
}

type SentimentStats struct {
	TotalComments      int64                         `json:"total_comments"`      // Toplam yorum sayısı
	TotalAnalyzed      int64                         `json:"total_analyzed"`      // Analiz edilmiş yorum sayısı
	TeamComments       map[string]int64              `json:"team_comments"`       // Takım bazında yorum sayıları
	OverallSentiment   float64                       `json:"overall_sentiment"`
	SentimentBreakdown map[string]int64              `json:"sentiment_breakdown"`
	ConfidenceStats    ConfidenceStats               `json:"confidence_stats"`
	TeamComparison     []TeamSentimentComparison     `json:"team_comparison"`
	RecentTrends       []SentimentTrend              `json:"recent_trends"`
	ModelPerformance   map[string]ModelPerformance   `json:"model_performance"`
}

type ConfidenceStats struct {
	AverageConfidence float64 `json:"average_confidence"`
	HighConfidence    int64   `json:"high_confidence"`
	MediumConfidence  int64   `json:"medium_confidence"`
	LowConfidence     int64   `json:"low_confidence"`
}

type TeamSentimentComparison struct {
	TeamID       primitive.ObjectID `json:"team_id"`
	TeamName     string             `json:"team_name"`
	AvgSentiment float64            `json:"avg_sentiment"`
	TotalComments int64             `json:"total_comments"`
	Ranking      int                `json:"ranking"`
}

type SentimentTrend struct {
	Date     string  `json:"date"`
	Positive int64   `json:"positive"`
	Negative int64   `json:"negative"`
	Neutral  int64   `json:"neutral"`
	Score    float64 `json:"score"`
}

type ModelPerformance struct {
	Model           string  `json:"model"`
	TotalProcessed  int64   `json:"total_processed"`
	AverageTime     float64 `json:"average_time"`
	SuccessRate     float64 `json:"success_rate"`
	AverageConfidence float64 `json:"average_confidence"`
}

func (s *Sentiment) BeforeCreate() {
	s.ID = primitive.NewObjectID()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Metadata.ProcessedAt = time.Now()
}

func (s *Sentiment) BeforeUpdate() {
	s.UpdatedAt = time.Now()
}

func (s *Sentiment) GetSentimentValue() float64 {
	switch s.Label {
	case "POSITIVE":
		return s.Score
	case "NEGATIVE":
		return -s.Score
	case "NEUTRAL":
		return 0
	default:
		return 0
	}
}

func (s *Sentiment) IsHighConfidence() bool {
	return s.Confidence >= 0.8
}

func (s *Sentiment) GetConfidenceLevel() string {
	if s.Confidence >= 0.8 {
		return "high"
	} else if s.Confidence >= 0.6 {
		return "medium"
	}
	return "low"
}

type CleanupResult struct {
	TotalDuplicatesFound int64     `json:"total_duplicates_found"`
	DuplicatesRemoved    int64     `json:"duplicates_removed"`
	Errors               []string  `json:"errors,omitempty"`
	CompletedAt          time.Time `json:"completed_at"`
}

// Yeni Grok AI özellikleri için modeller
type CommentSummary struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TeamID      primitive.ObjectID `json:"team_id" bson:"team_id"`
	Period      ReportPeriod       `json:"period" bson:"period"`
	Summary     string             `json:"summary" bson:"summary"`
	TotalComments int              `json:"total_comments" bson:"total_comments"`
	MainTopics  []string           `json:"main_topics" bson:"main_topics"`
	GeneratedBy string             `json:"generated_by" bson:"generated_by"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type TrendInsight struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TeamID      primitive.ObjectID `json:"team_id,omitempty" bson:"team_id,omitempty"`
	Period      ReportPeriod       `json:"period" bson:"period"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	TrendType   string             `json:"trend_type" bson:"trend_type"` // "positive", "negative", "topic"
	Confidence  float64            `json:"confidence" bson:"confidence"`
	Keywords    []string           `json:"keywords" bson:"keywords"`
	GeneratedBy string             `json:"generated_by" bson:"generated_by"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type CategoryStats struct {
	Category    string  `json:"category"`
	Count       int64   `json:"count"`
	Percentage  float64 `json:"percentage"`
	AvgSentiment float64 `json:"avg_sentiment"`
	Keywords    []string `json:"keywords,omitempty"`
}

type EnhancedSentimentStats struct {
	TotalAnalyzed       int64                         `json:"total_analyzed"`
	OverallSentiment    float64                       `json:"overall_sentiment"`
	SentimentBreakdown  map[string]int64              `json:"sentiment_breakdown"`
	CategoryBreakdown   []CategoryStats               `json:"category_breakdown"`
	ToxicityStats       ToxicityStats                 `json:"toxicity_stats"`
	ConfidenceStats     ConfidenceStats               `json:"confidence_stats"`
	TeamComparison      []TeamSentimentComparison     `json:"team_comparison"`
	RecentTrends        []SentimentTrend              `json:"recent_trends"`
	ModelPerformance    map[string]ModelPerformance   `json:"model_performance"`
	LatestSummary       string                        `json:"latest_summary,omitempty"`
	TrendInsights       []TrendInsight               `json:"trend_insights,omitempty"`
}

type ToxicityStats struct {
	TotalScanned    int64   `json:"total_scanned"`
	HighToxicity    int64   `json:"high_toxicity"`    // >0.7
	MediumToxicity  int64   `json:"medium_toxicity"`  // 0.3-0.7
	LowToxicity     int64   `json:"low_toxicity"`     // <0.3
	AverageToxicity float64 `json:"average_toxicity"`
}

var SentimentColorMapping = map[string]string{
	"POSITIVE": "#10B981",
	"NEGATIVE": "#EF4444",
	"NEUTRAL":  "#6B7280",
}

var CategoryColorMapping = map[string]string{
	"Takım Performansı": "#3B82F6",
	"Oyuncu Eleştirisi": "#EF4444", 
	"Hakem Kararları":   "#F59E0B",
	"Transfer Haberleri": "#10B981",
	"Teknik Direktör":   "#8B5CF6",
	"Genel":             "#6B7280",
}