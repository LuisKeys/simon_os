package model

import "context"

// ModelProvider abstracts a concrete text generation backend.
type ModelProvider interface {
	Name() string
	Generate(ctx context.Context, prompt string) (string, error)
	Stream(ctx context.Context, prompt string) (<-chan string, error)
}
