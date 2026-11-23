// Package core holds the typed app config code. Codes are declared in package appconfigcode.
package core

import (
	"bytes"
	"encoding/json"

	"backend/pkg/core/validation"
)

// AppConfig is the type-erased view of a code, used by the registry and the verify endpoint.
type AppConfig interface {
	Code() string
	// Validate decodes data strictly (unknown fields are errors) and runs the value's rules.
	Validate(data string) error
}

// AppConfigCode is a config code whose JSON value decodes into T.
type AppConfigCode[T any] struct{ code string }

func NewJsonConfig[T any](code string) *AppConfigCode[T] {
	return &AppConfigCode[T]{code: code}
}

func (c *AppConfigCode[T]) Code() string { return c.code }

// ToValue decodes the stored JSON into T. Unknown fields are ignored, so a row written by newer
// code still reads on older code.
func (c *AppConfigCode[T]) ToValue(data string) (T, error) {
	var t T
	err := json.Unmarshal([]byte(data), &t)
	return t, err
}

// Validate decodes data into T, rejecting unknown fields (typos in hand-edited rows), then
// validates it. T (or each element, when T is a slice or map) is checked when it has a
// `Validate() error` method (value receiver) built with pkg/core/validation.
func (c *AppConfigCode[T]) Validate(data string) error {
	var t T
	dec := json.NewDecoder(bytes.NewReader([]byte(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return err
	}
	return validation.Validate(t)
}
