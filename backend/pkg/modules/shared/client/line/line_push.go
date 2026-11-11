package line

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const pushURL = "https://api.line.me/v2/bot/message/push"

// ErrPushQuota is returned when the OA has used up its monthly message quota (HTTP 429).
var ErrPushQuota = errors.New("line: monthly message quota reached")

// PushMessage sends messages (Messaging API message objects) to one user. retryKey (a UUID) makes
// the push idempotent: LINE drops a second request with the same key within 24 hours.
func (c *client) PushMessage(ctx context.Context, to string, messages []any, retryKey string) error {
	if c.channelAccessToken == "" {
		return ErrNotConfigured
	}
	body, err := json.Marshal(map[string]any{"to": to, "messages": messages})
	if err != nil {
		return fmt.Errorf("line: encode push: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pushURL, bytes.NewReader(body))
	if err != nil {
		return errors.New("line: invalid push request")
	}
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", "application/json")
	if retryKey != "" {
		req.Header.Set("X-Line-Retry-Key", retryKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("line: %w", ctx.Err())
		}
		return errors.New("line: push request failed")
	}
	defer res.Body.Close()
	detail, _ := io.ReadAll(io.LimitReader(res.Body, 1024))

	switch {
	case res.StatusCode == http.StatusOK:
		return nil
	// 409: a request with this retry key was already accepted — the message went out.
	case res.StatusCode == http.StatusConflict && retryKey != "":
		return nil
	case res.StatusCode == http.StatusTooManyRequests:
		return ErrPushQuota
	default:
		return fmt.Errorf("line: push failed (status %d): %s", res.StatusCode, detail)
	}
}
