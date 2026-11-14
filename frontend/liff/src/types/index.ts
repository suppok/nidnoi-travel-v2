// GET /tour-programs/public — selling and closed tours (closed ones have only sold-out periods).
export type TourPublicItem = {
  code: string
  name: string
  cardImage: string | null
  countries: { slug: string; nameTh: string }[]
  startingPrice: number | null
}

export type TourPeriodItem = {
  id: string
  startDate: string
  endDate: string
  price: number
  priceBeforeDiscount: number | null
  singleSupplement: number | null
  childWithBedPrice: number | null
  childNoBedPrice: number | null
  deposit: number | null // per person
  seatsAvailable: number
  // True when no seat is left or the tour is closed.
  soldOut: boolean
}

// GET /liff/me — the backoffice user linked to this LINE account.
export type LIFFMe = {
  name: string
  email: string
  roles: string[]
  lineDisplayName: string | null
}

// POST /liff/line-contacts/list — a customer known from LINE OA chats.
export type LineContact = {
  id: string
  displayName: string
  pictureUrl: string | null
  // False once the customer blocked the OA: a LINE push would not reach them.
  following: boolean
  lastMessageAt: string | null
}

// GET /liff/sales-users — sales owner options; defaultId is the caller when they are sales.
export type SalesUser = { id: string; name: string }
export type SalesUserList = { items: SalesUser[]; defaultId: string }

// GET /liff/booking-sources — where the customer booked (LINE OA, Facebook, ...).
export type BookingSource = {
  id: string
  name: string
  // Label of the customer's name on that source ("ชื่อ <shortName>"); "" = none.
  shortName: string
}

// MUST match backend pkg/constants/linebookingpayment.
export type PaymentMethod = 'transfer' | 'credit'
export type ChargeKind = 'visa' | 'foreignPassport' | 'joinLand' | 'fuelTax' | 'airportTax' | 'option'

export type LineBookingCharge = {
  kind: ChargeKind
  name: string // only for "option"
  count: number
  price: number
}

export type LineBookingCreateRequest = {
  // The sales person who owns the booking.
  salesUserId: string
  // The customer's LINE OA chat; null when they chat on a personal LINE account.
  lineContactId: string | null
  bookingSourceId: string | null
  refName: string
  refBookingNo: string
  tourCode: string
  periodId: string
  adultTwin: number
  adultSingle: number
  childWithBed: number
  childNoBed: number
  infant: number
  // Unit prices agreed with the customer; null = the period's price.
  adultPrice: number | null
  singleSupplement: number | null
  childWithBedPrice: number | null
  childNoBedPrice: number | null
  infantPrice: number | null
  adultPriceBeforeDiscount: number | null
  discountPerPerson: number
  charges: LineBookingCharge[]
  paymentMethod: PaymentMethod
  creditFeePercent: number
  isDeposit: boolean
  depositPerPerson: number
  depositDueDate: string | null // YYYY-MM-DD
  depositDueTime: string // HH:mm or ''
  remainingDueDate: string | null
  fullDueDate: string | null
  fullDueTime: string
  customerName: string
  customerPhone: string
  customerAddress: string
  slipImage: string
  note: string
}

export type LineBookingCreateResponse = {
  bookingNo: string
  totalPrice: number
}

export type MediaUploadResponse = {
  url: string
}

// GET /liff/line-bookings/:bookingNo — a paid booking (deposit or in full) reopened by its booking no.
// Room types are the backend's; dealFromBooking maps them back.
export type LineBookingItem = {
  id: string
  bookingNo: string
  state: number // 20 deposit paid (the balance may still be recorded), 25 paid in full
  tourCode: string
  tourName: string
  tourProgramPeriodId: string
  startDate: string
  endDate: string
  adultTwinCount: number
  adultTwinPrice: number
  adultSingleCount: number
  adultSinglePrice: number | null
  childWithBedCount: number
  childWithBedPrice: number | null
  childNoBedCount: number
  childNoBedPrice: number | null
  infantCount: number
  infantPrice: number | null
  adultPriceBeforeDiscount: number | null
  discountPerPerson: number
  charges: LineBookingCharge[]
  paymentMethod: PaymentMethod
  creditFeePercent: number
  cardFee: number
  isDeposit: boolean
  depositPerPerson: number
  depositDueDate: string | null
  depositDueTime: string
  remainingDueDate: string | null
  fullDueDate: string | null
  fullDueTime: string
  totalPrice: number
  customerName: string
  customerPhone: string
  customerAddress: string
  refName: string
  refBookingNo: string
  bookingSource: BookingSource | null
  salesUser: SalesUser | null
  lineContact: { id: string; displayName: string; pictureUrl: string | null; following: boolean } | null
  slipImage: string
  note: string
  // The balance recorded earlier, still waiting for the sales owner (null = none yet).
  balanceSlipImage: string
  balanceSubmittedAt: string | null
  balanceConfirmedAt: string | null
  // The finance documents of the current round: sent by finance (documentsAt) and approved by the
  // sales owner (documentsApprovedAt). Not approved yet = finance still works on them.
  financeUser: SalesUser | null
  documentsAt: string | null
  documentsApprovedAt: string | null
}

// PUT /liff/line-bookings — the whole deal again; tour, period, first slip, sales owner and the
// payment plan stay.
export type LineBookingLIFFUpdateRequest = Omit<LineBookingCreateRequest, 'salesUserId' | 'tourCode' | 'periodId' | 'slipImage'> & {
  bookingNo: string
  // The balance slip, only for a deposit booking in state 20; '' = none.
  balanceSlipImage: string
}

export type LineBookingLIFFUpdateResponse = {
  bookingNo: string
  totalPrice: number
  // The balance waits for the sales owner to confirm it in the backoffice.
  balancePending: boolean
}

// GET /liff/line-bookings/:bookingNo/events — one action in the booking's history (oldest first).
export type LineBookingEvent = {
  id: string
  action: string // backend pkg/constants/linebookingevent
  changes: { field: string; from: string; to: string }[]
  target: string
  note: string
  createDate: string
  user: { id: string; name: string } | null
}
