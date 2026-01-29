package anthropic

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/chenzhiguo/market-sentinel/internal/llm"
)

func init() {
	llm.Register("anthropic", NewFactory)
}

type Client struct {
	client *anthropic.Client
	model  string
}

// NewFactory creates a new Anthropic provider from config map
func NewFactory(cfg map[string]string) (llm.Provider, error) {
	apiKey := cfg["api_key"]
	model := cfg["model"]
	
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for anthropic")
	}
	if model == "" {
		model = "claude-3-opus-20240229" // default
	}

	return New(apiKey, model), nil
}

func New(apiKey, model string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Client{
		client: client,
		model:  model,
	}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.Int(1024),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})
	if err != nil {
		return "", err
	}

	for _, block := range message.Content {
		if block.Type == anthropic.ContentBlockTypeText {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}
