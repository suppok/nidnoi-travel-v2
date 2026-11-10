package linebookingstate

// Sales workflow states of a LINE booking (line_booking.state).
// MUST match frontend/backoffice/src/constants/lineBookingStates.ts.
const (
	New       = 10 // just submitted from LIFF; nobody has reviewed it yet
	Confirmed = 20 // staff verified payment and tour
	Cancelled = 30 // rejected or customer backed out
)

var All = []int{New, Confirmed, Cancelled}
