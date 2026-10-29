package apicode

import "strings"

// Code is a permission code "module:action", checked by Casbin as (role, module, action).
type Code struct {
	Value       string
	Description string
}

var all []Code

// ApiCode declares a permission code. Declare codes as package-level vars so All() lists them.
func ApiCode(value, description string) Code {
	c := Code{Value: value, Description: description}
	all = append(all, c)
	return c
}

// Object and Action split the code for Casbin.
func (c Code) Object() string {
	obj, _, _ := strings.Cut(c.Value, ":")
	return obj
}

func (c Code) Action() string {
	_, act, _ := strings.Cut(c.Value, ":")
	return act
}

// All returns every declared code (used to compute a user's permission list for /auth/me).
func All() []Code {
	return append([]Code(nil), all...)
}
