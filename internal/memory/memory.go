package memory

import "context"

// Message captures a conversational exchange.
type Message struct {
	Role    string
	Content string
}

// MemoryResult is a simple search hit for long-term memory.
type MemoryResult struct {
	Content  string
	Metadata map[string]interface{}
}

// MemoryStore abstracts short-term and long-term memory operations.
type MemoryStore interface {
	AddMessage(ctx context.Context, sessionID string, role string, content string) error
	GetMessages(ctx context.Context, sessionID string) ([]Message, error)
	StoreEmbedding(ctx context.Context, content string, metadata map[string]interface{}) error
	Search(ctx context.Context, query string, limit int) ([]MemoryResult, error)
}
