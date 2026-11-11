package usecase

import (
	"testing"

	"backend/pkg/core/validation"
)

func TestPhone(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"0812345678", "0812345678", false},
		{" 081-234 5678 ", "0812345678", false},
		{"(02) 123.4567", "021234567", false},
		{"+66812345678", "+66812345678", false},
		{"12345678", "12345678", true},
		{"081234567a", "081234567a", true},
		{"66+812345678", "66+812345678", true},
	}
	for _, tc := range cases {
		got := tc.in
		normalizePhone(&got)
		if got != tc.want {
			t.Errorf("normalizePhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
		err := validation.Validate(got, phoneRules()...)
		if (err != nil) != tc.wantErr {
			t.Errorf("phoneRules(%q) error = %v, wantErr %v", got, err, tc.wantErr)
		}
	}
}
