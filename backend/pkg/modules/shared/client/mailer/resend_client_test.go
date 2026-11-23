package mailer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type captured struct {
	method, path, auth, key, contentType string
	body                                 map[string]any
}

func newServer(t *testing.T, status int, respBody string, got *captured) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method, got.path = r.Method, r.URL.Path
		got.auth = r.Header.Get("Authorization")
		got.key = r.Header.Get("Idempotency-Key")
		got.contentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		got.body = map[string]any{}
		_ = json.Unmarshal(raw, &got.body)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func message() Message {
	return Message{
		From:           "nidnoi travel <booking@example.com>",
		To:             []string{"ws@example.com", "ws2@example.com"},
		Cc:             []string{"owner@example.com"},
		ReplyTo:        "owner@example.com",
		Subject:        "subject",
		HTML:           "<p>hi</p>",
		Text:           "hi",
		IdempotencyKey: "outbox-1-1",
	}
}

func TestResendSend(t *testing.T) {
	var got captured
	srv := newServer(t, http.StatusOK, `{"id":"msg_123"}`, &got)
	id, err := NewResend("re_test", srv.URL+"/emails", time.Second).Send(context.Background(), message())
	if err != nil || id != "msg_123" {
		t.Fatalf("Send = %q, %v", id, err)
	}
	if got.method != http.MethodPost || got.path != "/emails" {
		t.Errorf("request %s %s", got.method, got.path)
	}
	if got.auth != "Bearer re_test" || got.key != "outbox-1-1" || got.contentType != "application/json" {
		t.Errorf("headers auth=%q key=%q type=%q", got.auth, got.key, got.contentType)
	}
	if got.body["reply_to"] != "owner@example.com" || got.body["from"] != "nidnoi travel <booking@example.com>" {
		t.Errorf("body %v", got.body)
	}
	if cc, _ := got.body["cc"].([]any); len(cc) != 1 || cc[0] != "owner@example.com" {
		t.Errorf("cc %v", got.body["cc"])
	}
	if to, _ := got.body["to"].([]any); len(to) != 2 {
		t.Errorf("to %v", got.body["to"])
	}
	for _, k := range []string{"subject", "html", "text"} {
		if _, ok := got.body[k]; !ok {
			t.Errorf("missing %s", k)
		}
	}
}

func TestResendOmitsEmptyCcAndReplyTo(t *testing.T) {
	var got captured
	srv := newServer(t, http.StatusOK, `{"id":"msg_1"}`, &got)
	msg := message()
	msg.Cc, msg.ReplyTo = nil, ""
	if _, err := NewResend("re_test", srv.URL, time.Second).Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if _, ok := got.body["cc"]; ok {
		t.Error("cc sent while empty")
	}
	if _, ok := got.body["reply_to"]; ok {
		t.Error("reply_to sent while empty")
	}
}

func TestResendErrors(t *testing.T) {
	cases := []struct {
		status    int
		retryable bool
	}{
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusUnprocessableEntity, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
	}
	for _, c := range cases {
		var got captured
		srv := newServer(t, c.status, `{"message":"nope"}`, &got)
		_, err := NewResend("re_secret", srv.URL, time.Second).Send(context.Background(), message())
		var se *SendError
		if !errors.As(err, &se) || se.Retryable != c.retryable || se.Status != c.status || se.Message != "nope" {
			t.Errorf("status %d: err = %#v", c.status, err)
		}
		if IsRetryable(err) != c.retryable {
			t.Errorf("status %d: IsRetryable mismatch", c.status)
		}
	}
}

func TestResendNetworkErrorIsRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	_, err := NewResend("re_secret", url, time.Second).Send(context.Background(), message())
	if !IsRetryable(err) {
		t.Fatalf("network error not retryable: %v", err)
	}
	if err != nil && (strings.Contains(err.Error(), "re_secret") || strings.Contains(err.Error(), url)) {
		t.Errorf("error leaks details: %v", err)
	}
}
