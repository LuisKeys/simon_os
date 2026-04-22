package events

// Event is the common envelope emitted by runtime components.
type Event struct {
	Type    string
	Payload map[string]interface{}
}

const (
	EventTokenStream  = "token_stream"
	EventToolCall     = "tool_call"
	EventToolResult   = "tool_result"
	EventMemoryUpdate = "memory_update"
	EventApproval     = "approval_request"
	EventFinalOutput  = "final_output"
	EventError        = "error"
)
