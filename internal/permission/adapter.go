package permission

import (
	"context"
	"fmt"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// PolicyProvider adapts a Policy to implement core.PermissionProvider.
type PolicyProvider struct {
	policy Policy
}

// NewPolicyProvider creates a PolicyProvider that delegates to the given Policy.
func NewPolicyProvider(policy Policy) *PolicyProvider {
	return &PolicyProvider{policy: policy}
}

// Check implements core.PermissionProvider.
func (p *PolicyProvider) Check(ctx context.Context, operation, resource string) error {
	d := p.policy.Evaluate(ctx, operation, resource)
	switch d {
	case DecisionAllow:
		return nil
	case DecisionDeny:
		return fmt.Errorf("permission denied: %s on %s", operation, resource)
	case DecisionNeedConfirmation:
		return fmt.Errorf("confirmation required: %s on %s", operation, resource)
	default:
		return fmt.Errorf("permission denied: %s on %s", operation, resource)
	}
}

// Compile-time interface check.
var _ core.PermissionProvider = (*PolicyProvider)(nil)
