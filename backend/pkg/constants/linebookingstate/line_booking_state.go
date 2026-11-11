package linebookingstate

// Sales workflow states of a LINE booking (line_booking.state).
// MUST match frontend/backoffice/src/constants/lineBookingStates.ts.
const (
	New         = 10 // just submitted from LIFF; the sales owner has not confirmed it yet
	DepositPaid = 20 // confirmed by the sales owner, the customer paid a deposit (was "confirmed")
	PaidInFull  = 25 // confirmed by the sales owner, the customer paid the whole amount
	Cancelled   = 30 // rejected or customer backed out
)

var All = []int{New, DepositPaid, PaidInFull, Cancelled}

// Paid are the confirmed states: a finance user has to issue the receipt.
var Paid = []int{DepositPaid, PaidInFull}
