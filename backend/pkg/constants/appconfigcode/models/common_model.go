// Package models holds the value type of every app config code. JSON tags are camelCase.
// Give a value a `Validate() error` method (value receiver, pkg/core/validation) to have
// POST /app-configs/verify check its rules.
package models

// Shared wrappers for configs that hold a single scalar: {"value": ...}.
type AppSingleValue struct {
	Value string `json:"value"`
}

type AppSingleValueBool struct {
	Value bool `json:"value"`
}

type AppSingleValueNumber struct {
	Value int `json:"value"`
}
