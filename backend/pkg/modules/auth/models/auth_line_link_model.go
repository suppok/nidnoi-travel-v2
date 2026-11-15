package models

import (
	"time"

	coremodels "backend/pkg/core/api/models"
)

// AuthLineAccount is the LINE account a staff user linked; null on /auth/me until linked.
type AuthLineAccount struct {
	DisplayName string    `json:"displayName"`
	LinkedAt    time.Time `json:"linkedAt"`
}

// --- POST /auth/line-link (backoffice): link the caller's LINE account with a LINE Login code ---

// AuthLineLinkRequest carries the LINE Login callback, like AuthLineLoginRequest.
type AuthLineLinkRequest struct {
	Code string `json:"code"`
	// RedirectURI is <backoffice origin>/login/line/callback, the origin one of CORS_ALLOWED_ORIGINS.
	RedirectURI  string `json:"redirectUri"`
	CodeVerifier string `json:"codeVerifier"` // PKCE
	Nonce        string `json:"nonce"`

	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AuthLineLinkResponse struct {
	Line AuthLineAccount `json:"line"`
}

// --- POST /auth/line-unlink (backoffice) ---

type AuthLineUnlinkRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AuthLineUnlinkResponse struct{}
