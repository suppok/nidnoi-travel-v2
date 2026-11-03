// Package deployhook triggers a Cloudflare Workers Builds deploy hook (customer site rebuild).
package deployhook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrNotConfigured is returned when CLOUDFLARE_DEPLOY_HOOK_URL is not set.
var ErrNotConfigured = errors.New("deploy hook is not configured")

const requestTimeout = 15 * time.Second

type DeployHook interface {
	// Trigger asks Cloudflare to build and deploy the customer site once.
	Trigger(ctx context.Context) error
	Configured() bool
}

type client struct {
	url  string
	http *http.Client
}

// New returns a client for the hook URL; an empty url gives a client whose Trigger returns ErrNotConfigured.
func New(url string) DeployHook {
	return &client{url: url, http: &http.Client{Timeout: requestTimeout}}
}

func (c *client) Configured() bool { return c.url != "" }

// Trigger POSTs to the hook. The URL is a credential: it never appears in returned errors or logs.
func (c *client) Trigger(ctx context.Context) error {
	if c.url == "" {
		return ErrNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, nil)
	if err != nil {
		return errors.New("deploy hook: invalid request")
	}
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("deploy hook: %w", ctx.Err())
		}
		return errors.New("deploy hook: request failed") // net errors embed the URL
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("deploy hook: unexpected status %d", res.StatusCode)
	}
	return nil
}
