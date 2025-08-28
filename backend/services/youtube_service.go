package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type YouTubeService struct {
	apiKey         string
	commentService *CommentService
	teamService    *TeamService
}

type YouTubeSearchResponse struct {
	Items []YouTubeVideoItem `json:"items"`
}

type YouTubeVideoItem struct {
	ID      YouTubeVideoID `json:"id"`
	Snippet YouTubeSnippet `json:"snippet"`
}

type YouTubeVideoID struct {
	VideoID string `json:"videoId"`
}

type YouTubeSnippet struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ChannelTitle string `json:"channelTitle"`
	PublishedAt string `json:"publishedAt"`
}

type YouTubeCommentsResponse struct {
	Items []YouTubeCommentItem `json:"items"`
	NextPageToken string `json:"nextPageToken"`
}

type YouTubeCommentItem struct {
	Snippet YouTubeCommentSnippet `json:"snippet"`
}

type YouTubeCommentSnippet struct {
	TopLevelComment YouTubeComment `json:"topLevelComment"`
}

type YouTubeComment struct {
	Snippet YouTubeCommentDetail `json:"snippet"`
}

type YouTubeCommentDetail struct {
	TextDisplay    string `json:"textDisplay"`
	AuthorDisplayName string `json:"authorDisplayName"`
	PublishedAt    string `json:"publishedAt"`
	LikeCount      int    `json:"likeCount"`
}

func NewYouTubeService() *YouTubeService {
	return &YouTubeService{
		apiKey:         config.AppConfig.YouTubeAPIKey,
		commentService: NewCommentService(),
		teamService:    NewTeamService(),
	}
}

// Türk futbol kanallarından video arama - GÜNCEL İÇERİKLER İÇİN
func (ys *YouTubeService) SearchFootballVideos(query string, maxResults int) ([]YouTubeVideoItem, error) {
	if ys.apiKey == "" {
		return nil, fmt.Errorf("YouTube API key not configured")
	}

	// Güncel içerik için zaman filtreleri ekle - basketbol hariç
	searchQuery := fmt.Sprintf("%s süper lig türkiye -basketbol -basket", query)
	
	// Son 1 ay içindeki videolar (publishedAfter parametresi) - YouTube API Z formatı istiyor
	oneMonthAgo := time.Now().AddDate(0, -1, 0).UTC().Format("2006-01-02T15:04:05Z")
	
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&maxResults=%d&key=%s&relevanceLanguage=tr&regionCode=TR&publishedAfter=%s&order=date",
		url.QueryEscape(searchQuery), maxResults, ys.apiKey, oneMonthAgo)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call YouTube API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var searchResp YouTubeSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return searchResp.Items, nil
}

// Video yorumlarını çekme
func (ys *YouTubeService) GetVideoComments(videoID string, maxResults int) ([]YouTubeCommentDetail, error) {
	if ys.apiKey == "" {
		return nil, fmt.Errorf("YouTube API key not configured")
	}

	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/commentThreads?part=snippet&videoId=%s&maxResults=%d&order=time&key=%s",
		videoID, maxResults, ys.apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call YouTube API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var commentsResp YouTubeCommentsResponse
	if err := json.Unmarshal(body, &commentsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var comments []YouTubeCommentDetail
	for _, item := range commentsResp.Items {
		comments = append(comments, item.Snippet.TopLevelComment.Snippet)
	}

	return comments, nil
}

// Takım tespiti için - sadece futbol yorumları kabul et, basketbol içeriğini filtrele
func (ys *YouTubeService) detectTeamFromText(text string) *primitive.ObjectID {
	text = strings.ToLower(text)
	
	// BASKETBOL kelimelerini kontrol et ve hariç tut
	basketballKeywords := []string{
		"basket", "basketbol", "euroleague", "final four", "nba",
		"pivot", "guard", "üç sayı", "three point", "smash", "dunk",
		"rebounds", "ribaund", "assist", "asist", "blok", "block",
		"coach", "koç", "playoff", "play off", "euroleague",
	}
	
	for _, keyword := range basketballKeywords {
		if strings.Contains(text, keyword) {
			return nil // Basketbol içeriği varsa kabul etme
		}
	}
	
	// Futbol ile alakalı kelimeler var mı kontrol et
	footballKeywords := []string{
		"maç", "gol", "futbol", "lig", "takım", "oyuncu", "transfer",
		"şampiyonluk", "derbi", "sahada", "forvet", "kaleci", "defans",
		"orta saha", "pas", "şut", "penaltı", "korner", "kart", "hakem",
		"süper lig", "şampiyonlar ligi", "uefa", "fifa", "milli takım",
		"kadro", "forma", "saha", "stadyum", "taraftar", "tribün",
	}
	
	hasFootballContent := false
	for _, keyword := range footballKeywords {
		if strings.Contains(text, keyword) {
			hasFootballContent = true
			break
		}
	}
	
	if !hasFootballContent {
		return nil // Futbol ile alakalı değilse takım tespiti yapma
	}
	
	teams, err := ys.teamService.GetAllTeams()
	if err != nil {
		return nil
	}

	for _, team := range teams {
		teamName := strings.ToLower(team.Name)
		keywords := []string{
			teamName,
			strings.ReplaceAll(teamName, "ş", "s"),
			strings.ReplaceAll(teamName, "ç", "c"),
			strings.ReplaceAll(teamName, "ğ", "g"),
			strings.ReplaceAll(teamName, "ı", "i"),
			strings.ReplaceAll(teamName, "ö", "o"),
			strings.ReplaceAll(teamName, "ü", "u"),
		}

		// Kısaltmalar ekle
		switch teamName {
		case "galatasaray":
			keywords = append(keywords, "gs", "cimbom", "aslan")
		case "fenerbahçe":
			keywords = append(keywords, "fb", "fener", "kanaryalar")
		case "beşiktaş":
			keywords = append(keywords, "bjk", "beşik", "kartal", "siyah", "beyaz")
		case "trabzonspor":
			keywords = append(keywords, "ts", "trabzon", "bordo", "mavi")
		}

		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				return &team.ID
			}
		}
	}

	return nil
}

// YouTube yorumlarını database'e kaydetme  
func (ys *YouTubeService) SaveYouTubeComments(videoID, videoTitle string, comments []YouTubeCommentDetail) error {
	var savedCount, skippedCount int

	for _, comment := range comments {
		// Takım tespiti
		teamID := ys.detectTeamFromText(comment.TextDisplay + " " + videoTitle)
		if teamID == nil {
			skippedCount++
			continue // Takım tespit edilmezse kaydetme
		}

		// PublishedAt'i time.Time'a çevir (şimdilik kullanmıyoruz ama ileride lazım olabilir)
		_, err := time.Parse(time.RFC3339, comment.PublishedAt)
		if err != nil {
			_ = time.Now()
		}

		// Comment oluştur
		req := models.CommentCreateRequest{
			SourceID: fmt.Sprintf("yt_%d", time.Now().UnixNano()), // Unique ID
			Source:   "youtube",
			TeamID:   teamID.Hex(),
			Author:   comment.AuthorDisplayName,
			Text:     comment.TextDisplay,
			URL:      fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
			Score:    int64(comment.LikeCount),
			Language: "tr",
			Metadata: models.CommentMetadata{
				Platform:   "youtube",
				IsReply:    false,
				LikeCount:  int64(comment.LikeCount),
				Tags:       []string{"youtube", "futbol"},
			},
		}

		// Duplicate kontrolü - aynı metin ve author varsa kaydetme
		existingComment, _ := ys.commentService.GetCommentByTextAndAuthor(comment.TextDisplay, comment.AuthorDisplayName)
		if existingComment != nil {
			skippedCount++
			continue
		}

		// Kaydet
		savedComment, err := ys.commentService.CreateComment(req)
		if err != nil {
			log.Printf("Failed to save YouTube comment: %v", err)
			continue
		}

		log.Printf("Saved YouTube comment: %s by %s (Team: %s)", 
			savedComment.Text[:min(50, len(savedComment.Text))], 
			savedComment.Author, 
			teamID.Hex())
		
		savedCount++
	}

	log.Printf("YouTube comments processed: %d saved, %d skipped", savedCount, skippedCount)
	return nil
}

// Popüler Türk futbol videolarından yorum toplama
func (ys *YouTubeService) CollectTurkishFootballComments() error {
	log.Printf("Starting YouTube comment collection...")
	
	if ys.apiKey == "" {
		log.Printf("ERROR: YouTube API key is empty!")
		return fmt.Errorf("YouTube API key not configured")
	}
	
	log.Printf("YouTube API key is configured (length: %d)", len(ys.apiKey))
	
	// GÜNCEL FUTBOL HABERLERİ İÇİN ARAMAÇLAR - BASKETBOL İÇERİĞİ HARİÇ
	searchQueries := []string{
		// Güncel transferler - futbol odaklı
		"galatasaray futbol transfer haberleri",
		"fenerbahçe futbol yeni transferler", 
		"beşiktaş futbol transfer gündemi",
		"trabzonspor futbol son dakika transfer",
		
		// Teknik direktör haberleri - futbol
		"süper lig futbol teknik direktör",
		"galatasaray futbol hoca haberleri",
		"fenerbahçe futbol antrenör açıklamaları",
		"beşiktaş futbol teknik direktör röportaj",
		
		// Son maçlar ve güncel maç analizleri - futbol
		"süper lig futbol son maç",
		"galatasaray futbol son maç yorumları",
		"fenerbahçe futbol maç sonrası",
		"beşiktaş futbol puan durumu",
		"trabzonspor futbol form analizi",
		
		// Güncel gelişmeler - futbol
		"süper lig futbol puan durumu",
		"türk futbolu son dakika",
		"süper lig futbol haberleri bugün",
		"türkiye futbol ligi gündem",
	}

	log.Printf("Will search %d queries for YouTube videos", len(searchQueries))
	totalCollected := 0

	for _, query := range searchQueries {
		log.Printf("Searching YouTube for: %s", query)
		
		// Video ara
		videos, err := ys.SearchFootballVideos(query, 3) // Her aramada 3 video
		if err != nil {
			log.Printf("Failed to search videos for '%s': %v", query, err)
			continue
		}

		// Her video için yorumları çek
		for _, video := range videos {
			log.Printf("Collecting comments from: %s", video.Snippet.Title)
			
			comments, err := ys.GetVideoComments(video.ID.VideoID, 50) // Video başına 50 yorum
			if err != nil {
				log.Printf("Failed to get comments for video %s: %v", video.ID.VideoID, err)
				continue
			}

			// Yorumları kaydet
			err = ys.SaveYouTubeComments(video.ID.VideoID, video.Snippet.Title, comments)
			if err != nil {
				log.Printf("Failed to save comments: %v", err)
				continue
			}

			totalCollected += len(comments)
			
			// Rate limit için bekle
			time.Sleep(1 * time.Second)
		}

		// Aramalar arası bekle
		time.Sleep(2 * time.Second)
	}

	log.Printf("Total YouTube comments collected: %d", totalCollected)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}