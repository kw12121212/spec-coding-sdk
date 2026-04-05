// Package permission provides a permission model with typed decisions,
// composable rules, and policy evaluation for tool execution authorization.
package permission

// Decision represents the outcome of a permission evaluation.
type Decision string

const (
	// DecisionAllow grants permission to proceed.
	DecisionAllow Decision = "allow"
	// DecisionDeny blocks the operation.
	DecisionDeny Decision = "deny"
	// DecisionNeedConfirmation indicates the operation requires user confirmation.
	DecisionNeedConfirmation Decision = "need_confirmation"
)

// String returns the string representation of the decision.
func (d Decision) String() string {
	return string(d)
}
