package tourprogramstate

// Workflow states of a tour program (tour_program.state). Gaps leave room for new states.
// MUST match frontend/backoffice/src/constants/tourProgramStates.ts.
const (
	Draft     = 10 // being prepared, hidden from the customer site
	Selling   = 20 // public and open for inquiries
	Closed    = 30 // public, shown as closed / sold out
	Cancelled = 40 // hidden from the customer site
)

var All = []int{Draft, Selling, Closed, Cancelled}

// Public lists the states the customer site shows.
var Public = []int{Selling, Closed}

// IsPublic reports whether the customer site shows a tour in this state.
func IsPublic(state int) bool {
	for _, s := range Public {
		if s == state {
			return true
		}
	}
	return false
}
