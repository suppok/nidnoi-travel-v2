package apicode

var (
	// LineBookingCreate is checked on the LIFF app's endpoints, for the staff user linked to the LINE account.
	LineBookingCreate = ApiCode("line-booking:create", "Submit LINE bookings from the LIFF app")
	LineBookingList   = ApiCode("line-booking:list", "List LINE bookings")
	LineBookingGet    = ApiCode("line-booking:get", "Get a LINE booking")
	LineBookingUpdate = ApiCode("line-booking:update", "Confirm or cancel a LINE booking, edit its note")
	LineBookingStatus = ApiCode("line-booking:status", "Delete or restore a LINE booking")
	// LineBookingDocument: finance uploads and sends the AI/IV/RE PDFs of a booking sent to them.
	LineBookingDocument = ApiCode("line-booking:document", "Send the finance documents of a confirmed LINE booking")
)
