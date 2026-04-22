package memory

import (
	"context"
	"strings"
	"sync"
)

type ShortTermStore struct {
	mu       sync.RWMutex
	messages map[string][]Message
	entries  []MemoryResult
}

func NewShortTermStore() *ShortTermStore {
	return &ShortTermStore{messages: map[string][]Message{}, entries: []MemoryResult{}}
}

func (s *ShortTermStore) AddMessage(ctx context.Context, sessionID string, role string, content string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[sessionID] = append(s.messages[sessionID], Message{Role: role, Content: content})
	return nil
}

func (s *ShortTermStore) GetMessages(ctx context.Context, sessionID string) ([]Message, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	messages := s.messages[sessionID]
	copyMessages := make([]Message, len(messages))
	copy(copyMessages, messages)
	return copyMessages, nil
}

func (s *ShortTermStore) StoreEmbedding(ctx context.Context, content string, metadata map[string]interface{}) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, MemoryResult{Content: content, Metadata: metadata})
	return nil
}

func (s *ShortTermStore) Search(ctx context.Context, query string, limit int) ([]MemoryResult, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := make([]MemoryResult, 0, limit)
	for _, entry := range s.entries {
		if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(query)) {
			results = append(results, entry)
			if len(results) == limit {
				break
			}
		}
	}
	return results, nil
}

type CombinedStore struct {
	shortTerm *ShortTermStore
	longTerm  *SQLiteStore
}

func NewCombinedStore(shortTerm *ShortTermStore, longTerm *SQLiteStore) *CombinedStore {
	return &CombinedStore{shortTerm: shortTerm, longTerm: longTerm}
}

func (c *CombinedStore) AddMessage(ctx context.Context, sessionID string, role string, content string) error {
	if err := c.shortTerm.AddMessage(ctx, sessionID, role, content); err != nil {
		return err
	}
	return c.longTerm.AddMessage(ctx, sessionID, role, content)
}

func (c *CombinedStore) GetMessages(ctx context.Context, sessionID string) ([]Message, error) {
	messages, err := c.longTerm.GetMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(messages) > 0 {
		return messages, nil
	}
	return c.shortTerm.GetMessages(ctx, sessionID)
}

func (c *CombinedStore) StoreEmbedding(ctx context.Context, content string, metadata map[string]interface{}) error {
	if err := c.shortTerm.StoreEmbedding(ctx, content, metadata); err != nil {
		return err
	}
	return c.longTerm.StoreEmbedding(ctx, content, metadata)
}

func (c *CombinedStore) Search(ctx context.Context, query string, limit int) ([]MemoryResult, error) {
	results, err := c.longTerm.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results, nil
	}
	return c.shortTerm.Search(ctx, query, limit)
}
