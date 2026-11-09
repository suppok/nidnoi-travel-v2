package utils

import (
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	const alphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	seen := map[string]bool{}
	for range 1000 {
		s, err := RandomString(6, alphabet)
		if err != nil {
			t.Fatal(err)
		}
		if len(s) != 6 {
			t.Fatalf("RandomString length = %d, want 6", len(s))
		}
		for _, c := range s {
			if !strings.ContainsRune(alphabet, c) {
				t.Fatalf("RandomString(%q) contains %q outside the alphabet", s, c)
			}
		}
		seen[s] = true
	}
	if len(seen) < 990 {
		t.Fatalf("only %d distinct values out of 1000", len(seen))
	}
}
