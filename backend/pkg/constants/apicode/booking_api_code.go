package apicode

var (
	BookingList   = ApiCode("booking:list", "List booking requests from the customer site")
	BookingGet    = ApiCode("booking:get", "Get a booking request")
	BookingUpdate = ApiCode("booking:update", "Update a booking request's sales state and note")
	BookingStatus = ApiCode("booking:status", "Delete or restore a booking request")
)
