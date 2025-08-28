package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"github.com/go-resty/resty/v2"
)

type RedditService struct {
	client       *resty.Client
	accessToken  string
	tokenExpiry  time.Time
}

type RedditAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type RedditPost struct {
	Data RedditPostData `json:"data"`
}

type RedditPostData struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Selftext    string  `json:"selftext"`
	Author      string  `json:"author"`
	Subreddit   string  `json:"subreddit"`
	Score       int64   `json:"score"`
	NumComments int64   `json:"num_comments"`
	Created     float64 `json:"created_utc"`
	URL         string  `json:"url"`
	Permalink   string  `json:"permalink"`
}

type RedditComment struct {
	Data RedditCommentData `json:"data"`
}

type RedditCommentData struct {
	ID        string  `json:"id"`
	Body      string  `json:"body"`
	Author    string  `json:"author"`
	Score     int64   `json:"score"`
	Created   float64 `json:"created_utc"`
	ParentID  string  `json:"parent_id"`
	LinkID    string  `json:"link_id"`
	Subreddit string  `json:"subreddit"`
	Replies   interface{} `json:"replies"`
}

type RedditListing struct {
	Data struct {
		Children []json.RawMessage `json:"children"`
		After    string            `json:"after"`
		Before   string            `json:"before"`
	} `json:"data"`
}

const (
	RedditOAuthURL = "https://www.reddit.com/api/v1/access_token"
	RedditAPIURL   = "https://oauth.reddit.com"
	UserAgent      = "TaraftarAnalizi/1.0 by /u/taraftar_bot"
)

func NewRedditService() *RedditService {
	client := resty.New().
		SetTimeout(30*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(2*time.Second).
		SetHeader("User-Agent", UserAgent)

	return &RedditService{
		client: client,
	}
}

func (rs *RedditService) Authenticate() error {
	if rs.accessToken != "" && time.Now().Before(rs.tokenExpiry) {
		return nil
	}

	data := url.Values{
		"grant_type": {"password"},
		"username":   {config.AppConfig.RedditUsername},
		"password":   {config.AppConfig.RedditPassword},
	}

	var authResp RedditAuthResponse
	resp, err := rs.client.R().
		SetBasicAuth(config.AppConfig.RedditClientID, config.AppConfig.RedditClientSecret).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(data.Encode()).
		SetResult(&authResp).
		Post(RedditOAuthURL)

	if err != nil {
		return fmt.Errorf("reddit authentication failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("reddit authentication failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	rs.accessToken = authResp.AccessToken
	rs.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn-300) * time.Second)

	log.Printf("Reddit authentication successful, token expires in %d seconds", authResp.ExpiresIn)
	return nil
}

func (rs *RedditService) GetSubredditPosts(subreddit string, limit int) ([]models.CommentCreateRequest, error) {
	if err := rs.Authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	url := fmt.Sprintf("%s/r/%s/hot.json?limit=%d", RedditAPIURL, subreddit, limit)

	var listing RedditListing
	resp, err := rs.client.R().
		SetAuthToken(rs.accessToken).
		SetResult(&listing).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts from r/%s: %w", subreddit, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("reddit API returned status %d for r/%s: %s", resp.StatusCode(), subreddit, resp.String())
	}

	var comments []models.CommentCreateRequest

	for _, child := range listing.Data.Children {
		var post RedditPost
		if err := json.Unmarshal(child, &post); err != nil {
			continue
		}

		if post.Data.Selftext != "" && post.Data.Selftext != "[removed]" && post.Data.Selftext != "[deleted]" {
			if rs.isValidTurkishText(post.Data.Selftext) {
				comment := models.CommentCreateRequest{
					SourceID:  post.Data.ID,
					Source:    "reddit",
					Author:    post.Data.Author,
					Text:      post.Data.Selftext,
					URL:       "https://reddit.com" + post.Data.Permalink,
					Score:     post.Data.Score,
					Subreddit: post.Data.Subreddit,
					Language:  "tr",
					Metadata: models.CommentMetadata{
						Platform:   "reddit",
						PostID:     post.Data.ID,
						IsReply:    false,
						ReplyCount: post.Data.NumComments,
						Quality:    rs.assessTextQuality(post.Data.Selftext),
						Tags:       []string{"post", "selftext"},
					},
				}
				comments = append(comments, comment)
			}
		}

		if post.Data.NumComments > 0 {
			postComments, err := rs.getPostComments(post.Data.Subreddit, post.Data.ID, 50)
			if err != nil {
				log.Printf("Failed to get comments for post %s: %v", post.Data.ID, err)
				continue
			}
			comments = append(comments, postComments...)
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Collected %d comments from r/%s", len(comments), subreddit)
	return comments, nil
}

func (rs *RedditService) getPostComments(subreddit, postID string, limit int) ([]models.CommentCreateRequest, error) {
	url := fmt.Sprintf("%s/r/%s/comments/%s.json?limit=%d&sort=top", RedditAPIURL, subreddit, postID, limit)

	var listings []RedditListing
	resp, err := rs.client.R().
		SetAuthToken(rs.accessToken).
		SetResult(&listings).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("reddit API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if len(listings) < 2 {
		return nil, fmt.Errorf("unexpected response format")
	}

	var comments []models.CommentCreateRequest
	rs.extractComments(listings[1].Data.Children, postID, subreddit, &comments)

	return comments, nil
}

func (rs *RedditService) extractComments(children []json.RawMessage, postID, subreddit string, comments *[]models.CommentCreateRequest) {
	for _, child := range children {
		var comment RedditComment
		if err := json.Unmarshal(child, &comment); err != nil {
			continue
		}

		if comment.Data.Body == "" || comment.Data.Body == "[removed]" || comment.Data.Body == "[deleted]" {
			continue
		}

		if comment.Data.Author == "AutoModerator" || comment.Data.Author == "[deleted]" {
			continue
		}

		if !rs.isValidTurkishText(comment.Data.Body) {
			continue
		}

		commentReq := models.CommentCreateRequest{
			SourceID:  comment.Data.ID,
			Source:    "reddit",
			Author:    comment.Data.Author,
			Text:      comment.Data.Body,
			URL:       fmt.Sprintf("https://reddit.com/r/%s/comments/%s/_/%s", subreddit, postID, comment.Data.ID),
			Score:     comment.Data.Score,
			ParentID:  comment.Data.ParentID,
			Subreddit: comment.Data.Subreddit,
			Language:  "tr",
			Metadata: models.CommentMetadata{
				Platform:    "reddit",
				PostID:      postID,
				ThreadID:    comment.Data.LinkID,
				IsReply:     comment.Data.ParentID != comment.Data.LinkID,
				Quality:     rs.assessTextQuality(comment.Data.Body),
				Tags:        []string{"comment"},
				ProcessedBy: "reddit_service",
			},
		}

		*comments = append(*comments, commentReq)

		if comment.Data.Replies != nil {
			if repliesMap, ok := comment.Data.Replies.(map[string]interface{}); ok {
				if dataMap, ok := repliesMap["data"].(map[string]interface{}); ok {
					if childrenArray, ok := dataMap["children"].([]interface{}); ok {
						var repliesChildren []json.RawMessage
						for _, child := range childrenArray {
							if childBytes, err := json.Marshal(child); err == nil {
								repliesChildren = append(repliesChildren, childBytes)
							}
						}
						rs.extractComments(repliesChildren, postID, subreddit, comments)
					}
				}
			}
		}
	}
}

func (rs *RedditService) CollectFromAllSubreddits() ([]models.CommentCreateRequest, error) {
	turkishFootballSubreddits := []string{
		"galatasaray",
		"fenerbahce", 
		"besiktas",
		"trabzonspor",
		"superlig",
		"turkishfootball",
		"SuperLigTurkey",
	}

	var allComments []models.CommentCreateRequest

	for _, subreddit := range turkishFootballSubreddits {
		log.Printf("Collecting from r/%s...", subreddit)
		
		comments, err := rs.GetSubredditPosts(subreddit, 25)
		if err != nil {
			log.Printf("Error collecting from r/%s: %v", subreddit, err)
			continue
		}

		allComments = append(allComments, comments...)
		
		time.Sleep(2 * time.Second)
	}

	log.Printf("Total collected comments: %d", len(allComments))
	return allComments, nil
}

func (rs *RedditService) isValidTurkishText(text string) bool {
	if len(text) < 10 {
		return false
	}

	if len(text) > 5000 {
		return false
	}

	text = strings.ToLower(text)

	turkishWords := []string{
		"bu", "bir", "ve", "de", "da", "için", "ile", "var", "yok", "çok", "şu", "o", "ben", "sen",
		"galatasaray", "fenerbahçe", "beşiktaş", "trabzonspor", "futbol", "maç", "takım", "oyuncu",
		"gol", "lig", "şampiyonluk", "derbi", "hakem", "saha", "tribün", "taraftar", "forvet", "kaleci",
		"müdür", "hoca", "antrenör", "transfer", "sezon", "puan", "galibiyet", "mağlubiyet",
	}

	turkishChars := regexp.MustCompile(`[çğıöşüÇĞIİÖŞÜ]`)
	if turkishChars.MatchString(text) {
		return true
	}

	wordCount := 0
	for _, word := range turkishWords {
		if strings.Contains(text, word) {
			wordCount++
		}
	}

	return wordCount >= 2
}

func (rs *RedditService) assessTextQuality(text string) string {
	text = strings.TrimSpace(text)
	
	if len(text) < 10 {
		return "low"
	}
	
	if len(text) > 500 {
		return "high"
	}
	
	if len(text) > 100 {
		return "medium"
	}
	
	uppercaseRatio := float64(len(regexp.MustCompile(`[A-ZÇĞIİÖŞÜ]`).FindAllString(text, -1))) / float64(len(text))
	if uppercaseRatio > 0.5 {
		return "low"
	}
	
	spaceCount := strings.Count(text, " ")
	if spaceCount < 3 {
		return "low"
	}
	
	return "medium"
}