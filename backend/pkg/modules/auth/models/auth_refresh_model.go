package models

import "time"

type AuthRefreshRequest struct {
	RefreshToken string `json:"-"` // from the cookie

	ClientIP  string `json:"-"`
	UserAgent string `json:"-"`
}

type AuthRefreshResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // seconds

	RefreshToken     string    `json:"-"`
	RefreshExpiresAt time.Time `json:"-"`
	Persistent       bool      `json:"-"`
}
