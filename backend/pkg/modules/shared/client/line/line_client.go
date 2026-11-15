// Package line verifies Messaging API webhooks, reads OA friends' profiles, pushes messages to them
// and redeems LINE Login codes for the backoffice.
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
	botProfileURL = "https://api.line.me/v2/bot/profile/"
	// SignatureHeader carries the webhook body's HMAC-SHA256, keyed with the channel secret.
	SignatureHeader = "X-Line-Signature"
	requestTimeout  = 10 * time.Second
)

// TokenInfo is the LINE account a LINE Login code belongs to.
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
	// VerifySignature reports whether signature is the webhook body's signature for our channel secret.
	VerifySignature(body []byte, signature string) bool
	// GetProfile reads an OA friend's profile. ErrProfileNotFound when they are not a friend.
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	// WebhookConfigured reports whether the Messaging API channel secret is set.
	WebhookConfigured() bool
	// PushMessage sends Messaging API messages to one OA friend (uses the OA's message quota).
	// ErrNotConfigured without the channel access token, ErrPushQuota when the quota is used up.
	PushMessage(ctx context.Context, to string, messages []any, retryKey string) error

	// LoginConfigured reports whether backoffice LINE Login can run (channel ID and secret set).
	LoginConfigured() bool
	// LoginChannelID is the LINE Login channel ID for the browser's authorize URL.
	LoginChannelID() string
	// ExchangeLoginCode redeems a LINE Login code and returns the verified LINE user.
	// ErrLoginRejected when LINE refuses it, ErrNotConfigured without the channel secret.
	ExchangeLoginCode(ctx context.Context, code, redirectURI, codeVerifier, nonce string) (*TokenInfo, error)
}

type client struct {
	loginChannelID     string
	channelSecret      string
	channelAccessToken string
	loginChannelSecret string
	http               *http.Client
}

// New returns a LINE client. Methods whose credentials are empty return ErrNotConfigured (or false).
func New(cfg config.LINEConfig) Client {
	return &client{
		loginChannelID:     cfg.LoginChannelID,
		channelSecret:      cfg.ChannelSecret,
		channelAccessToken: cfg.ChannelAccessToken,
		loginChannelSecret: cfg.LoginChannelSecret,
		http:               &http.Client{Timeout: requestTimeout},
	}
}

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
