package builtins

import (
	"context"
	"fmt"
	"os"

	"simonos/internal/workspace"
)

type FileReadTool struct {
	workspace workspace.Workspace
}

func NewFileReadTool(workspace workspace.Workspace) *FileReadTool {
	return &FileReadTool{workspace: workspace}
}

func (t *FileReadTool) Name() string { return "file_read" }

func (t *FileReadTool) Description() string { return "Read a file within the workspace" }

func (t *FileReadTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	_ = ctx
	path, _ := input["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	resolved, err := t.workspace.Resolve(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"path": resolved, "content": string(content)}, nil
}
