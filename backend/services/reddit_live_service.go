package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"taraftar-analizi/models"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type RedditLiveService struct {
	client         *reddit.Client
	commentService *CommentService
	isRunning      bool
	stopChannel    chan bool
}

func NewRedditLiveService() *RedditLiveService {
	// Read-only client oluştur (authentication gerektirmez)
	redditClient, err := reddit.NewReadonlyClient()
	if err != nil {
		log.Printf("Failed to create Reddit client: %v", err)
		return nil
	}

	return &RedditLiveService{
		client:         redditClient,
		commentService: NewCommentService(),
		isRunning:      false,
		stopChannel:    make(chan bool),
	}
}

type LiveStreamConfig struct {
	Subreddits []string          `json:"subreddits"`
	Teams      map[string]string `json:"teams"` // team_name -> team_id mapping
	Interval   time.Duration     `json:"interval"`
	Limit      int               `json:"limit"`
}

func (rls *RedditLiveService) StartLiveStream(config LiveStreamConfig) error {
	if rls.isRunning {
		return fmt.Errorf("live stream is already running")
	}

	if rls.client == nil {
		return fmt.Errorf("reddit client not initialized")
	}

	rls.isRunning = true
	log.Printf("Starting Reddit live stream for subreddits: %v", config.Subreddits)

	go rls.streamLoop(config)
	return nil
}

func (rls *RedditLiveService) StopLiveStream() {
	if !rls.isRunning {
		return
	}

	log.Println("Stopping Reddit live stream...")
	rls.stopChannel <- true
	rls.isRunning = false
}

func (rls *RedditLiveService) IsRunning() bool {
	return rls.isRunning
}

func (rls *RedditLiveService) streamLoop(config LiveStreamConfig) {
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-rls.stopChannel:
			log.Println("Reddit live stream stopped")
			return
		case <-ticker.C:
			rls.collectLatestPosts(config)
		}
	}
}

func (rls *RedditLiveService) collectLatestPosts(config LiveStreamConfig) {
	for _, subreddit := range config.Subreddits {
		posts, err := rls.getLatestPosts(subreddit, config.Limit)
		if err != nil {
			log.Printf("Error collecting from r/%s: %v", subreddit, err)
			continue
		}

		newComments := 0
		for _, post := range posts {
			// Takım tespiti
			postText := post.Title
			if post.Body != "" {
				postText += " " + post.Body
			}
			teamID := rls.detectTeam(postText, config.Teams)
			if teamID == "" {
				continue // Takımla ilgili değilse atla
			}

			// Comment modeline çevir
			comment := models.CommentCreateRequest{
				SourceID:  post.ID,
				Source:    "reddit",
				TeamID:    teamID,
				Author:    post.Author,
				Text:      rls.cleanText(postText),
				URL:       "https://reddit.com" + post.Permalink,
				Score:     int64(post.Score),
				Subreddit: subreddit,
				Language:  "tr",
			}

			// Veritabanına kaydet
			_, err := rls.commentService.CreateComment(comment)
			if err != nil {
				if err.Error() != "comment already exists" {
					log.Printf("Error saving comment: %v", err)
				}
			} else {
				newComments++
				log.Printf("New comment from r/%s: %s (Team: %s)", subreddit, post.Author, teamID)
			}
		}

		if newComments > 0 {
			log.Printf("Collected %d new comments from r/%s", newComments, subreddit)
		}
	}
}

func (rls *RedditLiveService) getLatestPosts(subreddit string, limit int) ([]*reddit.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	posts, _, err := rls.client.Subreddit.NewPosts(ctx, subreddit, &reddit.ListOptions{
		Limit: limit,
	})

	if err != nil {
		return nil, err
	}

	// Son 1 saat içindeki postları filtrele
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var recentPosts []*reddit.Post

	for _, post := range posts {
		if post.Created.After(oneHourAgo) {
			recentPosts = append(recentPosts, post)
		}
	}

	return recentPosts, nil
}

func (rls *RedditLiveService) detectTeam(text string, teams map[string]string) string {
	text = strings.ToLower(text)
	
	// Türkçe karakterleri normalize et
	text = strings.ReplaceAll(text, "ç", "c")
	text = strings.ReplaceAll(text, "ğ", "g")
	text = strings.ReplaceAll(text, "ı", "i")
	text = strings.ReplaceAll(text, "ö", "o")
	text = strings.ReplaceAll(text, "ş", "s")
	text = strings.ReplaceAll(text, "ü", "u")

	// Takım keywords'leri kontrol et
	teamKeywords := map[string][]string{
		"galatasaray": {"galatasaray", "gs", "cimbom", "aslan", "sari kirmizi"},
		"fenerbahce":  {"fenerbahce", "fb", "fener", "kanarya", "sari lacivert"},
		"besiktas":    {"besiktas", "bjk", "kartal", "siyah beyaz"},
		"trabzonspor": {"trabzonspor", "ts", "firtina", "bordo mavi"},
	}

	for teamName, keywords := range teamKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				if teamID, exists := teams[teamName]; exists {
					return teamID
				}
			}
		}
	}

	return ""
}

func (rls *RedditLiveService) cleanText(text string) string {
	// HTML karakterlerini temizle
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	
	// Fazla boşlukları temizle
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	
	return strings.TrimSpace(text)
}

func (rls *RedditLiveService) GetStreamStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running":  rls.isRunning,
		"started_at":  time.Now().Format(time.RFC3339),
		"client_ready": rls.client != nil,
	}
}