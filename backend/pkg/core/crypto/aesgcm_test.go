package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func testKey() []byte { return bytes.Repeat([]byte{7}, KeySize) }

func TestSealOpenRoundTrip(t *testing.T) {
	s, err := NewSealer(testKey())
	if err != nil {
		t.Fatal(err)
	}
	ct, err := s.Seal([]byte("sk-ant-secret"), []byte("AI_SETTINGS"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Open(ct, []byte("AI_SETTINGS"))
	if err != nil || string(got) != "sk-ant-secret" {
		t.Fatalf("Open = %q, %v", got, err)
	}
	if _, err := s.Open(ct, []byte("OTHER")); err != ErrDecrypt {
		t.Fatalf("wrong additional data: err = %v, want ErrDecrypt", err)
	}
	other, _ := NewSealer(bytes.Repeat([]byte{8}, KeySize))
	if _, err := other.Open(ct, []byte("AI_SETTINGS")); err != ErrDecrypt {
		t.Fatalf("wrong key: err = %v, want ErrDecrypt", err)
	}
	ct2, _ := s.Seal([]byte("sk-ant-secret"), []byte("AI_SETTINGS"))
	if ct == ct2 {
		t.Fatal("two seals of the same value must differ (random nonce)")
	}
}

func TestParseKey(t *testing.T) {
	key := testKey()
	for _, s := range []string{base64.StdEncoding.EncodeToString(key), base64.RawURLEncoding.EncodeToString(key)} {
		got, err := ParseKey(s)
		if err != nil || !bytes.Equal(got, key) {
			t.Fatalf("ParseKey(%q) = %v, %v", s, got, err)
		}
	}
	if _, err := ParseKey(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Fatal("short key must fail")
	}
	if _, err := ParseKey("not base64 !!"); err == nil {
		t.Fatal("non-base64 must fail")
	}
}
