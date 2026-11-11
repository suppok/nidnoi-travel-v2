// Local-only preview without LINE (VITE_LIFF_MOCK=true, VITE_APP_ENV=local). Lazy-loaded by
// client.ts, so it never ships when the flag is off. Only the endpoints that need a LIFF token are
// faked; public tour/period endpoints still go to the real backend at VITE_API_BASE_URL, so a
// summary copied from the local backoffice matches real tours.
import { env } from '../config/env.ts'
import type { BookingSource, LineBookingCreateRequest, LineContact, SalesUserList, TourPeriodItem } from '../types/index.ts'

const LATENCY_MS = 250

// Names match the backoffice mock users, so "ผู้ดูแล" in a mock summary resolves.
const SALES: SalesUserList = {
  items: [
    { id: '0b9c6a52-0000-4000-8000-000000000003', name: 'มานี บัญชีดี' },
    { id: '0b9c6a52-0000-4000-8000-000000000004', name: 'พนักงานใหม่' },
    { id: '0b9c6a52-0000-4000-8000-000000000002', name: 'สมใจ ขายเก่ง' },
  ],
  defaultId: '0b9c6a52-0000-4000-8000-000000000002',
}

const CONTACTS: LineContact[] = [
  { id: 'c1000000-0000-4000-8000-000000000001', displayName: 'Somchai', pictureUrl: null, following: true, lastMessageAt: '2026-11-11T08:30:00Z' },
  { id: 'c1000000-0000-4000-8000-000000000002', displayName: 'น้องแพร 🌸', pictureUrl: null, following: true, lastMessageAt: '2026-11-10T14:00:00Z' },
  { id: 'c1000000-0000-4000-8000-000000000003', displayName: 'Ploy', pictureUrl: null, following: false, lastMessageAt: '2026-11-01T09:00:00Z' },
]

const json = (data: unknown, status = 200) => Response.json({ success: true, data }, { status })
const failure = (status: number, code: string, message: string, fields?: Record<string, string>) =>
  Response.json({ success: false, error: { code, message, ...(fields ? { fields } : {}) } }, { status })

const SOURCES: BookingSource[] = [
  { id: 'b5000000-0000-4000-8000-000000000001', name: 'LINE OA', shortName: 'LINE' },
  { id: 'b5000000-0000-4000-8000-000000000002', name: 'Facebook', shortName: 'FB' },
  { id: 'b5000000-0000-4000-8000-000000000003', name: 'Walk-in', shortName: '' },
]

// Same pricing as the backend's LINE booking quote (quoteLine): entered prices, else the period's.
async function quote(b: LineBookingCreateRequest): Promise<number | null> {
  const res = await fetch(`${env.API_BASE_URL}/tour-programs/public/${encodeURIComponent(b.tourCode)}/periods`)
  const p = ((await res.json()) as { data?: { items: TourPeriodItem[] } }).data?.items.find((x) => x.id === b.periodId)
  if (!p) return null
  const adult = b.adultPrice ?? p.price
  const pax = b.adultTwin + b.adultSingle + b.childWithBed + b.childNoBed + b.infant
  const subtotal =
    b.adultTwin * adult +
    b.adultSingle * (adult + (b.singleSupplement ?? p.singleSupplement ?? 0)) +
    b.childWithBed * (b.childWithBedPrice ?? p.childWithBedPrice ?? 0) +
    b.childNoBed * (b.childNoBedPrice ?? p.childNoBedPrice ?? 0) +
    b.infant * (b.infantPrice ?? 0) -
    b.discountPerPerson * pax +
    b.charges.reduce((sum, c) => sum + c.count * c.price, 0)
  const bps = Math.round(b.creditFeePercent * 100)
  const base = b.isDeposit ? b.depositPerPerson * pax : subtotal
  return subtotal + (b.paymentMethod === 'credit' ? Math.ceil((base * bps) / 10000) : 0)
}

export async function mockSend(method: string, path: string, body: unknown): Promise<Response | null> {
  const route = `${method} ${path.split('?')[0]}`
  const handlers: Record<string, () => Response | Promise<Response>> = {
    'GET /liff/me': () => json({ name: 'สมใจ ขายเก่ง (mock)', email: 'sales@nidnoi.test', roles: ['sales'], lineDisplayName: 'Somjai' }),
    'POST /liff/line-link': () => json({ name: 'สมใจ ขายเก่ง (mock)', email: 'sales@nidnoi.test' }),
    'GET /liff/sales-users': () => json(SALES),
    'GET /liff/booking-sources': () => json({ items: SOURCES }),
    'POST /liff/line-contacts/list': () => {
      const keyword = String((body as { keyword?: string })?.keyword ?? '').toLowerCase()
      return json({ items: CONTACTS.filter((c) => !keyword || c.displayName.toLowerCase().includes(keyword)) })
    },
    'POST /media/upload/liff': () => {
      const file = body instanceof FormData ? body.get('file') : null
      if (!(file instanceof File)) return failure(400, 'VALIDATION', 'Validation failed', { file: 'required' })
      return json({ url: URL.createObjectURL(file) })
    },
    'POST /line-bookings': async () => {
      const req = body as LineBookingCreateRequest
      const total = await quote(req)
      if (total === null) return failure(409, 'PERIOD_UNAVAILABLE', 'This departure is no longer open for booking')
      const bookingNo = `26${Math.random().toString(36).slice(2, 8).toUpperCase()}`
      console.info('[mock] POST /line-bookings', req)
      return json({ bookingNo, totalPrice: total })
    },
  }
  const handler = handlers[route]
  if (!handler) return null // not mocked: let client.ts call the real backend
  await new Promise((r) => setTimeout(r, LATENCY_MS))
  return handler()
}
