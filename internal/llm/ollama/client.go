package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/llm"
)

func init() {
	llm.Register("ollama", NewFactory)
}

type Client struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewFactory creates a new Ollama provider from config map
func NewFactory(cfg map[string]string) (llm.Provider, error) {
	baseURL := cfg["url"]
	model := cfg["model"]
	
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for ollama")
	}

	return New(baseURL, model), nil
}

func New(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type request struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
	Format   string `json:"format,omitempty"` 
	System   string `json:"system,omitempty"`
	Template string `json:"template,omitempty"`
}

type response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := request{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp response
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}
