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
