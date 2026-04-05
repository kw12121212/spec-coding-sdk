package permission

import "context"

// Policy evaluates permission requests and returns a Decision.
type Policy interface {
	Evaluate(ctx context.Context, operation string, resource string) Decision
}

// StaticPolicy is an ordered list of rules with a default decision.
type StaticPolicy struct {
	Rules           []Rule
	DefaultDecision Decision
}

// NewStaticPolicy creates a StaticPolicy with the given default decision and rules.
func NewStaticPolicy(defaultDecision Decision, rules ...Rule) *StaticPolicy {
	return &StaticPolicy{
		Rules:           rules,
		DefaultDecision: defaultDecision,
	}
}

// Evaluate returns the decision of the first matching rule, or DefaultDecision if none match.
func (p *StaticPolicy) Evaluate(_ context.Context, operation, resource string) Decision {
	for _, r := range p.Rules {
		if r.Match(operation, resource) {
			return r.Decision
		}
	}
	return p.DefaultDecision
}
