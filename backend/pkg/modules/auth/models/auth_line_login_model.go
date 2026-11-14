package models

// --- GET /auth/line/config: whether "ล็อกอินด้วย LINE" is available ---

type AuthLineLoginConfigResponse struct {
	Enabled bool `json:"enabled"`
	// ChannelID is the LINE Login channel for the authorize URL ("" when disabled).
	ChannelID string `json:"channelId"`
}

// --- POST /auth/line/login: finish LINE Login with the code LINE redirected back with ---

type AuthLineLoginRequest struct {
	Code string `json:"code"`
	// RedirectURI is the callback the authorize request used: <backoffice origin>/login/line/callback,
	// where the origin is one of CORS_ALLOWED_ORIGINS.
	RedirectURI  string `json:"redirectUri"`
	CodeVerifier string `json:"codeVerifier"` // PKCE
	Nonce        string `json:"nonce"`
	RememberMe   bool   `json:"rememberMe"`

	ClientIP  string `json:"-"`
	UserAgent string `json:"-"`
}
