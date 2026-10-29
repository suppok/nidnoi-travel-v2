package models

type AuthLogoutRequest struct {
	RefreshToken string `json:"-"` // from the cookie
}

// AuthLogoutResponse is always nil (data: null).
type AuthLogoutResponse struct{}
