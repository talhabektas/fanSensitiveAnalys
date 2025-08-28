package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"
)

type ReportGeneratorService struct {
	trendService     *TrendService
	sentimentService *SentimentService
	commentService   *CommentService
}

type ExecutiveReport struct {
	GeneratedAt     string           `json:"generated_at"`
	Period          string           `json:"period"`
	Summary         ExecutiveSummary `json:"summary"`
	KeyFindings     []KeyFinding     `json:"key_findings"`
	TeamAnalysis    []TeamReport     `json:"team_analysis"`
	Recommendations []string         `json:"recommendations"`
	DataSources     []string         `json:"data_sources"`
}

type ExecutiveSummary struct {
	TotalComments      int     `json:"total_comments"`
	AnalyzedComments   int     `json:"analyzed_comments"`
	AverageDaily       float64 `json:"average_daily"`
	SentimentScore     string  `json:"sentiment_score"`
	MostActiveTeam     string  `json:"most_active_team"`
	TrendDirection     string  `json:"trend_direction"`
	DataCoverageScore  string  `json:"data_coverage_score"`
}

type KeyFinding struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"` // "high", "medium", "low"
	Category    string `json:"category"` // "positive", "negative", "neutral"
	Value       string `json:"value"`
}

type TeamReport struct {
	TeamName        string  `json:"team_name"`
	CommentCount    int     `json:"comment_count"`
	SentimentScore  float64 `json:"sentiment_score"`
	TrendDirection  string  `json:"trend_direction"`
	WeeklyChange    float64 `json:"weekly_change"`
	TopKeywords     []string `json:"top_keywords"`
	Summary         string  `json:"summary"`
}

func NewReportGeneratorService() *ReportGeneratorService {
	return &ReportGeneratorService{
		trendService:     NewTrendService(),
		sentimentService: NewSentimentService(),
		commentService:   NewCommentService(),
	}
}

// GenerateExecutiveReport - Yöneticiler için özet rapor
func (rgs *ReportGeneratorService) GenerateExecutiveReport(period string) (*ExecutiveReport, error) {
	// Trend analizi verilerini al
	trendAnalysis, err := rgs.trendService.GetTrendAnalysis(period)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend analysis: %v", err)
	}

	// Insights verilerini al
	insights, err := rgs.trendService.GenerateInsights(period)
	if err != nil {
		log.Printf("Failed to get insights: %v", err)
		insights = []InsightData{} // Boş slice ile devam et
	}

	report := &ExecutiveReport{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Period:      period,
		DataSources: []string{"Reddit", "YouTube", "Sosyal Medya Platformları"},
	}

	// Executive Summary oluştur
	report.Summary = rgs.createExecutiveSummary(trendAnalysis)

	// Key Findings oluştur
	report.KeyFindings = rgs.createKeyFindings(trendAnalysis, insights)

	// Team Analysis oluştur
	report.TeamAnalysis = rgs.createTeamAnalysis(trendAnalysis)

	// Recommendations oluştur
	report.Recommendations = rgs.generateRecommendations(trendAnalysis, insights)

	return report, nil
}

// createExecutiveSummary - Genel özet
func (rgs *ReportGeneratorService) createExecutiveSummary(analysis *TrendAnalysisResponse) ExecutiveSummary {
	totalComments := analysis.Summary.TotalComments
	averageDaily := analysis.Summary.AverageDaily

	// Genel trend yönünü belirle
	trendDirection := "kararlı"
	positiveChanges := 0
	negativeChanges := 0

	for _, team := range analysis.Teams {
		if team.Overall.WeeklyChange > 5 {
			positiveChanges++
		} else if team.Overall.WeeklyChange < -5 {
			negativeChanges++
		}
	}

	if positiveChanges > negativeChanges {
		trendDirection = "yükselen"
	} else if negativeChanges > positiveChanges {
		trendDirection = "düşen"
	}

	// En aktif takımı bul
	mostActiveTeam := ""
	maxComments := 0
	for _, team := range analysis.Teams {
		if team.Overall.TotalComments > maxComments {
			maxComments = team.Overall.TotalComments
			mostActiveTeam = team.TeamName
		}
	}

	// Data coverage score hesapla
	coverageScore := "İyi"
	if totalComments < 100 {
		coverageScore = "Düşük"
	} else if totalComments > 500 {
		coverageScore = "Mükemmel"
	}

	return ExecutiveSummary{
		TotalComments:     totalComments,
		AnalyzedComments:  totalComments, // Şimdilik aynı değer
		AverageDaily:      averageDaily,
		SentimentScore:    "Nötr", // Şimdilik sabit
		MostActiveTeam:    mostActiveTeam,
		TrendDirection:    trendDirection,
		DataCoverageScore: coverageScore,
	}
}

// createKeyFindings - Ana bulgular
func (rgs *ReportGeneratorService) createKeyFindings(analysis *TrendAnalysisResponse, insights []InsightData) []KeyFinding {
	var findings []KeyFinding

	// Volume bazlı bulgular
	if analysis.Summary.TotalComments > 200 {
		findings = append(findings, KeyFinding{
			Title:       "Yüksek Taraftar Aktivitesi",
			Description: fmt.Sprintf("%d adet taraftar yorumu analiz edildi. Bu, güçlü bir topluluk katılımını gösteriyor.", analysis.Summary.TotalComments),
			Impact:      "high",
			Category:    "positive",
			Value:       fmt.Sprintf("%d yorum", analysis.Summary.TotalComments),
		})
	}

	// Günlük ortalama bulgusu
	if analysis.Summary.AverageDaily > 30 {
		findings = append(findings, KeyFinding{
			Title:       "Sürekli Veri Akışı",
			Description: fmt.Sprintf("Günlük ortalama %.0f yorum ile sürekli veri toplanıyor.", analysis.Summary.AverageDaily),
			Impact:      "medium",
			Category:    "positive",
			Value:       fmt.Sprintf("%.0f/gün", analysis.Summary.AverageDaily),
		})
	}

	// En aktif takım bulgusu
	if analysis.Summary.MostPositiveTeam != "" {
		maxComments := 0
		for _, team := range analysis.Teams {
			if team.Overall.TotalComments > maxComments {
				maxComments = team.Overall.TotalComments
			}
		}

		findings = append(findings, KeyFinding{
			Title:       "En Aktif Taraftar Grubu",
			Description: fmt.Sprintf("%s taraftarları en yüksek aktiviteye sahip.", analysis.Summary.MostPositiveTeam),
			Impact:      "medium",
			Category:    "neutral",
			Value:       fmt.Sprintf("%d yorum", maxComments),
		})
	}

	// Multi-platform veri bulgusu
	findings = append(findings, KeyFinding{
		Title:       "Çok Platform Veri Toplama",
		Description: "Reddit ve YouTube platformlarından gerçek zamanlı veri toplama aktif durumda.",
		Impact:      "high",
		Category:    "positive",
		Value:       "2 platform",
	})

	// Insights'tan bulgular ekle
	for _, insight := range insights {
		if insight.Severity == "high" {
			category := "neutral"
			if insight.Type == "improvement" || insight.Type == "spike" {
				category = "positive"
			} else if insight.Type == "decline" {
				category = "negative"
			}

			findings = append(findings, KeyFinding{
				Title:       insight.TeamName + " - " + strings.Title(insight.Type),
				Description: insight.Description,
				Impact:      insight.Severity,
				Category:    category,
				Value:       insight.Value,
			})
		}
	}

	return findings
}

// createTeamAnalysis - Takım bazlı analiz
func (rgs *ReportGeneratorService) createTeamAnalysis(analysis *TrendAnalysisResponse) []TeamReport {
	var teamReports []TeamReport

	for _, team := range analysis.Teams {
		// Trend açıklaması
		trendDescription := "kararlı"
		if team.Overall.TrendDirection == "up" {
			trendDescription = "yükselen"
		} else if team.Overall.TrendDirection == "down" {
			trendDescription = "düşen"
		}

		// Özet oluştur
		summary := fmt.Sprintf("%s taraftarlarından %d yorum analiz edildi. ", 
			team.TeamName, team.Overall.TotalComments)
		
		if team.Overall.WeeklyChange > 10 {
			summary += fmt.Sprintf("Son hafta %%.1f artış gösterdi.", team.Overall.WeeklyChange)
		} else if team.Overall.WeeklyChange < -10 {
			summary += fmt.Sprintf("Son hafta %%.1f düşüş yaşandı.", -team.Overall.WeeklyChange)
		} else {
			summary += "Aktivite seviyesi kararlı kaldı."
		}

		teamReports = append(teamReports, TeamReport{
			TeamName:       team.TeamName,
			CommentCount:   team.Overall.TotalComments,
			SentimentScore: 0.0, // Şimdilik 0
			TrendDirection: trendDescription,
			WeeklyChange:   team.Overall.WeeklyChange,
			TopKeywords:    []string{"takım", "maç", "futbol"}, // Şimdilik statik
			Summary:        summary,
		})
	}

	return teamReports
}

// generateRecommendations - Öneriler
func (rgs *ReportGeneratorService) generateRecommendations(analysis *TrendAnalysisResponse, insights []InsightData) []string {
	var recommendations []string

	// Volume bazlı öneriler
	if analysis.Summary.TotalComments < 100 {
		recommendations = append(recommendations, 
			"Veri toplama kapsamını genişletmek için ek sosyal medya platformları değerlendirilebilir.")
	}

	if analysis.Summary.AverageDaily > 50 {
		recommendations = append(recommendations,
			"Yüksek veri akışı nedeniyle gerçek zamanlı monitoring sistemi kurulması önerilir.")
	}

	// Genel öneriler
	recommendations = append(recommendations, []string{
		"Sentiment analizi özelliği aktif hale getirilerek daha detaylı duygusal analiz yapılabilir.",
		"Takım bazlı karşılaştırmalı raporlar düzenli olarak yönetime sunulabilir.",
		"Kritik eğilimlerin tespit edilmesi için otomatik alert sistemi kurulabilir.",
		"Veri kalitesini artırmak için kaynak çeşitliliği genişletilebilir.",
		"Dashboard'da gerçek zamanlı veri görselleştirmesi ile operasyonel takip geliştirilebilir.",
	}...)

	return recommendations
}

// GenerateHTMLReport - HTML formatında rapor
func (rgs *ReportGeneratorService) GenerateHTMLReport(period string) (string, error) {
	report, err := rgs.GenerateExecutiveReport(period)
	if err != nil {
		return "", err
	}

	htmlTemplate := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Taraftar Sentiment Analizi - Executive Report</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .report { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 800px; margin: 0 auto; }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { color: #34495e; margin-top: 30px; }
        .summary { background: #ecf0f1; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .metric { display: inline-block; margin: 10px 20px 10px 0; }
        .metric-value { font-size: 24px; font-weight: bold; color: #2980b9; }
        .metric-label { font-size: 14px; color: #7f8c8d; }
        .finding { border-left: 4px solid #3498db; padding: 15px; margin: 10px 0; background: #f8f9fa; }
        .finding.positive { border-left-color: #27ae60; }
        .finding.negative { border-left-color: #e74c3c; }
        .team-analysis { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0; }
        .team-card { border: 1px solid #ddd; padding: 20px; border-radius: 8px; }
        .recommendations { background: #fff3cd; padding: 20px; border-radius: 8px; border: 1px solid #ffeaa7; }
        ul { line-height: 1.6; }
        .footer { text-align: center; margin-top: 40px; color: #7f8c8d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="report">
        <h1>🚀 Taraftar Sentiment Analizi - Executive Report</h1>
        
        <div class="summary">
            <h2>📊 Executive Summary ({{.Period}})</h2>
            <div class="metric">
                <div class="metric-value">{{.Summary.TotalComments}}</div>
                <div class="metric-label">Toplam Yorum</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{printf "%.0f" .Summary.AverageDaily}}</div>
                <div class="metric-label">Günlük Ortalama</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Summary.MostActiveTeam}}</div>
                <div class="metric-label">En Aktif Takım</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Summary.DataCoverageScore}}</div>
                <div class="metric-label">Veri Kapsamı</div>
            </div>
        </div>

        <h2>🔍 Key Findings</h2>
        {{range .KeyFindings}}
        <div class="finding {{.Category}}">
            <strong>{{.Title}}</strong> ({{.Value}})<br>
            {{.Description}}
        </div>
        {{end}}

        <h2>🏆 Team Analysis</h2>
        <div class="team-analysis">
            {{range .TeamAnalysis}}
            <div class="team-card">
                <h3>{{.TeamName}}</h3>
                <p><strong>Yorum Sayısı:</strong> {{.CommentCount}}</p>
                <p><strong>Trend:</strong> {{.TrendDirection}}</p>
                <p><strong>Haftalık Değişim:</strong> {{printf "%.1f" .WeeklyChange}}%</p>
                <p>{{.Summary}}</p>
            </div>
            {{end}}
        </div>

        <div class="recommendations">
            <h2>💡 Öneriler</h2>
            <ul>
                {{range .Recommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>

        <div class="footer">
            <p>Rapor Oluşturulma: {{.GeneratedAt}} | Veri Kaynakları: {{join .DataSources ", "}}</p>
            <p>🤖 Bu rapor otomatik olarak oluşturulmuştur - Taraftar Analizi Sistemi</p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}