package line

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// LINE Login (web) for the backoffice: the browser is sent to LINE's authorize page with PKCE, LINE
// redirects back with a code, and the backend exchanges it here with the channel secret.

const tokenURL = "https://api.line.me/oauth2/v2.1/token"

// ErrLoginRejected is returned when LINE refuses the code or the ID token (expired, reused, wrong
// verifier / nonce / redirect URI).
var ErrLoginRejected = errors.New("line: login rejected")

type tokenResponse struct {
	IDToken string `json:"id_token"`
}

type idTokenClaims struct {
	Sub   string `json:"sub"`
	Name  string `json:"name"`
	Nonce string `json:"nonce"`
}

func (c *client) LoginConfigured() bool { return c.loginChannelID != "" && c.loginChannelSecret != "" }

// LoginChannelID is the LINE Login channel the backoffice's authorize URL uses (public).
func (c *client) LoginChannelID() string { return c.loginChannelID }

// ExchangeLoginCode trades an authorization code for the LINE user: it redeems the code (with the
// PKCE verifier and the channel secret), then has LINE verify the ID token for our channel and the
// nonce the browser sent.
func (c *client) ExchangeLoginCode(ctx context.Context, code, redirectURI, codeVerifier, nonce string) (*TokenInfo, error) {
	if !c.LoginConfigured() {
		return nil, ErrNotConfigured
	}

	var token tokenResponse
	if err := c.postForm(ctx, tokenURL, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {c.loginChannelID},
		"client_secret": {c.loginChannelSecret},
		"code_verifier": {codeVerifier},
	}, &token); err != nil {
		return nil, err
	}
	if token.IDToken == "" {
		return nil, ErrLoginRejected
	}

	var claims idTokenClaims
	if err := c.postForm(ctx, verifyURL, url.Values{
		"id_token":  {token.IDToken},
		"client_id": {c.loginChannelID},
		"nonce":     {nonce},
	}, &claims); err != nil {
		return nil, err
	}
	if claims.Sub == "" || claims.Nonce != nonce {
		return nil, ErrLoginRejected
	}
	return &TokenInfo{UserID: claims.Sub, DisplayName: claims.Name}, nil
}

// postForm POSTs a form to LINE and decodes the JSON answer. A 4xx is ErrLoginRejected.
func (c *client) postForm(ctx context.Context, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return errors.New("line: invalid login request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("line: %w", ctx.Err())
		}
		return errors.New("line: login request failed")
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64*1024))
	switch {
	case res.StatusCode >= 400 && res.StatusCode < 500:
		return ErrLoginRejected
	case res.StatusCode != http.StatusOK:
		return fmt.Errorf("line: login request failed (status %d)", res.StatusCode)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return errors.New("line: invalid login response")
	}
	return nil
}
