package wholesalestate

// Workflow states of a wholesale (wholesale.state). Gaps leave room for new states.
// MUST match frontend/backoffice/src/constants/wholesaleStates.ts.
const (
	OnRequest = 10
	Connected = 20
	Blocked   = 30
	Declined  = 40
)

var All = []int{OnRequest, Connected, Blocked, Declined}
