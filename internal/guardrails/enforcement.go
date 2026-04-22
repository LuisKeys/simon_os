package guardrails

type DefaultPolicyEngine struct {
	requireApproval bool
}

func NewPolicyEngine(requireApproval bool) *DefaultPolicyEngine {
	return &DefaultPolicyEngine{requireApproval: requireApproval}
}

func (p *DefaultPolicyEngine) Evaluate(action Action) (Decision, error) {
	if action.Type == "tool_call" && action.Name == "shell" {
		if p.requireApproval {
			return RequireApproval, nil
		}
		return Deny, nil
	}
	return Allow, nil
}
