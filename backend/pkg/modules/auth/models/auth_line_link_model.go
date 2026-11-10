package models

import (
	"time"

	"github.com/guregu/null/v5"

	coremodels "backend/pkg/core/api/models"
)

// AuthLineAccount is the LINE account a staff user linked; null on /auth/me until linked.
type AuthLineAccount struct {
	DisplayName string    `json:"displayName"`
	LinkedAt    time.Time `json:"linkedAt"`
}

// --- POST /auth/line-link (backoffice): start linking ---

type AuthLineLinkCreateRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AuthLineLinkCreateResponse struct {
	// URL opens the LIFF app with the one-time code; the backoffice shows it as a QR code.
	URL       string    `json:"url" example:"https://liff.line.me/1234567890-AbCdEfGh?lineLink=..."`
	ExpiresAt time.Time `json:"expiresAt"`
}

// --- POST /auth/line-unlink (backoffice) ---

type AuthLineUnlinkRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type AuthLineUnlinkResponse struct{}

// --- POST /liff/line-link (LIFF): finish linking ---

type LIFFLineLinkRequest struct {
	Code string `json:"code"`

	AccessToken string `json:"-"` // LIFF access token from the Authorization header
}

type LIFFLineLinkResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// --- GET /liff/me (LIFF) ---

type LIFFMeRequest struct {
	UserRequestInfo *coremodels.UserRequestInfo `json:"-"`
}

type LIFFMeResponse struct {
	Name            string      `json:"name"`
	Email           string      `json:"email"`
	Roles           []string    `json:"roles"`
	LineDisplayName null.String `json:"lineDisplayName" swaggertype:"string"`
}
