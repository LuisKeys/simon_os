package builtins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"simonos/internal/workspace"
)

type FileWriteTool struct {
	workspace workspace.Workspace
}

func NewFileWriteTool(workspace workspace.Workspace) *FileWriteTool {
	return &FileWriteTool{workspace: workspace}
}

func (t *FileWriteTool) Name() string { return "file_write" }

func (t *FileWriteTool) Description() string { return "Write a file within the workspace" }

func (t *FileWriteTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	_ = ctx
	path, _ := input["path"].(string)
	content, _ := input["content"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	resolved, err := t.workspace.Resolve(path)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return nil, err
	}

	return map[string]interface{}{"path": resolved, "bytes": len(content)}, nil
}
