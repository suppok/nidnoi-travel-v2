export type TourPublicItem = {
  id: string
  code: string
  name: string
  coverImage: string
  countries: { slug: string; nameTh: string }[]
  startingPrice: number | null
  isSelling: boolean
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
  status: boolean
}

// GET /liff/me — the backoffice user linked to this LINE account.
export type LIFFMe = {
  name: string
  email: string
  roles: string[]
  lineDisplayName: string | null
}

export type LineBookingCreateRequest = {
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
