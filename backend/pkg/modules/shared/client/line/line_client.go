// Package line verifies LIFF access tokens against the LINE Platform.
package line

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrNotConfigured is returned when LINE credentials are not set.
var ErrNotConfigured = errors.New("LINE is not configured")

const (
	verifyURL      = "https://api.line.me/oauth2/v2.1/verify"
	profileURL     = "https://api.line.me/v2/profile"
	requestTimeout = 10 * time.Second
)

// TokenInfo is the result of a successful LIFF access token verification.
type TokenInfo struct {
	UserID      string
	DisplayName string
}

// Client talks to the LINE Platform API.
type Client interface {
	// VerifyAccessToken validates a LIFF access token and returns the user's profile.
	// It checks that the token's client_id matches the configured LIFF channel.
	VerifyAccessToken(ctx context.Context, accessToken string) (*TokenInfo, error)
	Configured() bool
}

type client struct {
	liffChannelID string
	http          *http.Client
}

// New returns a LINE client. An empty channelID gives a client whose methods return ErrNotConfigured.
func New(liffChannelID string) Client {
	return &client{liffChannelID: liffChannelID, http: &http.Client{Timeout: requestTimeout}}
}

func (c *client) Configured() bool { return c.liffChannelID != "" }

// verifyResponse is the JSON returned by GET /oauth2/v2.1/verify.
type verifyResponse struct {
	ClientID  string `json:"client_id"`
	ExpiresIn int64  `json:"expires_in"`
	Scope     string `json:"scope"`
}

// profileResponse is the JSON returned by GET /v2/profile.
type profileResponse struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

func (c *client) VerifyAccessToken(ctx context.Context, accessToken string) (*TokenInfo, error) {
	if c.liffChannelID == "" {
		return nil, ErrNotConfigured
	}

	// 1. Verify the token is valid and belongs to our LIFF channel.
	vReq, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL+"?access_token="+accessToken, nil)
	if err != nil {
		return nil, errors.New("line: invalid verify request")
	}
	vRes, err := c.http.Do(vReq)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("line: %w", ctx.Err())
		}
		return nil, errors.New("line: verify request failed")
	}
	defer vRes.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(vRes.Body, 4096))
	if vRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("line: token verification failed (status %d)", vRes.StatusCode)
	}

	var vr verifyResponse
	if err := json.Unmarshal(body, &vr); err != nil {
		return nil, errors.New("line: invalid verify response")
	}
	if vr.ClientID != c.liffChannelID {
		return nil, fmt.Errorf("line: token channel %s does not match configured %s", vr.ClientID, c.liffChannelID)
	}

	// 2. Get the user profile.
	pReq, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, errors.New("line: invalid profile request")
	}
	pReq.Header.Set("Authorization", "Bearer "+accessToken)
	pRes, err := c.http.Do(pReq)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("line: %w", ctx.Err())
		}
		return nil, errors.New("line: profile request failed")
	}
	defer pRes.Body.Close()

	pBody, _ := io.ReadAll(io.LimitReader(pRes.Body, 4096))
	if pRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("line: profile fetch failed (status %d)", pRes.StatusCode)
	}

	var pr profileResponse
	if err := json.Unmarshal(pBody, &pr); err != nil {
		return nil, errors.New("line: invalid profile response")
	}

	return &TokenInfo{UserID: pr.UserID, DisplayName: pr.DisplayName}, nil
}
