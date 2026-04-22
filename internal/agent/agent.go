package agent

import (
	"context"

	"simonos/internal/events"
)

// Agent is the top-level runtime contract exposed to the CLI.
type Agent interface {
	Run(ctx context.Context, input string) (string, error)
	Stream(ctx context.Context, input string) (<-chan events.Event, error)
}

// Controller adapts the Engine to the Agent interface.
type Controller struct {
	engine *Engine
}

func NewController(engine *Engine) *Controller {
	return &Controller{engine: engine}
}

func (c *Controller) Run(ctx context.Context, input string) (string, error) {
	return c.engine.Run(ctx, input)
}

func (c *Controller) Stream(ctx context.Context, input string) (<-chan events.Event, error) {
	return c.engine.Stream(ctx, input)
}
