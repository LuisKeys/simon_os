package builtins

import (
	"context"
	"fmt"
	"os/exec"

	"simonos/internal/workspace"
)

type ShellTool struct {
	workspace workspace.Workspace
	enabled   bool
}

func NewShellTool(workspace workspace.Workspace, enabled bool) *ShellTool {
	return &ShellTool{workspace: workspace, enabled: enabled}
}

func (t *ShellTool) Name() string { return "shell" }

func (t *ShellTool) Description() string {
	return "Run a restricted shell command from the workspace root"
}

func (t *ShellTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if !t.enabled {
		return nil, fmt.Errorf("shell tool is disabled by configuration")
	}

	command, _ := input["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Dir = t.workspace.RootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{"output": string(output)}, fmt.Errorf("shell command failed: %w", err)
	}

	return map[string]interface{}{"output": string(output)}, nil
}
