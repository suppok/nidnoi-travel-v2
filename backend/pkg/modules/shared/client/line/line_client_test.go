package line

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"backend/pkg/core/config"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"destination":"U0","events":[]}`)
	c := New(config.LINEConfig{ChannelSecret: "secret"})

	cases := []struct {
		name      string
		client    Client
		body      []byte
		signature string
		want      bool
	}{
		{"valid", c, body, sign("secret", body), true},
		{"other secret", c, body, sign("other", body), false},
		{"tampered body", c, []byte(`{"destination":"U0","events":[{}]}`), sign("secret", body), false},
		{"empty signature", c, body, "", false},
		{"not base64", c, body, "%%%", false},
		{"no secret configured", New(config.LINEConfig{}), body, sign("", body), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.client.VerifySignature(tc.body, tc.signature); got != tc.want {
				t.Fatalf("VerifySignature() = %v, want %v", got, tc.want)
			}
		})
	}
}
