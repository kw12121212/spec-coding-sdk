package permission

import (
	"context"
	"os"
	"strings"
)

// DefaultPolicy returns a full-access policy that allows all operations.
func DefaultPolicy() *StaticPolicy {
	return NewStaticPolicy(DecisionAllow)
}

// safePolicy wraps a StaticPolicy and adds a CWD guard for write/edit operations.
type safePolicy struct {
	rules *StaticPolicy
	cwd   string
}

// Evaluate checks CWD guard for write/edit operations, then delegates to the inner StaticPolicy.
func (p *safePolicy) Evaluate(ctx context.Context, operation, resource string) Decision {
	if operation == "file:write" || operation == "file:edit" {
		if p.cwd != "" && !strings.HasPrefix(resource, p.cwd) {
			return DecisionDeny
		}
	}
	return p.rules.Evaluate(ctx, operation, resource)
}

// SafePolicy returns a restrictive policy that:
//   - Denies destructive bash commands (rm -rf, mkfs, dd if=, fork bomb)
//   - Denies file write/edit outside the current working directory
//   - Allows everything else
//
// The working directory is captured at call time.
func SafePolicy() Policy {
	cwd, _ := os.Getwd()
	return &safePolicy{
		cwd: cwd,
		rules: NewStaticPolicy(DecisionAllow,
			Rule{OperationPattern: "bash:execute", ResourcePattern: "rm -rf", Decision: DecisionDeny},
			Rule{OperationPattern: "bash:execute", ResourcePattern: "rm -r /*", Decision: DecisionDeny},
			Rule{OperationPattern: "bash:execute", ResourcePattern: "mkfs", Decision: DecisionDeny},
			Rule{OperationPattern: "bash:execute", ResourcePattern: "dd if=", Decision: DecisionDeny},
			Rule{OperationPattern: "bash:execute", ResourcePattern: ":(){ :|:& };:", Decision: DecisionDeny},
		),
	}
}
