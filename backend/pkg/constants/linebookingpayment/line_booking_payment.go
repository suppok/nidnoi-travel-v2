package linebookingpayment

// Payment methods and extra-charge kinds of a LINE booking (line_booking.payment_method, charges[].kind).
// MUST match frontend/liff/src/types/index.ts and frontend/backoffice/src/types/lineBooking.ts.
const (
	Transfer = "transfer"
	Credit   = "credit"
)

var Methods = []any{Transfer, Credit}

const (
	ChargeVisa            = "visa"
	ChargeForeignPassport = "foreignPassport"
	ChargeJoinLand        = "joinLand"
	ChargeFuelTax         = "fuelTax"
	ChargeAirportTax      = "airportTax"
	ChargeOption          = "option" // a named extra; the only kind with a name
)

var ChargeKinds = []any{ChargeVisa, ChargeForeignPassport, ChargeJoinLand, ChargeFuelTax, ChargeAirportTax, ChargeOption}
