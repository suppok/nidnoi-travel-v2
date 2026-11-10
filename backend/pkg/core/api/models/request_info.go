package coremodels

// UserRequestInfo identifies the authenticated staff user. Always built from the JWT (or, on LIFF
// routes, the verified LIFF access token) and the database by the auth middleware — never from a
// request body.
type UserRequestInfo struct {
	UserID             string
	Email              string
	Name               string
	Roles              []string
	MustChangePassword bool

	// Set on LIFF routes only: the LINE account the LIFF access token belongs to.
	LineUserID      string
	LineDisplayName string
}

// RequestInfo is passed by ApiHandler.Do to every handler closure.
type RequestInfo struct {
	UserRequestInfo *UserRequestInfo // nil on public routes
	ClientIP        string
	UserAgent       string
}
