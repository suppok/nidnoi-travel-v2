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
  singleSupplement: number | null
  childWithBedPrice: number | null
  childNoBedPrice: number | null
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

export type LineBookingCreateRequest = {
  // The customer's LINE OA chat; null when they chat on a personal LINE account.
  lineContactId: string | null
  tourCode: string
  periodId: string
  adultTwin: number
  adultSingle: number
  childWithBed: number
  childNoBed: number
  customerName: string
  customerPhone: string
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
