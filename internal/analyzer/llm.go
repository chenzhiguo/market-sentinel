package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"market-sentinel/internal/config"
	"market-sentinel/internal/storage"
)

// ClaudeAnalyzer Claude API分析器
type ClaudeAnalyzer struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

func NewClaudeAnalyzer(cfg config.ClaudeConfig) *ClaudeAnalyzer {
	return &ClaudeAnalyzer{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// AnalysisResult Claude返回的分析结果结构
type AnalysisResult struct {
	Sentiment      string   `json:"sentiment"`
	SentimentScore float64  `json:"sentiment_score"`
	Entities       []Entity `json:"entities"`
	RelatedStocks  []string `json:"related_stocks"`
	ImpactLevel    string   `json:"impact_level"`
	Summary        string   `json:"summary"`
	KeyPoints      []string `json:"key_points"`
}

type Entity struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

const analysisPrompt = `你是一个专业的金融舆情分析师。请分析以下新闻/社交媒体内容，并以JSON格式返回分析结果。

内容来源: %s
作者: %s
标题: %s
内容: %s

请返回以下JSON格式的分析结果（只返回JSON，不要其他内容）：
{
  "sentiment": "positive/negative/neutral",
  "sentiment_score": -1.0到1.0之间的数值,
  "entities": [{"name": "实体名称", "type": "person/company/stock/policy/event"}],
  "related_stocks": ["相关股票代码列表，如TSLA, AAPL"],
  "impact_level": "high/medium/low",
  "summary": "一句话总结",
  "key_points": ["关键要点1", "关键要点2"]
}

分析时请注意：
1. 情感分析要考虑对市场的实际影响
2. 识别所有提到的公司、人物、政策
3. 关联可能受影响的股票代码
4. 评估事件对市场的影响程度`

// Analyze 分析单条新闻
func (a *ClaudeAnalyzer) Analyze(news *storage.NewsItem) (*storage.Analysis, error) {
	prompt := fmt.Sprintf(analysisPrompt, news.Source, news.Author, news.Title, news.Content)

	result, rawResponse, err := a.callClaude(prompt)
	if err != nil {
		return nil, err
	}

	// 转换实体类型
	var entities []storage.Entity
	for _, e := range result.Entities {
		entities = append(entities, storage.Entity{
			Name: e.Name,
			Type: e.Type,
		})
	}

	analysis := &storage.Analysis{
		ID:             storage.NewUUID(),
		NewsID:         news.ID,
		Sentiment:      result.Sentiment,
		SentimentScore: result.SentimentScore,
		Entities:       entities,
		RelatedStocks:  result.RelatedStocks,
		ImpactLevel:    result.ImpactLevel,
		Summary:        result.Summary,
		KeyPoints:      result.KeyPoints,
		AnalyzedAt:     time.Now(),
		RawResponse:    rawResponse,
	}

	return analysis, nil
}

func (a *ClaudeAnalyzer) callClaude(prompt string) (*AnalysisResult, string, error) {
	reqBody := claudeRequest{
		Model:     a.model,
		MaxTokens: a.maxTokens,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, string(body), err
	}

	if claudeResp.Error != nil {
		return nil, string(body), fmt.Errorf("claude error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return nil, string(body), fmt.Errorf("empty response from claude")
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(claudeResp.Content[0].Text), &result); err != nil {
		// 尝试提取JSON
		text := claudeResp.Content[0].Text
		start := bytes.IndexByte([]byte(text), '{')
		end := bytes.LastIndexByte([]byte(text), '}')
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(text[start:end+1]), &result); err != nil {
				return nil, string(body), err
			}
		} else {
			return nil, string(body), err
		}
	}

	return &result, string(body), nil
}

// GenerateSummary 生成汇总报告
func (a *ClaudeAnalyzer) GenerateSummary(analyses []storage.Analysis) (string, error) {
	if len(analyses) == 0 {
		return "暂无分析数据", nil
	}

	var summaries []string
	for _, analysis := range analyses {
		summaries = append(summaries, fmt.Sprintf("- %s (情感: %s, 影响: %s)", 
			analysis.Summary, analysis.Sentiment, analysis.ImpactLevel))
	}

	prompt := fmt.Sprintf(`请根据以下舆情分析结果，生成一份简洁的市场舆情汇总报告：

%s

请生成一份200字以内的汇总报告，包括：
1. 整体市场情绪
2. 关键事件
3. 需要关注的风险点`, 
		fmt.Sprintf("%v", summaries))

	reqBody := claudeRequest{
		Model:     a.model,
		MaxTokens: a.maxTokens,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	if len(claudeResp.Content) > 0 {
		return claudeResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no response")
}
