package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// resendEndpoint: https://resend.com/docs/api-reference/emails/send-email
const resendEndpoint = "https://api.resend.com/emails"

type resendClient struct {
	apiKey   string
	endpoint string
	http     *http.Client
}

// NewResend returns a Resend client; endpoint is overridable for tests.
func NewResend(apiKey, endpoint string, timeout time.Duration) Mailer {
	return &resendClient{apiKey: apiKey, endpoint: endpoint, http: &http.Client{Timeout: timeout}}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

type resendResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"` // error responses
}

func (c *resendClient) Send(ctx context.Context, msg Message) (string, error) {
	body, err := json.Marshal(resendRequest{
		From: msg.From, To: msg.To, Cc: msg.Cc, ReplyTo: msg.ReplyTo,
		Subject: msg.Subject, HTML: msg.HTML, Text: msg.Text,
	})
	if err != nil {
		return "", &SendError{Message: "encode request"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", &SendError{Message: "build request"}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if msg.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", msg.IdempotencyKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		// Never wrap err: it may carry request details. Network failures and timeouts are retried.
		reason := "request failed"
		if ctx.Err() != nil {
			reason = "request cancelled"
		}
		return "", &SendError{Retryable: true, Message: reason}
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	var out resendResponse
	_ = json.Unmarshal(raw, &out)

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		if out.ID == "" {
			return "", &SendError{Status: res.StatusCode, Message: "response without an id"}
		}
		return out.ID, nil
	}
	message := strings.TrimSpace(out.Message)
	if message == "" {
		message = http.StatusText(res.StatusCode)
	}
	if len(message) > 500 {
		message = message[:500]
	}
	retryable := res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500
	return "", &SendError{Retryable: retryable, Status: res.StatusCode, Message: message}
}
