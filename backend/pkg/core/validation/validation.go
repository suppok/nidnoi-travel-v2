// Package validation wraps ozzo-validation so use cases depend on one project package.
package validation

import (
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type (
	Rule       = ozzo.Rule
	FieldRules = ozzo.FieldRules
	Errors     = ozzo.Errors
)

var (
	Required = ozzo.Required
	Nil      = ozzo.Nil
	NotNil   = ozzo.NotNil
	Email    = is.EmailFormat
	UUID     = is.UUID
)

func ValidateStruct(structPtr any, fields ...*FieldRules) error {
	return ozzo.ValidateStruct(structPtr, fields...)
}

func Validate(value any, rules ...Rule) error {
	return ozzo.Validate(value, rules...)
}

func Field(fieldPtr any, rules ...Rule) *FieldRules {
	return ozzo.Field(fieldPtr, rules...)
}

func Length(minLen, maxLen int) ozzo.LengthRule     { return ozzo.Length(minLen, maxLen) }
func RuneLength(minLen, maxLen int) ozzo.LengthRule { return ozzo.RuneLength(minLen, maxLen) }
func Min(min any) ozzo.ThresholdRule                { return ozzo.Min(min) }
func Max(max any) ozzo.ThresholdRule                { return ozzo.Max(max) }
func In(values ...any) ozzo.InRule                  { return ozzo.In(values...) }
func Each(rules ...Rule) ozzo.EachRule              { return ozzo.Each(rules...) }
func By(f ozzo.RuleFunc) Rule                       { return ozzo.By(f) }
