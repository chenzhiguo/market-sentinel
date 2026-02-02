package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type Reporter struct {
	cfg   *config.Config
	store *storage.Storage
}

func New(cfg *config.Config, store *storage.Storage) *Reporter {
	return &Reporter{
		cfg:   cfg,
		store: store,
	}
}

type ReportData struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	GeneratedAt time.Time         `json:"generated_at"`
	Period      Period            `json:"period"`
	Summary     string            `json:"summary"`
	MarketMood  MarketMood        `json:"market_mood"`
	Highlights  []HighlightItem   `json:"highlights"`
	StockSummary []StockSummary   `json:"stock_summary"`
}

type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type MarketMood struct {
	Overall    string  `json:"overall"`
	Confidence float64 `json:"confidence"`
	Bullish    int     `json:"bullish_count"`
	Bearish    int     `json:"bearish_count"`
	Neutral    int     `json:"neutral_count"`
}

type HighlightItem struct {
	NewsID    string                `json:"news_id"`
	Source    string                `json:"source"`
	Author    string                `json:"author"`
	Content   string                `json:"content"`
	Sentiment string                `json:"sentiment"`
	Impact    string                `json:"impact"`
	Summary   string                `json:"summary"`
	Stocks    []storage.StockImpact `json:"stocks"`
}

type StockSummary struct {
	Symbol       string  `json:"symbol"`
	MentionCount int     `json:"mention_count"`
	AvgScore     float64 `json:"avg_score"`
	Sentiment    string  `json:"sentiment"`
}

func (r *Reporter) GenerateMorningBrief(ctx context.Context) (*ReportData, error) {
	// Get analyses from the last 12 hours (overnight)
	since := time.Now().Add(-12 * time.Hour)
	return r.generateReport(ctx, "morning_brief", "美股盘前简报", since, time.Now())
}

func (r *Reporter) GenerateDailySummary(ctx context.Context) (*ReportData, error) {
	// Get analyses from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	return r.generateReport(ctx, "daily_summary", "每日舆情汇总", since, time.Now())
}

func (r *Reporter) GenerateCustomReport(ctx context.Context, reportType, title string, since, until time.Time) (*ReportData, error) {
	return r.generateReport(ctx, reportType, title, since, until)
}

func (r *Reporter) generateReport(ctx context.Context, reportType, title string, since, until time.Time) (*ReportData, error) {
	// Fetch analyses for the period
	analyses, _, err := r.store.ListAnalysis(since, until, "", 500, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analyses: %w", err)
	}

	// Calculate market mood
	mood := calculateMarketMood(analyses)

	// Get highlights (high impact items)
	var highlights []HighlightItem
	stockScores := make(map[string][]int)

	for _, a := range analyses {
		// Fetch original news
		news, _ := r.store.GetNews(a.NewsID)
		
		if a.ImpactLevel == "high" || a.ImpactLevel == "medium" {
			item := HighlightItem{
				NewsID:    a.NewsID,
				Sentiment: a.Sentiment,
				Impact:    a.ImpactLevel,
				Summary:   a.Summary,
				Stocks:    a.StockDetails,
			}
			if news != nil {
				item.Source = news.Source
				item.Author = news.Author
				item.Content = truncate(news.Content, 200)
			}
			highlights = append(highlights, item)
		}

		// Aggregate stock scores
		for _, stock := range a.StockDetails {
			stockScores[stock.Symbol] = append(stockScores[stock.Symbol], stock.Score)
		}
	}

	// Build stock summary
	var stockSummary []StockSummary
	for symbol, scores := range stockScores {
		if len(scores) == 0 {
			continue
		}
		sum := 0
		for _, s := range scores {
			sum += s
		}
		avg := float64(sum) / float64(len(scores))
		sentiment := "neutral"
		if avg > 2 {
			sentiment = "bullish"
		} else if avg < -2 {
			sentiment = "bearish"
		}
		stockSummary = append(stockSummary, StockSummary{
			Symbol:       symbol,
			MentionCount: len(scores),
			AvgScore:     avg,
			Sentiment:    sentiment,
		})
	}

	report := &ReportData{
		ID:          fmt.Sprintf("rpt_%s_%d", reportType, time.Now().Unix()),
		Type:        reportType,
		Title:       title,
		GeneratedAt: time.Now(),
		Period: Period{
			Start: since,
			End:   until,
		},
		Summary:      generateSummaryText(mood, len(analyses), len(highlights)),
		MarketMood:   mood,
		Highlights:   highlights,
		StockSummary: stockSummary,
	}

	// Save to database
	if err := r.saveReport(report); err != nil {
		return nil, err
	}

	// Save to file if configured
	if r.cfg.Reporter.SaveToFile {
		if err := r.saveToFile(report); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to save report to file: %v\n", err)
		}
	}

	return report, nil
}

func (r *Reporter) saveReport(report *ReportData) error {
	contentJSON, _ := json.Marshal(report)
	dbReport := &storage.Report{
		ID:          report.ID,
		Type:        report.Type,
		Title:       report.Title,
		Summary:     report.Summary,
		Content:     string(contentJSON),
		GeneratedAt: report.GeneratedAt,
	}
	return r.store.SaveReport(dbReport)
}

func (r *Reporter) saveToFile(report *ReportData) error {
	dir := r.cfg.Storage.ReportsDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.json", report.Type, report.GeneratedAt.Format("2006-01-02_150405"))
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func calculateMarketMood(analyses []storage.Analysis) MarketMood {
	var bullish, bearish, neutral int
	var totalConfidence float64

	for _, a := range analyses {
		totalConfidence += a.Confidence
		switch a.Sentiment {
		case "positive":
			bullish++
		case "negative":
			bearish++
		default:
			neutral++
		}
	}

	overall := "neutral"
	if bullish > bearish*2 {
		overall = "bullish"
	} else if bearish > bullish*2 {
		overall = "bearish"
	} else if bullish > bearish {
		overall = "slightly_bullish"
	} else if bearish > bullish {
		overall = "slightly_bearish"
	}

	avgConfidence := 0.5
	if len(analyses) > 0 {
		avgConfidence = totalConfidence / float64(len(analyses))
	}

	return MarketMood{
		Overall:    overall,
		Confidence: avgConfidence,
		Bullish:    bullish,
		Bearish:    bearish,
		Neutral:    neutral,
	}
}

func generateSummaryText(mood MarketMood, total, highlights int) string {
	moodText := map[string]string{
		"bullish":          "市场情绪积极乐观",
		"slightly_bullish": "市场情绪偏向乐观",
		"neutral":          "市场情绪中性",
		"slightly_bearish": "市场情绪偏向谨慎",
		"bearish":          "市场情绪较为悲观",
	}

	return fmt.Sprintf("共分析 %d 条信息，其中 %d 条重点关注。%s，看多 %d 条，看空 %d 条，中性 %d 条。",
		total, highlights, moodText[mood.Overall], mood.Bullish, mood.Bearish, mood.Neutral)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
