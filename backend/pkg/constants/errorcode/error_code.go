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

// Taxonomy (country, city)
var (
	SlugTaken  = ErrCode("SLUG_TAKEN", "Slug is already in use")
	RefIDTaken = ErrCode("REF_ID_TAKEN", "Ref ID is already in use")
)

// Wholesale
var (
	CodeTaken = ErrCode("CODE_TAKEN", "Code is already in use")
)

// Media upload
var (
	MediaTypeNotAllowed = ErrCode("MEDIA_TYPE_NOT_ALLOWED", "This file type is not allowed for this purpose")
	MediaTooLarge       = ErrCode("MEDIA_TOO_LARGE", "The file is too large")
	ImageTooSmall       = ErrCode("IMAGE_TOO_SMALL", "The image is smaller than the minimum size")
	StorageUnavailable  = ErrCode("STORAGE_UNAVAILABLE", "File storage is not configured")
)

// Booking
var (
	PeriodUnavailable = ErrCode("PERIOD_UNAVAILABLE", "This departure is no longer open for booking")
)

// LINE
var (
	LINENotConfigured    = ErrCode("LINE_NOT_CONFIGURED", "LINE integration is not configured")
	InvalidLINEToken     = ErrCode("INVALID_LINE_TOKEN", "Invalid or expired LINE access token")
	InvalidLINESignature = ErrCode("INVALID_LINE_SIGNATURE", "Invalid LINE webhook signature")
	LINENotLinked        = ErrCode("LINE_NOT_LINKED", "This LINE account is not linked to a staff user")
	LINELinkCodeInvalid  = ErrCode("LINE_LINK_CODE_INVALID", "The LINE link code is invalid or has expired")
	LINEAlreadyLinked    = ErrCode("LINE_ALREADY_LINKED", "This LINE account is already linked to another staff user")
)

// Customer site rebuild
var (
	SiteRebuildUnavailable = ErrCode("SITE_REBUILD_UNAVAILABLE", "The site rebuild hook is not configured")
	SiteRebuildFailed      = ErrCode("SITE_REBUILD_FAILED", "Could not start the site rebuild")
)
