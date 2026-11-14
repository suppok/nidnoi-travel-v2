package usecase

import (
	"slices"
	"testing"

	"backend/pkg/domain"
)

func TestNormalizeSlips(t *testing.T) {
	tests := []struct {
		name   string
		list   []string
		single string
		want   []string
	}{
		{"list wins", []string{" a ", "", "b"}, "x", []string{"a", "b"}},
		{"older single field", nil, " x ", []string{"x"}},
		{"empty list falls back", []string{"  "}, "x", []string{"x"}},
		{"nothing", nil, "", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSlips(tt.list, tt.single); !slices.Equal(got, tt.want) {
				t.Fatalf("normalizeSlips = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBookingSlips(t *testing.T) {
	older := &domain.LineBooking{SlipImage: "a"}
	if got := older.Slips(); !slices.Equal(got, []string{"a"}) {
		t.Fatalf("a row from before slip_images: Slips = %q", got)
	}
	multi := &domain.LineBooking{SlipImage: "a", SlipImages: []string{"a", "b"}}
	if got := multi.Slips(); !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("Slips = %q", got)
	}
	if got := (&domain.LineBooking{}).BalanceSlips(); got == nil || len(got) != 0 {
		t.Fatalf("no balance: BalanceSlips = %#v, want an empty list", got)
	}
}
