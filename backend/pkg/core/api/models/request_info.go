package coremodels

// UserRequestInfo identifies the authenticated staff user. Always built from the JWT and the
// database by jwtAuth.Auth — never from a request body.
type UserRequestInfo struct {
	UserID             string
	Email              string
	Name               string
	Roles              []string
	MustChangePassword bool
}

// RequestInfo is passed by ApiHandler.Do to every handler closure.
type RequestInfo struct {
	UserRequestInfo *UserRequestInfo // nil on public routes
	ClientIP        string
	UserAgent       string
}
