package apicode

var (
	AirlineCreate = ApiCode("airline:create", "Create an airline")
	AirlineUpdate = ApiCode("airline:update", "Update an airline")
	AirlineList   = ApiCode("airline:list", "List airlines")
	AirlineGet    = ApiCode("airline:get", "Get an airline")
	AirlineStatus = ApiCode("airline:status", "Delete or restore an airline")
)
