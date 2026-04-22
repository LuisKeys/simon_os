package guardrails

// Action is the policy-facing representation of a runtime action.
type Action struct {
	Type  string
	Name  string
	Input map[string]interface{}
}

// Decision is the result of policy evaluation.
type Decision string

const (
	Allow           Decision = "allow"
	Deny            Decision = "deny"
	RequireApproval Decision = "require_approval"
)

// PolicyEngine evaluates proposed actions.
type PolicyEngine interface {
	Evaluate(action Action) (Decision, error)
}
