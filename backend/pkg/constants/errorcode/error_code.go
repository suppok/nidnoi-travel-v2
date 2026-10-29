package errorcode

// Code is a business error code. Frontends branch on Code — never rename a shipped one.
type Code struct {
	Code    string
	Message string
}

func ErrCode(code, message string) Code {
	return Code{Code: code, Message: message}
}

// Generic codes (cerrors).
var (
	Validation       = ErrCode("VALIDATION", "Validation failed")
	Unauthenticated  = ErrCode("UNAUTHENTICATED", "Authentication required")
	PermissionDenied = ErrCode("PERMISSION_DENIED", "Permission denied")
	NotFound         = ErrCode("NOT_FOUND", "Not found")
	Conflict         = ErrCode("CONFLICT", "Conflict")
	RateLimited      = ErrCode("RATE_LIMITED", "Too many requests")
	Internal         = ErrCode("INTERNAL", "Internal server error")
)

// Auth
var (
	InvalidCredentials     = ErrCode("INVALID_CREDENTIALS", "Invalid email or password")
	InvalidPassword        = ErrCode("INVALID_PASSWORD", "Current password is incorrect")
	PasswordChangeRequired = ErrCode("PASSWORD_CHANGE_REQUIRED", "Password change required")
)

// User
var (
	EmailTaken          = ErrCode("EMAIL_TAKEN", "Email is already in use")
	SuperAdminProtected = ErrCode("SUPER_ADMIN_PROTECTED", "The super admin can only be changed by themselves")
	SuperAdminRequired  = ErrCode("SUPER_ADMIN_REQUIRED", "Only the super admin can grant the admin role or manage admins")
	CannotModifySelf    = ErrCode("CANNOT_MODIFY_SELF", "You cannot change your own roles or status")
)
