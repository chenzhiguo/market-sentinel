package llm

import (
	"context"
)

// Provider defines the interface for LLM interactions
type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
