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

export type LineBookingCreateRequest = {
  liffAccessToken: string
  lineUserId: string
  lineDisplayName: string
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
