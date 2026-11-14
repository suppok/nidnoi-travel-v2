package linebookingevent

// Actions in a LINE booking's history (line_booking_event.action).
// MUST match frontend/backoffice/src/constants/lineBookingEvents.ts.
const (
	Created           = "created"            // submitted from LIFF
	Confirmed         = "confirmed"          // the sales owner confirmed the payment; target = finance user
	Cancelled         = "cancelled"          // the sales owner cancelled a new booking
	DocumentsSent     = "documents_sent"     // finance sent the documents; target = document codes
	DocumentsUpdated  = "documents_updated"  // finance replaced files before the review; target = document codes
	DocumentsApproved = "documents_approved" // the sales owner approved; target = LINE chat it was pushed to ("" = not pushed)
	DocumentsRejected = "documents_rejected" // sent back to finance; note = the fix request
	LineContactSet    = "line_contact_set"   // the customer's LINE chat linked / changed / unlinked; changes
	StaffNoteUpdated  = "staff_note_updated" // changes
	Edited            = "edited"             // the deal edited from LIFF; changes
	BalanceSubmitted  = "balance_submitted"  // a balance slip recorded from LIFF
	BalanceConfirmed  = "balance_confirmed"  // the sales owner confirmed the balance; target = finance user
	Deleted           = "deleted"
	Restored          = "restored"
)
