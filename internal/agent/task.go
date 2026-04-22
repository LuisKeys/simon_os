package agent

// Task is the runtime representation of a single user request.
type Task struct {
	ID     string
	Input  string
	State  map[string]interface{}
	Result string
}
