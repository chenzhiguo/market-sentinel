package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type Analyzer struct {
	cfg    *config.Config
	store  *storage.Storage
	client *anthropic.Client
}

func New(cfg *config.Config, store *storage.Storage) *Analyzer {
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.Analyzer.APIKey),
	)

	return &Analyzer{
		cfg:    cfg,
		store:  store,
		client: client,
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

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(a.cfg.Analyzer.LLMModel),
		MaxTokens: anthropic.Int(1024),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if block.Type == anthropic.ContentBlockTypeText {
			responseText = block.Text
			break
		}
	}

	result, err := parseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	analysis := &storage.Analysis{
		ID:         fmt.Sprintf("ana_%d", time.Now().UnixNano()),
		NewsID:     news.ID,
		Sentiment:  result.Sentiment,
		Impact:     result.Impact,
		Summary:    result.Summary,
		Confidence: result.Confidence,
		AnalyzedAt: time.Now(),
	}

	for _, s := range result.Stocks {
		analysis.Stocks = append(analysis.Stocks, storage.StockImpact{
			Symbol:    s.Symbol,
			Score:     s.Score,
			Reasoning: s.Reasoning,
			Timeframe: s.Timeframe,
		})
	}

	return analysis, nil
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

Respond in JSON format:
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
  "confidence": 0.0 to 1.0
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
	// Find JSON in response
	start := -1
	end := -1
	depth := 0

	for i, c := range response {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start:end]
	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return &result, nil
}
