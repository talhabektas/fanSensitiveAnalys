package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"taraftar-analizi/config"
	"taraftar-analizi/models"
)

type GroqService struct {
	client     *http.Client
	apiKey     string
	collection string
}

type GroqRequest struct {
	Messages    []GroqMessage `json:"messages"`
	Model       string        `json:"model"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int          `json:"max_tokens"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type GroqSentimentResult struct {
	Sentiment    string  `json:"sentiment"`
	Confidence   float64 `json:"confidence"`
	Category     string  `json:"category"`
	Keywords     []string `json:"keywords"`
	Explanation  string  `json:"explanation"`
}

type GroqAnalysisResult struct {
	Enhanced      *models.SentimentResult `json:"enhanced_sentiment"`
	Category      string                  `json:"category"`
	Keywords      []string               `json:"keywords"`
	ToxicityScore float64               `json:"toxicity_score"`
	Summary       string                `json:"summary"`
}

const (
	GroqAPIURL = "https://api.groq.com/openai/v1/chat/completions"
	GroqModel  = "llama-3.3-70b-versatile"
)

func NewGroqService() *GroqService {
	return &GroqService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: config.AppConfig.GroqAPIKey,
	}
}

// 1. Gelişmiş Duygu Analizi
func (gs *GroqService) EnhancedSentimentAnalysis(text string) (*GroqAnalysisResult, error) {
	if gs.apiKey == "" {
		return nil, fmt.Errorf("Groq API key not configured")
	}

	prompt := fmt.Sprintf(`Sen Türkiye futbol uzmanı bir AI'sın. Aşağıdaki taraftar yorumunu futbol kategorisine göre sınıflandır:

YORUM: "%s"

ÖZEL FUTBOL KATEGORİLERİ:
- TRANSFER: Transfer haberleri, oyuncu alımları, satımları
- KADRO: Kadro tercihleri, oyuncu değişiklikleri, ilk 11
- MAÇ_SONUCU: Maç sonrası yorumlar, sonuç değerlendirmeleri
- OYUNCU_PERFORMANSI: Oyuncu performansı, övgü, eleştiri
- TEKNİK_DİREKTÖR: Antrenör kararları, taktik, eleştiri
- HAKEM: Hakem kararları, VAR, penaltı tartışmaları
- TAKIM_PERFORMANSI: Genel takım oyunu, taktik, form
- LİG_DURUMU: Puan durumu, şampiyonluk yarışı, küme düşme
- DERBİ: Derbi maçlar, rakip takım yorumları
- TARAFTAR: Taraftar tepkileri, tribün olayları
- YÖNETİM: Kulüp yönetimi, başkan, mali durum
- GENEL: Diğer futbol konuları

JSON formatında ver:
{
  "sentiment": "POSITIVE/NEGATIVE/NEUTRAL",
  "confidence": 0.0-1.0 arası,
  "category": "Yukarıdaki kategorilerden biri",
  "keywords": ["anahtar", "kelimeler"],
  "toxicity_score": 0.0-1.0 arası,
  "explanation": "Kategori seçim sebebi"
}`, text)

	response, err := gs.callGroqAPI(prompt)
	if err != nil {
		return nil, fmt.Errorf("Groq API call failed: %w", err)
	}

	var result GroqSentimentResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("JSON parse error: %v, response: %s", err, response)
		return gs.fallbackAnalysis(text), nil
	}

	analysisResult := &GroqAnalysisResult{
		Enhanced: &models.SentimentResult{
			Label:       gs.normalizeSentiment(result.Sentiment),
			Score:       result.Confidence,
			Confidence:  result.Confidence,
			ModelUsed:   "groq-" + GroqModel,
			ProcessedAt: time.Now(),
		},
		Category:      result.Category,
		Keywords:      result.Keywords,
		ToxicityScore: gs.extractToxicityScore(response),
		Summary:       result.Explanation,
	}

	return analysisResult, nil
}

// 2. Yorum Kategorilendirme
func (gs *GroqService) CategorizeComment(text string) (string, []string, error) {
	if gs.apiKey == "" {
		return "Genel", []string{}, fmt.Errorf("Groq API key not configured")
	}

	prompt := fmt.Sprintf(`Bu futbol taraftarı yorumunu kategorilere ayır ve anahtar kelimeleri bul:

YORUM: "%s"

Kategoriler:
- Takım Performansı
- Oyuncu Eleştirisi  
- Hakem Kararları
- Transfer Haberleri
- Teknik Direktör
- Genel

JSON formatında ver:
{
  "category": "kategori_adı",
  "keywords": ["anahtar", "kelime", "listesi"]
}`, text)

	response, err := gs.callGroqAPI(prompt)
	if err != nil {
		return "Genel", []string{}, err
	}

	var result struct {
		Category string   `json:"category"`
		Keywords []string `json:"keywords"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("Categorization JSON parse error: %v", err)
		return "Genel", []string{}, nil
	}

	return result.Category, result.Keywords, nil
}

// 3. Akıllı Yorum Özetleme
func (gs *GroqService) SummarizeComments(comments []string) (string, error) {
	if gs.apiKey == "" {
		return "", fmt.Errorf("Groq API key not configured")
	}

	if len(comments) == 0 {
		return "Analiz edilecek yorum bulunamadı.", nil
	}

	// Çok fazla yorum varsa ilk 50'sini al
	if len(comments) > 50 {
		comments = comments[:50]
	}

	commentsText := strings.Join(comments, "\n---\n")
	
	prompt := fmt.Sprintf(`Bu futbol taraftarı yorumlarını analiz et ve akıllı bir özet çıkar:

YORUMLAR:
%s

Lütfen şu formatta özet ver:
- Ana konular neler?
- Genel duygu durumu nasıl?
- En çok tartışılan konular?
- Önemli trendler var mı?

2-3 paragraf halinde Türkçe özet yaz.`, commentsText)

	summary, err := gs.callGroqAPI(prompt)
	if err != nil {
		return "", fmt.Errorf("comment summarization failed: %w", err)
	}

	return summary, nil
}

// 5. Trend Analizi & İçgörüler
func (gs *GroqService) AnalyzeTrends(sentiments []models.Sentiment) (string, error) {
	if gs.apiKey == "" {
		return "", fmt.Errorf("Groq API key not configured")
	}

	if len(sentiments) == 0 {
		return "Analiz edilecek veri bulunamadı.", nil
	}

	// Sentiment verilerini organize et
	trendData := gs.prepareTrendData(sentiments)

	prompt := fmt.Sprintf(`Bu haftalık futbol taraftarı sentiment verilerini analiz et ve trendleri belirle:

VERİ:
%s

Lütfen şunları analiz et:
1. Bu hafta hangi konular trend oldu?
2. Taraftarların genel ruh hali nasıl değişti?  
3. En çok tartışılan konular neler?
4. Dikkat çekici trendler var mı?

Türkçe, anlaşılır bir trend raporu yaz (2-3 paragraf).`, trendData)

	trends, err := gs.callGroqAPI(prompt)
	if err != nil {
		return "", fmt.Errorf("trend analysis failed: %w", err)
	}

	return trends, nil
}

func (gs *GroqService) callGroqAPI(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request := GroqRequest{
		Messages: []GroqMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model:       GroqModel,
		Temperature: 0.3,
		MaxTokens:   1500,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GroqAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+gs.apiKey)

	resp, err := gs.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq API")
	}

	return response.Choices[0].Message.Content, nil
}

func (gs *GroqService) normalizeSentiment(sentiment string) string {
	sentiment = strings.ToUpper(strings.TrimSpace(sentiment))
	
	switch sentiment {
	case "POSITIVE", "POZITIF", "POS":
		return "POSITIVE"
	case "NEGATIVE", "NEGATIF", "NEG":
		return "NEGATIVE"
	case "NEUTRAL", "NÖTR", "NEU":
		return "NEUTRAL"
	default:
		return "NEUTRAL"
	}
}

func (gs *GroqService) extractToxicityScore(response string) float64 {
	// JSON'dan toxicity_score'u çıkarmaya çalış
	if strings.Contains(response, `"toxicity_score"`) {
		start := strings.Index(response, `"toxicity_score":`) + 17
		end := strings.IndexAny(response[start:], ",}")
		if end != -1 {
			scoreStr := response[start : start+end]
			scoreStr = strings.TrimSpace(scoreStr)
			if score, err := json.Number(scoreStr).Float64(); err == nil {
				return score
			}
		}
	}
	return 0.0
}

func (gs *GroqService) fallbackAnalysis(text string) *GroqAnalysisResult {
	// Basit fallback analizi
	sentiment := "NEUTRAL"
	confidence := 0.5
	
	text = strings.ToLower(text)
	if strings.Contains(text, "harika") || strings.Contains(text, "mükemmel") || strings.Contains(text, "süper") {
		sentiment = "POSITIVE"
		confidence = 0.7
	} else if strings.Contains(text, "kötü") || strings.Contains(text, "berbat") || strings.Contains(text, "rezil") {
		sentiment = "NEGATIVE"
		confidence = 0.7
	}

	return &GroqAnalysisResult{
		Enhanced: &models.SentimentResult{
			Label:       sentiment,
			Score:       confidence,
			Confidence:  confidence,
			ModelUsed:   "groq-fallback",
			ProcessedAt: time.Now(),
		},
		Category:      "Genel",
		Keywords:      []string{},
		ToxicityScore: 0.0,
		Summary:       "Basit analiz yapıldı",
	}
}

func (gs *GroqService) prepareTrendData(sentiments []models.Sentiment) string {
	positiveCount := 0
	negativeCount := 0
	neutralCount := 0
	total := len(sentiments)

	for _, sentiment := range sentiments {
		switch sentiment.Label {
		case "POSITIVE":
			positiveCount++
		case "NEGATIVE":
			negativeCount++
		default:
			neutralCount++
		}
	}

	return fmt.Sprintf(`
Toplam Analiz: %d yorum
Pozitif: %d (%%.1f)
Negatif: %d (%%.1f) 
Nötr: %d (%%.1f)
Son 7 gün trendi`,
		total,
		positiveCount, float64(positiveCount)*100/float64(total),
		negativeCount, float64(negativeCount)*100/float64(total),
		neutralCount, float64(neutralCount)*100/float64(total))
}

// Hibrit analiz: HuggingFace + Groq AI
func (gs *GroqService) HybridSentimentAnalysis(text string, hfResult *models.SentimentResult) (*models.SentimentResult, error) {
	groqResult, err := gs.EnhancedSentimentAnalysis(text)
	if err != nil {
		log.Printf("Groq analysis failed, using HuggingFace only: %v", err)
		return hfResult, nil
	}

	// İki sonucu karşılaştır ve en güvenilir olanı seç
	if groqResult.Enhanced.Confidence > hfResult.Confidence {
		groqResult.Enhanced.ModelUsed = "hybrid-groq-primary"
		return groqResult.Enhanced, nil
	} else {
		hfResult.ModelUsed = "hybrid-hf-primary"
		return hfResult, nil
	}
}