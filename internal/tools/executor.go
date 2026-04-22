package tools

import (
	"context"
	"fmt"
)

// Executor is the default tool dispatcher.
type Executor struct {
	registry *ToolRegistry
}

func NewExecutor(registry *ToolRegistry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Execute(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	tool, ok := e.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return tool.Run(ctx, input)
}
