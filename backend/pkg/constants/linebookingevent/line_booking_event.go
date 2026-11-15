package linebookingevent

// Actions in a LINE booking's history (line_booking_event.action).
// MUST match frontend/backoffice/src/constants/lineBookingEvents.ts.
const (
	Created           = "created"            // entered by sales
	Confirmed         = "confirmed"          // the sales owner confirmed the payment; target = finance user
	Cancelled         = "cancelled"          // the sales owner cancelled a new booking
	DocumentsSent     = "documents_sent"     // finance sent the documents; target = document codes
	DocumentsUpdated  = "documents_updated"  // finance replaced files before the review; target = document codes
	DocumentsApproved = "documents_approved" // the sales owner approved; target = LINE chat it was pushed to ("" = not pushed)
	DocumentsRejected = "documents_rejected" // sent back to finance; note = the fix request
	LineContactSet    = "line_contact_set"   // the customer's LINE chat linked / changed / unlinked; changes
	StaffNoteUpdated  = "staff_note_updated" // changes
	Edited            = "edited"             // the deal edited on the booking entry page; changes
	BalanceSubmitted  = "balance_submitted"  // a balance slip recorded on the booking entry page
	BalanceConfirmed  = "balance_confirmed"  // the sales owner confirmed the balance; target = finance user
	Deleted           = "deleted"
	Restored          = "restored"
	Comment           = "comment" // a staff comment; note = the text, mentions = the @mentioned users
)

// Kinds of line_booking_notification.
const (
	NotifyMention = "mention" // the recipient was @mentioned
	NotifyComment = "comment" // a comment on a booking the recipient owns or handles as finance
)
