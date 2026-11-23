// Package emailoutboxstate holds the states and types of email_outbox rows.
package emailoutboxstate

// States. MUST match the CHECK constraint on email_outbox.state.
const (
	Pending = 10 // waiting for the sender (first try or a retry)
	Sent    = 20 // accepted by the provider
	Failed  = 30 // gave up: permanent rejection or max attempts reached
)

// Types (email_outbox.type). One email of each type per booking.
const (
	// TypeWholesaleBooking is the booking confirmation sent to the tour's wholesale when sales save a
	// LINE booking on the entry page.
	TypeWholesaleBooking = "wholesale_booking_confirm"
)
