package agent

// ActionType identifies a unit of work in the execution loop.
type ActionType string

const (
	ActionModelCall ActionType = "model_call"
	ActionToolCall  ActionType = "tool_call"
	ActionMemoryOp  ActionType = "memory_op"
)

// Action captures the selected next step from the model/runtime.
type Action struct {
	Type  ActionType
	Name  string
	Input map[string]interface{}
}
