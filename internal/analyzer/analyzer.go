package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/llm"
	
	// Register providers via init()
	_ "github.com/chenzhiguo/market-sentinel/internal/llm/anthropic"
	_ "github.com/chenzhiguo/market-sentinel/internal/llm/ollama"
	
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type Analyzer struct {
	cfg      *config.Config
	store    *storage.Storage
	provider llm.Provider
	mapper   *StockMapper // Added StockMapper
}

func New(cfg *config.Config, store *storage.Storage) *Analyzer {
	providerName := strings.ToLower(cfg.Analyzer.LLMProvider)
	
	providerConfig := map[string]string{
		"model":   cfg.Analyzer.LLMModel,
		"api_key": cfg.Analyzer.APIKey,
		"url":     cfg.Analyzer.OllamaURL,
	}

	provider, err := llm.NewProvider(providerName, providerConfig)
	if err != nil {
		log.Printf("Failed to create provider %s: %v", providerName, err)
		if provider == nil {
			log.Fatalf("Critical error: Could not initialize LLM provider '%s'. check config.", providerName)
		}
	}

	return &Analyzer{
		cfg:      cfg,
		store:    store,
		provider: provider,
		mapper:   NewStockMapper(), // Initialize StockMapper
	}
}

type AnalysisResult struct {
	Sentiment   string        `json:"sentiment"`
	Impact      string        `json:"impact"`
	Summary     string        `json:"summary"`
	Stocks      []StockResult `json:"stocks"`
	Confidence  float64       `json:"confidence"`
}

type StockResult struct {
	Symbol    string `json:"symbol"`
	Score     int    `json:"score"`
	Reasoning string `json:"reasoning"`
	Timeframe string `json:"timeframe"`
}

func (a *Analyzer) Analyze(ctx context.Context, news *storage.NewsItem) (*storage.Analysis, error) {
	prompt := buildAnalysisPrompt(news)

	responseText, err := a.provider.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation error: %w", err)
	}

	result, err := parseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nResponse was: %s", err, responseText)
	}

	analysis := &storage.Analysis{
		ID:             fmt.Sprintf("ana_%d", time.Now().UnixNano()),
		NewsID:         news.ID,
		Sentiment:      result.Sentiment,
		ImpactLevel:    result.Impact,
		Summary:        result.Summary,
		SentimentScore: float64(calculateOverallScore(result.Stocks)),
		AnalyzedAt:     time.Now(),
		RawResponse:    responseText,
	}
	
	// 1. Get stocks identified by LLM
	stockSet := make(map[string]bool)
	var relatedStocks []string
	for _, s := range result.Stocks {
		if !stockSet[s.Symbol] {
			relatedStocks = append(relatedStocks, s.Symbol)
			stockSet[s.Symbol] = true
		}
	}

	// 2. Augment with StockMapper (rule-based)
	// Combine Title + Content for better matching
	fullText := news.Title + " " + news.Content
	mappedStocks := a.mapper.FindRelatedStocks(fullText)
	
	for _, symbol := range mappedStocks {
		if !stockSet[symbol] {
			// Found a stock not mentioned by LLM
			relatedStocks = append(relatedStocks, symbol)
			stockSet[symbol] = true
			// Optional: We could add a "neutral" entry to result.Stocks for UI consistency
			// result.Stocks = append(result.Stocks, StockResult{Symbol: symbol, Score: 0, Reasoning: "Identified by keyword association"})
		}
	}

	analysis.RelatedStocks = relatedStocks

	return analysis, nil
}

func calculateOverallScore(stocks []StockResult) int {
	if len(stocks) == 0 {
		return 0
	}
	total := 0
	for _, s := range stocks {
		total += s.Score
	}
	return total / len(stocks)
}

func (a *Analyzer) AnalyzeAndSave(ctx context.Context, news *storage.NewsItem) (*storage.Analysis, error) {
	analysis, err := a.Analyze(ctx, news)
	if err != nil {
		return nil, err
	}

	if err := a.store.SaveAnalysis(analysis); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return analysis, nil
}

func (a *Analyzer) AnalyzeBatch(ctx context.Context, items []storage.NewsItem) ([]storage.Analysis, error) {
	var results []storage.Analysis

	for _, item := range items {
		analysis, err := a.AnalyzeAndSave(ctx, &item)
		if err != nil {
			log.Printf("Failed to analyze news %s: %v", item.ID, err)
			continue
		}
		results = append(results, *analysis)
	}

	return results, nil
}

func buildAnalysisPrompt(news *storage.NewsItem) string {
	return fmt.Sprintf(`Analyze the following news/social media post for stock market impact.

Source: %s
Author: %s
Published: %s
Content:
%s

You must respond with valid JSON only. No other text. The JSON schema is:
{
  "sentiment": "positive|negative|neutral",
  "impact": "high|medium|low",
  "summary": "Brief summary of the content and its market implications",
  "stocks": [
    {
      "symbol": "TICKER",
      "score": -10 to +10 (negative = bearish, positive = bullish),
      "reasoning": "Why this stock is affected",
      "timeframe": "immediate|short|long"
    }
  ],
  "confidence": 0.5 to 1.0
}

Focus on:
- Direct company mentions
- Industry/sector implications
- Policy/regulatory impact
- Macro economic signals

Only include stocks with clear connection to the content. If no specific stocks are affected, return empty stocks array.`,
		news.Source, news.Author, news.PublishedAt.Format(time.RFC3339), news.Content)
}

func parseAnalysisResponse(response string) (*AnalysisResult, error) {
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			response = strings.Join(lines, "\n")
		}
	}

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start : end+1]
	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return &result, nil
}
