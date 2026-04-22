package model

import (
	"context"
	"fmt"
	"time"
)

type OpenAIProvider struct {
	name string
}

func NewOpenAIProvider(name string) *OpenAIProvider {
	return &OpenAIProvider{name: name}
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return fmt.Sprintf("[%s] %s", p.name, prompt), nil
	}
}

func (p *OpenAIProvider) Stream(ctx context.Context, prompt string) (<-chan string, error) {
	output := make(chan string, 1)
	go func() {
		defer close(output)
		result, err := p.Generate(ctx, prompt)
		if err == nil {
			output <- result
		}
	}()
	return output, nil
}
