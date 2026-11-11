// Package line verifies LIFF access tokens and Messaging API webhooks, reads OA friends' profiles
// and pushes messages to them.
package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"backend/pkg/core/config"
)

// ErrNotConfigured is returned when LINE credentials are not set.
var ErrNotConfigured = errors.New("LINE is not configured")

// ErrProfileNotFound is returned when the user is not (or no longer) a friend of the OA.
var ErrProfileNotFound = errors.New("line: profile not found")

const (
	verifyURL     = "https://api.line.me/oauth2/v2.1/verify"
	profileURL    = "https://api.line.me/v2/profile"
	botProfileURL = "https://api.line.me/v2/bot/profile/"
	// SignatureHeader carries the webhook body's HMAC-SHA256, keyed with the channel secret.
	SignatureHeader = "X-Line-Signature"
	requestTimeout  = 10 * time.Second
)

// TokenInfo is the result of a successful LIFF access token verification.
type TokenInfo struct {
	UserID      string
	DisplayName string
}

// Profile is an OA friend's profile from the Messaging API.
type Profile struct {
	UserID      string
	DisplayName string
	PictureURL  string // empty when the user has no picture
}

// Client talks to the LINE Platform API.
type Client interface {
	// VerifyAccessToken validates a LIFF access token and returns the user's profile.
	// It checks that the token's client_id matches the configured LIFF channel.
	VerifyAccessToken(ctx context.Context, accessToken string) (*TokenInfo, error)
	Configured() bool

	// VerifySignature reports whether signature is the webhook body's signature for our channel secret.
	VerifySignature(body []byte, signature string) bool
	// GetProfile reads an OA friend's profile. ErrProfileNotFound when they are not a friend.
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	// WebhookConfigured reports whether the Messaging API channel secret is set.
	WebhookConfigured() bool
	// PushMessage sends Messaging API messages to one OA friend (uses the OA's message quota).
	// ErrNotConfigured without the channel access token, ErrPushQuota when the quota is used up.
	PushMessage(ctx context.Context, to string, messages []any, retryKey string) error
}

type client struct {
	liffChannelID      string
	channelSecret      string
	channelAccessToken string
	http               *http.Client
}

// New returns a LINE client. Methods whose credentials are empty return ErrNotConfigured (or false).
func New(cfg config.LINEConfig) Client {
	return &client{
		liffChannelID:      cfg.LIFFChannelID,
		channelSecret:      cfg.ChannelSecret,
		channelAccessToken: cfg.ChannelAccessToken,
		http:               &http.Client{Timeout: requestTimeout},
	}
}

func (c *client) Configured() bool { return c.liffChannelID != "" }

func (c *client) WebhookConfigured() bool { return c.channelSecret != "" }

func (c *client) VerifySignature(body []byte, signature string) bool {
	if c.channelSecret == "" || signature == "" {
		return false
	}
	got, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(c.channelSecret))
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

func (c *client) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	if c.channelAccessToken == "" {
		return nil, ErrNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, botProfileURL+url.PathEscape(userID), nil)
	if err != nil {
		return nil, errors.New("line: invalid profile request")
	}
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("line: %w", ctx.Err())
		}
		return nil, errors.New("line: bot profile request failed")
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrProfileNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("line: bot profile fetch failed (status %d)", res.StatusCode)
	}

	var pr struct {
		UserID      string `json:"userId"`
		DisplayName string `json:"displayName"`
		PictureURL  string `json:"pictureUrl"`
	}
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, errors.New("line: invalid bot profile response")
	}
	return &Profile{UserID: pr.UserID, DisplayName: pr.DisplayName, PictureURL: pr.PictureURL}, nil
}

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
