package permission

import (
	"fmt"

	"github.com/casbin/casbin/v2"

	"backend/pkg/constants/apicode"
)

// Checker answers "may a user with these roles use this code?". A user is allowed when any
// of their roles is allowed.
type Checker interface {
	Allowed(roles []string, code apicode.Code) bool
	// Codes returns every declared permission code the roles allow (union across roles).
	Codes(roles []string) []string
}

type checker struct{ enforcer *casbin.Enforcer }

func New(modelPath, policyPath string) (Checker, error) {
	e, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return nil, fmt.Errorf("load casbin policy: %w", err)
	}
	return &checker{enforcer: e}, nil
}

func (c *checker) Allowed(roles []string, code apicode.Code) bool {
	for _, r := range roles {
		if ok, err := c.enforcer.Enforce(r, code.Object(), code.Action()); err == nil && ok {
			return true
		}
	}
	return false
}

func (c *checker) Codes(roles []string) []string {
	out := []string{}
	for _, code := range apicode.All() {
		if c.Allowed(roles, code) {
			out = append(out, code.Value)
		}
	}
	return out
}
