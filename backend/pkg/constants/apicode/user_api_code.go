package apicode

var (
	UserCreate        = ApiCode("user:create", "Create a staff user")
	UserUpdate        = ApiCode("user:update", "Update a staff user's profile and roles")
	UserList          = ApiCode("user:list", "List staff users")
	UserGet           = ApiCode("user:get", "Get a staff user")
	UserStatus        = ApiCode("user:status", "Deactivate or restore a staff user")
	UserPasswordReset = ApiCode("user:password-reset", "Set a temporary password for a staff user")
)
