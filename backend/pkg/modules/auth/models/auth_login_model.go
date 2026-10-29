package models

import "time"

type AuthLoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`

	ClientIP  string `json:"-"`
	UserAgent string `json:"-"`
}

type AuthLoginResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // seconds

	// Set as the HttpOnly refresh cookie by the handler; never serialized.
	RefreshToken     string    `json:"-"`
	RefreshExpiresAt time.Time `json:"-"`
	Persistent       bool      `json:"-"`
}
