package password

import (
	"strings"
	"testing"
)

func TestProblem(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"abc12345", true},
		{"ABCDEFG1", true},
		{"abc1234", false},    // too short
		{"abcdefgh", false},   // no digit
		{"12345678", false},   // no letter
		{"รหัสผ่าน12", false}, // Thai letters don't count as an English letter
		{"รหัสผ่านa1", true},
		{strings.Repeat("a1", 36), true},        // 72 bytes
		{strings.Repeat("a1", 36) + "x", false}, // 73 bytes, bcrypt limit
	}
	for _, c := range cases {
		if got := Problem(c.in) == nil; got != c.ok {
			t.Errorf("Problem(%q) ok=%v, want %v", c.in, got, c.ok)
		}
	}
}

func TestGenerateTemporaryAndMatches(t *testing.T) {
	s, err := NewService(10)
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		p, err := s.GenerateTemporary()
		if err != nil {
			t.Fatal(err)
		}
		if Problem(p) != nil || strings.ContainsAny(p, "0O1lI") {
			t.Fatalf("bad temporary password %q", p)
		}
	}

	hash, err := s.Hash("abc12345")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Matches(hash, "abc12345") || s.Matches(hash, "abc12346") || s.Matches("", "abc12345") {
		t.Fatal("Matches returned the wrong result")
	}
}
