package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"simonos/internal/events"
	"simonos/internal/guardrails"
	"simonos/internal/memory"
	"simonos/internal/model"
	"simonos/internal/tools"
)

// Engine orchestrates the execution loop.
type Engine struct {
	tools    tools.ToolExecutor
	memory   memory.MemoryStore
	router   model.ModelRouter
	policies guardrails.PolicyEngine
	eventBus events.EventBus
}

func NewEngine(toolExecutor tools.ToolExecutor, memoryStore memory.MemoryStore, router model.ModelRouter, policies guardrails.PolicyEngine, eventBus events.EventBus) *Engine {
	return &Engine{
		tools:    toolExecutor,
		memory:   memoryStore,
		router:   router,
		policies: policies,
		eventBus: eventBus,
	}
}

func (e *Engine) Run(ctx context.Context, input string) (string, error) {
	task := Task{
		ID:    fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Input: strings.TrimSpace(input),
		State: map[string]interface{}{},
	}

	if err := e.memory.AddMessage(ctx, task.ID, "user", task.Input); err != nil {
		return "", err
	}
	e.eventBus.Publish(events.Event{Type: events.EventMemoryUpdate, Payload: map[string]interface{}{"task_id": task.ID, "role": "user"}})

	result, err := e.execute(ctx, task)
	if err != nil {
		e.eventBus.Publish(events.Event{Type: events.EventError, Payload: map[string]interface{}{"task_id": task.ID, "error": err.Error()}})
		return "", err
	}

	if err := e.memory.AddMessage(ctx, task.ID, "assistant", result); err != nil {
		return "", err
	}
	e.eventBus.Publish(events.Event{Type: events.EventFinalOutput, Payload: map[string]interface{}{"task_id": task.ID, "result": result}})
	return result, nil
}

func (e *Engine) Stream(ctx context.Context, input string) (<-chan events.Event, error) {
	stream := make(chan events.Event, 16)

	go func() {
		defer close(stream)
		forward := e.eventBus.Subscribe()

		result, err := e.Run(ctx, input)
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-forward:
				stream <- event
				if event.Type == events.EventFinalOutput || event.Type == events.EventError {
					return
				}
			default:
				if err != nil {
					stream <- events.Event{Type: events.EventError, Payload: map[string]interface{}{"error": err.Error()}}
					return
				}
				if result != "" {
					stream <- events.Event{Type: events.EventTokenStream, Payload: map[string]interface{}{"chunk": result}}
					stream <- events.Event{Type: events.EventFinalOutput, Payload: map[string]interface{}{"result": result}}
					return
				}
			}
		}
	}()

	return stream, nil
}

func (e *Engine) execute(ctx context.Context, task Task) (string, error) {
	action := nextAction(task.Input)
	decision, err := e.policies.Evaluate(guardrails.Action{Type: string(action.Type), Name: action.Name, Input: action.Input})
	if err != nil {
		return "", err
	}
	if decision == guardrails.Deny {
		return "", fmt.Errorf("action %q denied by policy", action.Name)
	}
	if decision == guardrails.RequireApproval {
		e.eventBus.Publish(events.Event{Type: events.EventApproval, Payload: map[string]interface{}{"task_id": task.ID, "action": action.Name}})
		return "", fmt.Errorf("action %q requires approval", action.Name)
	}

	if action.Type == ActionToolCall {
		e.eventBus.Publish(events.Event{Type: events.EventToolCall, Payload: map[string]interface{}{"task_id": task.ID, "tool": action.Name}})
		output, err := e.tools.Execute(ctx, action.Name, action.Input)
		if err != nil {
			return "", err
		}
		e.eventBus.Publish(events.Event{Type: events.EventToolResult, Payload: map[string]interface{}{"task_id": task.ID, "tool": action.Name, "output": output}})
		return fmt.Sprintf("tool %s result: %v", action.Name, output), nil
	}

	provider := e.router.SelectModel(task.Input)
	prompt := buildPrompt(task.Input)
	result, err := provider.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}
	_ = e.memory.StoreEmbedding(ctx, result, map[string]interface{}{"task_id": task.ID})
	return result, nil
}

func nextAction(input string) Action {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if strings.HasPrefix(trimmed, "read ") {
		return Action{Type: ActionToolCall, Name: "file_read", Input: map[string]interface{}{"path": strings.TrimSpace(input[5:])}}
	}
	if strings.HasPrefix(trimmed, "write ") {
		parts := strings.SplitN(input[6:], ":", 2)
		if len(parts) == 2 {
			return Action{Type: ActionToolCall, Name: "file_write", Input: map[string]interface{}{"path": strings.TrimSpace(parts[0]), "content": strings.TrimSpace(parts[1])}}
		}
	}
	if strings.HasPrefix(trimmed, "shell ") {
		return Action{Type: ActionToolCall, Name: "shell", Input: map[string]interface{}{"command": strings.TrimSpace(input[6:])}}
	}
	return Action{Type: ActionModelCall, Name: "generate", Input: map[string]interface{}{"prompt": input}}
}

func buildPrompt(input string) string {
	return fmt.Sprintf("You are SimonOS. Respond concisely and helpfully to: %s", strings.TrimSpace(input))
}
