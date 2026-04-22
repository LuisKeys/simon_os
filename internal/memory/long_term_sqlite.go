package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create memory directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS messages (session_id TEXT NOT NULL, role TEXT NOT NULL, content TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS embeddings (content TEXT NOT NULL, metadata TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) AddMessage(ctx context.Context, sessionID string, role string, content string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO messages(session_id, role, content) VALUES (?, ?, ?)`, sessionID, role, content)
	return err
}

func (s *SQLiteStore) GetMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT role, content FROM messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.Role, &message.Content); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (s *SQLiteStore) StoreEmbedding(ctx context.Context, content string, metadata map[string]interface{}) error {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO embeddings(content, metadata) VALUES (?, ?)`, content, string(encoded))
	return err
}

func (s *SQLiteStore) Search(ctx context.Context, query string, limit int) ([]MemoryResult, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT content, metadata FROM embeddings WHERE content LIKE ? ORDER BY created_at DESC LIMIT ?`, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]MemoryResult, 0, limit)
	for rows.Next() {
		var content string
		var metadataRaw string
		if err := rows.Scan(&content, &metadataRaw); err != nil {
			return nil, err
		}

		metadata := map[string]interface{}{}
		if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
			return nil, err
		}
		results = append(results, MemoryResult{Content: content, Metadata: metadata})
	}
	return results, rows.Err()
}
