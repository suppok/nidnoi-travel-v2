package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"backend/pkg/core/config"
)

func testClient(url string) *client {
	c := New(config.AIConfig{RequestTimeout: 10 * time.Second, MaxConcurrent: 1, MaxRetries: 2}, zap.NewNop()).(*client)
	c.baseURL = url
	c.backoff = []time.Duration{time.Millisecond}
	return c
}

// sse writes a minimal Messages stream whose text is answer.
func sse(w http.ResponseWriter, answer, stopReason string) {
	w.Header().Set("Content-Type", "text/event-stream")
	text, _ := json.Marshal(answer)
	events := []string{
		`{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5-5","content":[],"stop_reason":null,"usage":{"input_tokens":1200,"output_tokens":1,"cache_read_input_tokens":800,"cache_creation_input_tokens":0}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":` + string(text) + `}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"` + stopReason + `"},"usage":{"output_tokens":42}}`,
		`{"type":"message_stop"}`,
	}
	for _, e := range events {
		var probe struct{ Type string }
		_ = json.Unmarshal([]byte(e), &probe)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", probe.Type, e)
	}
}

func extract(c *client) (*ExtractResult, error) {
	return c.ExtractJSON(context.Background(), ExtractRequest{
		APIKey: "sk-test", Model: "claude-sonnet-5-5", PDF: []byte("%PDF-1.7"),
		Instruction: "go", Schema: json.RawMessage(`{"type":"object"}`), MaxTokens: 100,
		System: []SystemBlock{{Text: "rules", Cache: true}},
	})
}

func TestExtractRetriesBusyThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("missing api key header")
		}
		if !strings.Contains(r.Header.Get("Anthropic-Beta"), "server-side-fallback") {
			t.Errorf("missing fallback beta header: %q", r.Header.Get("Anthropic-Beta"))
		}
		switch calls.Add(1) {
		case 1:
			w.WriteHeader(529)
			_, _ = w.Write([]byte(`{"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`))
		case 2:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`))
		default:
			sse(w, `{"ok":true}`, "end_turn")
		}
	}))
	defer srv.Close()

	res, err := extract(testClient(srv.URL))
	if err != nil {
		t.Fatalf("ExtractJSON: %v", err)
	}
	if string(res.JSON) != `{"ok":true}` || res.Usage.OutputTokens != 42 || res.Usage.CacheReadInputTokens != 800 {
		t.Fatalf("result = %s %+v", res.JSON, res.Usage)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
	}
}

func TestExtractGivesUpWhenStillBusy(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(529)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`))
	}))
	defer srv.Close()
	if _, err := extract(testClient(srv.URL)); !errors.Is(err, ErrBusy) {
		t.Fatalf("err = %v, want ErrBusy", err)
	}
	if calls.Load() != 3 { // 1 + MaxRetries
		t.Fatalf("calls = %d, want 3", calls.Load())
	}
}

func TestExtractDoesNotRetryOtherErrors(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrInvalidAPIKey},
		{http.StatusNotFound, ErrModelNotFound},
		{http.StatusInternalServerError, nil},
	}
	for _, tc := range cases {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"type":"error","error":{"type":"x","message":"boom"}}`))
		}))
		_, err := extract(testClient(srv.URL))
		srv.Close()
		if tc.want != nil && !errors.Is(err, tc.want) {
			t.Errorf("status %d: err = %v, want %v", tc.status, err, tc.want)
		}
		var re *RequestError
		if tc.want == nil && (!errors.As(err, &re) || re.Status != tc.status) {
			t.Errorf("status %d: err = %v, want RequestError", tc.status, err)
		}
		if calls.Load() != 1 {
			t.Errorf("status %d: calls = %d, want 1 (no retry)", tc.status, calls.Load())
		}
	}
}

func TestExtractStopReasons(t *testing.T) {
	for stop, want := range map[string]error{"refusal": ErrRefused, "max_tokens": ErrTruncated} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { sse(w, `{"partial":`, stop) }))
		_, err := extract(testClient(srv.URL))
		srv.Close()
		if !errors.Is(err, want) {
			t.Errorf("%s: err = %v, want %v", stop, err, want)
		}
	}
}

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"claude-sonnet-5-5","type":"model","display_name":"Claude Sonnet 5.5","created_at":"2026-01-01T00:00:00Z"}`))
	}))
	defer srv.Close()
	c := testClient(srv.URL)
	if err := c.Ping(context.Background(), "good", "claude-sonnet-5-5"); err != nil {
		t.Fatalf("Ping good key: %v", err)
	}
	if err := c.Ping(context.Background(), "bad", "claude-sonnet-5-5"); !errors.Is(err, ErrInvalidAPIKey) {
		t.Fatalf("Ping bad key: err = %v, want ErrInvalidAPIKey", err)
	}
}
