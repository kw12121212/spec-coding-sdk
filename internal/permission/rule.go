package permission

import (
	"path/filepath"
	"strings"
)

// Rule matches an operation and optional resource against patterns to produce a Decision.
type Rule struct {
	// OperationPattern matches the operation string using filepath.Match syntax (e.g. "bash:*", "file:read").
	OperationPattern string
	// ResourcePattern matches using substring containment (strings.Contains). Empty means match all resources.
	ResourcePattern string
	// Decision is the result when this rule matches.
	Decision Decision
}

// Match returns true if the rule's patterns match the given operation and resource.
func (r Rule) Match(operation, resource string) bool {
	ok, _ := filepath.Match(r.OperationPattern, operation)
	if !ok {
		return false
	}
	if r.ResourcePattern == "" {
		return true
	}
	return strings.Contains(resource, r.ResourcePattern)
}
