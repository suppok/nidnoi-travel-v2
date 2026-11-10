package apicode

var (
	// LineBookingCreate is checked on the LIFF app's endpoints, for the staff user linked to the LINE account.
	LineBookingCreate = ApiCode("line-booking:create", "Submit LINE bookings from the LIFF app")
	LineBookingList   = ApiCode("line-booking:list", "List LINE bookings")
	LineBookingGet    = ApiCode("line-booking:get", "Get a LINE booking")
	LineBookingUpdate = ApiCode("line-booking:update", "Update a LINE booking's state and note")
	LineBookingStatus = ApiCode("line-booking:status", "Delete or restore a LINE booking")
)
