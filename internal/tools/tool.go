package tools

import "context"

// Tool is the contract implemented by all built-in and future plugin tools.
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// ToolExecutor dispatches tools by name.
type ToolExecutor interface {
	Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error)
}
