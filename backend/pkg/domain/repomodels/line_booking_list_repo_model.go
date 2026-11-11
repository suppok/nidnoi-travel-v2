package repomodels

// LineBookingScope is what a user who can't see every LINE booking sees: the bookings they own as
// sales (when Sales) or were sent as finance (when Finance). Neither = nothing.
type LineBookingScope struct {
	UserID  string
	Sales   bool
	Finance bool
}

// LineBookingListCriteria filters POST /line-bookings/list.
type LineBookingListCriteria struct {
	Keyword string // matches booking no, customer name, customer phone, tour code or LINE display name
	State   int    // 0 = any
	// Paid limits the list to confirmed bookings (linebookingstate.Paid), newest confirmation first.
	Paid   bool
	Status *bool // nil = any
	// Scope limits the list to the bookings the caller may see; nil = every booking.
	Scope *LineBookingScope
	// FinanceUserID limits the list to bookings sent to this finance user; "" = any.
	FinanceUserID string
	// SalesUserID limits the list to bookings owned by this sales user; "" = any.
	SalesUserID string
	// DocumentsDone: true = finance sent the documents (newest first), false = still waiting, nil = any.
	DocumentsDone *bool
	// DocumentsApproved: true = the sales owner approved them, false = not (yet), nil = any.
	DocumentsApproved *bool
	Page              int
	PageSize          int
}
