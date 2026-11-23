package core

import (
	"testing"

	"backend/pkg/core/validation"
)

type testValue struct {
	Size int `json:"size"`
}

func (v testValue) Validate() error {
	return validation.ValidateStruct(&v, validation.Field(&v.Size, validation.Required, validation.Min(1)))
}

func TestToValue(t *testing.T) {
	c := NewJsonConfig[testValue]("TEST")
	v, err := c.ToValue(`{"size": 3, "extra": true}`)
	if err != nil || v.Size != 3 {
		t.Fatalf("ToValue = %+v, %v", v, err)
	}
}

func TestValidate(t *testing.T) {
	c := NewJsonConfig[testValue]("TEST")
	cases := map[string]bool{
		`{"size": 3}`:            true,
		`{"size": 0}`:            false, // rule
		`{"size": "3"}`:          false, // type
		`{"size": 3, "typo": 1}`: false, // unknown field
		`not json`:               false,
	}
	for data, ok := range cases {
		if err := c.Validate(data); (err == nil) != ok {
			t.Errorf("Validate(%s) = %v, want ok=%v", data, err, ok)
		}
	}

	list := NewJsonConfig[[]testValue]("TEST_LIST")
	if err := list.Validate(`[{"size": 1}, {"size": 2}]`); err != nil {
		t.Errorf("slice valid: %v", err)
	}
	if err := list.Validate(`[{"size": 1}, {"size": 0}]`); err == nil {
		t.Error("slice with an invalid element passed")
	}
}
