// The deal of a LINE booking, with the same rows and math as the backoffice "สรุปยอดโอน" summary
// (frontend/backoffice/src/features/sales-tool/booking-summary/summary.ts) and the backend quote
// (backend/pkg/modules/linebooking/usecase/line_booking_create_usecase.go quoteLine). Keep all three
// in sync: the backend recomputes the total and it should match what the form shows.
//
// Rows follow the summary (adult, child, infant, single room per room); `dealRequest` maps them to
// the backend's room types: adults in single rooms = rooms, single supplement = the room price,
// children = "with bed".
import type { ChargeKind, LineBookingCharge, LineBookingCreateRequest, PaymentMethod, TourPeriodItem } from '../types/index.ts'
import type { ParsedSummary, PricedLine } from './parseSummary.ts'

// A count or price; null = not entered.
export type Line = { count: number | null; price: number | null }

export type FixedChargeKind = Exclude<ChargeKind, 'option'>
export type OptionCharge = Line & { name: string }

export type Payment = {
  method: PaymentMethod
  creditFeePercent: number | null
  isDeposit: boolean
  depositPerPerson: number | null
  depositDueDate: string // YYYY-MM-DD or ''
  depositDueTime: string // HH:mm or ''
  remainingDueDate: string
  fullDueDate: string
  fullDueTime: string
}

export type Deal = {
  adult: Line
  child: Line // under 20; booked as "with bed"
  infant: Line
  singleRoom: Line // count = rooms, price per room
  adultPriceBeforeDiscount: number | null // shown to the customer as "(จาก …)"
  discountPerPerson: number | null // off adult, child and infant prices
  // A fuel/airport tax count of null follows the number of travellers.
  charges: Record<FixedChargeKind, Line>
  options: OptionCharge[]
  payment: Payment
}

// Rows of the price table, in the summary tool's order.
export const DEAL_ROWS: { key: 'adult' | 'child' | 'infant' | FixedChargeKind | 'singleRoom'; label: string; perTraveller?: boolean }[] = [
  { key: 'adult', label: 'ผู้ใหญ่' },
  { key: 'child', label: 'เด็กอายุต่ำกว่า 20 ปี' },
  { key: 'infant', label: 'ทารก' },
  { key: 'visa', label: 'วีซ่า' },
  { key: 'foreignPassport', label: 'พาสปอร์ตต่างชาติ' },
  { key: 'singleRoom', label: 'ห้องพักเดี่ยว (ห้อง)' },
  { key: 'joinLand', label: 'Join land' },
  { key: 'fuelTax', label: 'ภาษีน้ำมัน', perTraveller: true },
  { key: 'airportTax', label: 'ภาษีสนามบิน', perTraveller: true },
]

const FIXED_KINDS: FixedChargeKind[] = ['visa', 'foreignPassport', 'joinLand', 'fuelTax', 'airportTax']

export const DEFAULT_CREDIT_FEE_PERCENT = 2.6

const emptyLine = (): Line => ({ count: null, price: null })

export const emptyDeal = (): Deal => ({
  adult: emptyLine(),
  child: emptyLine(),
  infant: emptyLine(),
  singleRoom: emptyLine(),
  adultPriceBeforeDiscount: null,
  discountPerPerson: null,
  charges: { visa: emptyLine(), foreignPassport: emptyLine(), joinLand: emptyLine(), fuelTax: emptyLine(), airportTax: emptyLine() },
  options: [],
  payment: {
    method: 'transfer',
    creditFeePercent: DEFAULT_CREDIT_FEE_PERCENT,
    isDeposit: false,
    depositPerPerson: null,
    depositDueDate: '',
    depositDueTime: '',
    remainingDueDate: '',
    fullDueDate: '',
    fullDueTime: '',
  },
})

const n = (v: number | null | undefined) => (typeof v === 'number' && Number.isFinite(v) ? v : 0)

export function travellers(d: Deal): number {
  return n(d.adult.count) + n(d.child.count) + n(d.infant.count)
}

// The charge lines sent to the backend: counts resolved, unused lines dropped.
export function chargeLines(deal: Deal): LineBookingCharge[] {
  const pax = travellers(deal)
  const lines: LineBookingCharge[] = []
  for (const kind of FIXED_KINDS) {
    const c = deal.charges[kind]
    const perTraveller = kind === 'fuelTax' || kind === 'airportTax'
    const count = perTraveller && c.count === null ? pax : n(c.count)
    if (count > 0 && n(c.price) > 0) lines.push({ kind, name: '', count, price: n(c.price) })
  }
  for (const o of deal.options) {
    if (n(o.count) > 0 && n(o.price) > 0) lines.push({ kind: 'option', name: o.name.trim() || 'Option', count: n(o.count), price: n(o.price) })
  }
  return lines
}

export type Totals = {
  travellers: number
  subtotal: number // before the card fee
  fee: number
  deposit: number // before the card fee
  depositToPay: number // deposit + card fee
  remaining: number
  grandTotal: number
}

// Card fee rounded up to the next baht, in basis points like the backend.
function cardFee(amount: number, percent: number): number {
  return Math.ceil((amount * Math.round(percent * 100)) / 10000)
}

const lineTotal = (l: Line) => n(l.count) * n(l.price)

export function computeTotals(deal: Deal): Totals {
  const { payment } = deal
  const pax = travellers(deal)
  const subtotal =
    lineTotal(deal.adult) +
    lineTotal(deal.child) +
    lineTotal(deal.infant) +
    lineTotal(deal.singleRoom) -
    n(deal.discountPerPerson) * pax +
    chargeLines(deal).reduce((sum, l) => sum + l.count * l.price, 0)

  const deposit = payment.isDeposit ? n(payment.depositPerPerson) * pax : 0
  const fee = payment.method === 'credit' ? cardFee(payment.isDeposit ? deposit : subtotal, n(payment.creditFeePercent)) : 0
  return {
    travellers: pax,
    subtotal,
    fee,
    deposit,
    depositToPay: payment.isDeposit ? deposit + fee : 0,
    remaining: payment.isDeposit ? subtotal - deposit : 0,
    grandTotal: subtotal + fee,
  }
}

const DAY_MS = 24 * 60 * 60 * 1000

function addDays(date: string, days: number): string {
  const d = new Date(`${date}T00:00:00Z`)
  return new Date(d.getTime() + days * DAY_MS).toISOString().slice(0, 10)
}

function todayInThailand(): string {
  return new Date(Date.now() + 7 * 60 * 60 * 1000).toISOString().slice(0, 10)
}

// Picking a period fills its prices and the payment defaults, like the summary tool: a deposit when
// departure is more than 30 days away and the period has one; the balance due 30 days before.
export function applyPeriod(deal: Deal, period: TourPeriodItem): Deal {
  const start = period.startDate.slice(0, 10)
  const far = new Date(`${start}T00:00:00Z`).getTime() - new Date(`${todayInThailand()}T00:00:00Z`).getTime() > 30 * DAY_MS
  const isDeposit = far && n(period.deposit) > 0
  return {
    ...deal,
    adult: { ...deal.adult, price: period.price },
    child: { ...deal.child, price: period.childWithBedPrice ?? period.price },
    singleRoom: { ...deal.singleRoom, price: period.singleSupplement },
    adultPriceBeforeDiscount: period.priceBeforeDiscount,
    payment: {
      ...deal.payment,
      isDeposit,
      depositPerPerson: isDeposit ? period.deposit : deal.payment.depositPerPerson,
      remainingDueDate: addDays(start, -30),
    },
  }
}

// Client-side checks before submit; the backend checks the same and more.
export function dealError(deal: Deal): string | null {
  const adults = n(deal.adult.count)
  if (adults < 1) return 'กรุณาระบุผู้ใหญ่อย่างน้อย 1 ท่าน'
  if (travellers(deal) > 50) return 'ผู้เดินทางได้ไม่เกิน 50 ท่านต่อการจอง'
  if (deal.adult.price === null) return 'กรุณาระบุราคาผู้ใหญ่'
  if (n(deal.child.count) > 0 && deal.child.price === null) return 'กรุณาระบุราคาเด็ก'
  if (n(deal.infant.count) > 0 && deal.infant.price === null) return 'กรุณาระบุราคาทารก'
  if (n(deal.singleRoom.count) > adults) return 'ห้องพักเดี่ยวมากกว่าจำนวนผู้ใหญ่'
  if (n(deal.singleRoom.count) > 0 && deal.singleRoom.price === null) return 'กรุณาระบุราคาห้องพักเดี่ยว'
  if (deal.options.some((o) => n(o.count) > 0 && n(o.price) > 0 && !o.name.trim())) return 'กรุณาระบุชื่อรายการเพิ่มเติม'
  const t = computeTotals(deal)
  if (t.subtotal < 0) return 'ส่วนลดมากกว่าราคา'
  if (deal.payment.isDeposit && n(deal.payment.depositPerPerson) <= 0) return 'กรุณาระบุมัดจำต่อท่าน'
  if (deal.payment.isDeposit && t.deposit > t.subtotal) return 'มัดจำมากกว่ายอดรวม'
  return null
}

// The deal part of POST /line-bookings, in the backend's room types.
export function dealRequest(deal: Deal): Omit<
  LineBookingCreateRequest,
  | 'salesUserId' | 'lineContactId' | 'bookingSourceId' | 'refName' | 'refBookingNo' | 'tourCode' | 'periodId'
  | 'customerName' | 'customerPhone' | 'customerAddress' | 'slipImage' | 'note'
> {
  const adults = n(deal.adult.count)
  const single = Math.min(n(deal.singleRoom.count), adults)
  const children = n(deal.child.count)
  const infants = n(deal.infant.count)
  const pay = deal.payment
  return {
    adultTwin: adults - single,
    adultSingle: single,
    childWithBed: children,
    childNoBed: 0,
    infant: infants,
    adultPrice: n(deal.adult.price),
    singleSupplement: single > 0 ? n(deal.singleRoom.price) : null,
    childWithBedPrice: children > 0 ? n(deal.child.price) : null,
    childNoBedPrice: null,
    infantPrice: infants > 0 ? n(deal.infant.price) : null,
    adultPriceBeforeDiscount: deal.adultPriceBeforeDiscount,
    discountPerPerson: n(deal.discountPerPerson),
    charges: chargeLines(deal),
    paymentMethod: pay.method,
    creditFeePercent: pay.method === 'credit' ? n(pay.creditFeePercent) : 0,
    isDeposit: pay.isDeposit,
    depositPerPerson: pay.isDeposit ? n(pay.depositPerPerson) : 0,
    depositDueDate: pay.isDeposit && pay.depositDueDate ? pay.depositDueDate : null,
    depositDueTime: pay.isDeposit ? pay.depositDueTime : '',
    remainingDueDate: pay.isDeposit && pay.remainingDueDate ? pay.remainingDueDate : null,
    fullDueDate: !pay.isDeposit && pay.fullDueDate ? pay.fullDueDate : null,
    fullDueTime: pay.isDeposit ? '' : pay.fullDueTime,
  }
}

const fromPriced = (line: PricedLine | null): Line => (line ? { count: line.count || null, price: line.price } : emptyLine())

// The deal from a pasted summary: the same rows, so it maps one to one.
export function dealFromSummary(s: ParsedSummary): Deal {
  const base = emptyDeal()
  const child = fromPriced(s.child)
  const infant = fromPriced(s.infant)
  const adults = Math.max(0, (s.travellers ?? 0) - n(child.count) - n(infant.count))
  return {
    adult: { count: adults || null, price: s.adultPrice },
    child,
    infant,
    singleRoom: fromPriced(s.singleRoom),
    adultPriceBeforeDiscount: s.adultPriceBeforeDiscount,
    discountPerPerson: s.discountPerPerson,
    charges: {
      visa: fromPriced(s.visa),
      foreignPassport: fromPriced(s.foreignPassport),
      joinLand: fromPriced(s.joinLand),
      fuelTax: fromPriced(s.fuelTax),
      airportTax: fromPriced(s.airportTax),
    },
    options: s.options.map((o) => ({ name: o.name, count: o.count, price: o.price })),
    payment: {
      method: s.creditFeePercent !== null ? 'credit' : 'transfer',
      creditFeePercent: s.creditFeePercent ?? base.payment.creditFeePercent,
      isDeposit: s.isDeposit ?? false,
      depositPerPerson: s.depositPerPerson,
      depositDueDate: s.depositDueDate ?? '',
      depositDueTime: s.depositDueTime ?? '',
      remainingDueDate: s.remainingDueDate ?? '',
      fullDueDate: s.fullDueDate ?? '',
      fullDueTime: s.fullDueTime ?? '',
    },
  }
}
