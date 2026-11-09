package bookingstate

// Sales workflow states of a booking request (booking.state). Gaps leave room for new states.
// MUST match frontend/backoffice/src/constants/bookingStates.ts.
const (
	New       = 10 // just submitted on the customer site; nobody has contacted the customer yet
	Contacted = 20 // sales has reached the customer
	Confirmed = 30 // the customer confirmed (seats and payment are handled outside this system)
	Cancelled = 40 // the customer declined, could not be reached, or the request was invalid
)

var All = []int{New, Contacted, Confirmed, Cancelled}

// Booking numbers are the 2-digit year (Thailand time, Gregorian) followed by NoLength random
// characters from NoAlphabet, e.g. 26X8K9AQ: the reference a customer quotes to staff. The alphabet
// leaves out 0/O and 1/I/L, which are easy to misread over the phone or LINE (31 characters, about
// 887 million numbers per year).
const (
	NoLength   = 6
	NoAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
)

// SystemUserID is create_by for bookings submitted on the customer site (no staff user).
const SystemUserID = "00000000-0000-0000-0000-000000000000"
