// Local-only preview without LINE (VITE_LIFF_MOCK=true, VITE_APP_ENV=local). Lazy-loaded by
// client.ts, so it never ships when the flag is off. Only the endpoints that need a LIFF token are
// faked; public tour/period endpoints still go to the real backend at VITE_API_BASE_URL, so a
// summary copied from the local backoffice matches real tours.
import { env } from '../config/env.ts'
import type {
  BookingSource,
  LineBookingCreateRequest,
  LineBookingItem,
  LineBookingLIFFUpdateRequest,
  LineContact,
  SalesUserList,
  TourPeriodItem,
} from '../types/index.ts'

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

// A deposit booking whose deposit documents were sent: open it with booking no "26DEMO01".
const DEPOSIT_BOOKING: LineBookingItem = {
  id: 'd1000000-0000-4000-8000-000000000001',
  bookingNo: '26DEMO01',
  state: 20,
  tourCode: 'NIDN000001',
  tourName: 'ทัวร์ญี่ปุ่น โตเกียว ฟูจิ 5 วัน 3 คืน (mock)',
  tourProgramPeriodId: 'e1000000-0000-4000-8000-000000000001',
  startDate: '2027-03-10',
  endDate: '2027-03-14',
  adultTwinCount: 2,
  adultTwinPrice: 39900,
  adultSingleCount: 0,
  adultSinglePrice: null,
  childWithBedCount: 0,
  childWithBedPrice: null,
  childNoBedCount: 0,
  childNoBedPrice: null,
  infantCount: 0,
  infantPrice: null,
  adultPriceBeforeDiscount: 42900,
  discountPerPerson: 0,
  charges: [],
  paymentMethod: 'transfer',
  creditFeePercent: 0,
  cardFee: 0,
  isDeposit: true,
  depositPerPerson: 10000,
  depositDueDate: '2026-11-12',
  depositDueTime: '18:00',
  remainingDueDate: '2027-02-08',
  fullDueDate: null,
  fullDueTime: '',
  totalPrice: 79800,
  customerName: 'สมชาย ใจดี',
  customerPhone: '0812345678',
  customerAddress: '',
  refName: '',
  refBookingNo: '',
  bookingSource: SOURCES[0],
  salesUser: SALES.items[2],
  lineContact: { id: CONTACTS[0].id, displayName: CONTACTS[0].displayName, pictureUrl: null, following: true },
  slipImage: 'https://placehold.co/400x600?text=deposit+slip+1',
  slipImages: ['https://placehold.co/400x600?text=deposit+slip+1', 'https://placehold.co/400x600?text=deposit+slip+2'],
  note: '',
  balanceSlipImage: '',
  balanceSlipImages: [],
  balanceSubmittedAt: null,
  balanceConfirmedAt: null,
  financeUser: { id: '0b9c6a52-0000-4000-8000-000000000003', name: 'มานี บัญชีดี' },
  documentsAt: '2026-11-10T09:00:00Z',
  documentsApprovedAt: '2026-11-10T10:00:00Z',
}

// A booking paid in full whose RE was sent: "26FULL01".
const FULL_BOOKING: LineBookingItem = {
  ...DEPOSIT_BOOKING,
  id: 'd1000000-0000-4000-8000-000000000002',
  bookingNo: '26FULL01',
  state: 25,
  isDeposit: false,
  depositPerPerson: 0,
  depositDueDate: null,
  depositDueTime: '',
  remainingDueDate: null,
  fullDueDate: '2026-11-20',
  fullDueTime: '18:00',
  slipImage: 'https://placehold.co/400x600?text=full+payment+slip',
  slipImages: ['https://placehold.co/400x600?text=full+payment+slip'],
}

// A deposit booking whose AI + IV finance hasn't sent yet: "26WAIT01".
const WAITING_BOOKING: LineBookingItem = {
  ...DEPOSIT_BOOKING,
  id: 'd1000000-0000-4000-8000-000000000003',
  bookingNo: '26WAIT01',
  documentsAt: null,
  documentsApprovedAt: null,
}

const EDITABLE = [DEPOSIT_BOOKING, FULL_BOOKING, WAITING_BOOKING]

// The same history for every mock booking.
const SOMJAI = { id: SALES.items[2].id, name: SALES.items[2].name }
const MANEE = { id: SALES.items[0].id, name: SALES.items[0].name }
const MOCK_HISTORY = [
  { id: 'e1', action: 'created', changes: [], target: '', note: '', createDate: '2026-11-09T08:00:00Z', user: SOMJAI },
  { id: 'e2', action: 'confirmed', changes: [], target: MANEE.name, note: '', createDate: '2026-11-09T09:30:00Z', user: SOMJAI },
  { id: 'e3', action: 'documents_sent', changes: [], target: 'AI + IV', note: '', createDate: '2026-11-10T09:00:00Z', user: MANEE },
  { id: 'e4', action: 'documents_approved', changes: [], target: 'Somchai', note: 'AI + IV', createDate: '2026-11-10T10:00:00Z', user: SOMJAI },
  {
    id: 'e5',
    action: 'edited',
    changes: [
      { field: 'adultTwin', from: '2', to: '3' },
      { field: 'totalPrice', from: '79800', to: '119700' },
    ],
    target: '',
    note: '',
    createDate: '2026-11-12T04:15:00Z',
    user: SOMJAI,
  },
  {
    id: 'e6',
    action: 'comment',
    changes: [],
    target: '',
    note: '@สมใจ ขายเก่ง ลูกค้าแจ้งว่าจะโอนส่วนที่เหลือพรุ่งนี้',
    mentions: [SOMJAI],
    createDate: '2026-11-13T02:00:00Z',
    user: MANEE,
  },
]

// The update's total from its own prices (no period lookup: every price is sent).
function mockTotal(b: LineBookingLIFFUpdateRequest): number {
  const adult = b.adultPrice ?? 0
  const pax = b.adultTwin + b.adultSingle + b.childWithBed + b.childNoBed + b.infant
  const subtotal =
    b.adultTwin * adult +
    b.adultSingle * (adult + (b.singleSupplement ?? 0)) +
    b.childWithBed * (b.childWithBedPrice ?? 0) +
    b.childNoBed * (b.childNoBedPrice ?? 0) +
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
      // Whole display name only, like the backend.
      return json({ items: CONTACTS.filter((c) => !keyword || c.displayName.toLowerCase() === keyword.trim()) })
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
    // Bookings of one chat: the mock deposit booking is linked to Somchai.
    'POST /liff/line-bookings/list': () => {
      const id = String((body as { lineContactId?: string })?.lineContactId ?? '')
      return json({ items: EDITABLE.filter((b) => b.lineContact?.id === id) })
    },
    'PUT /liff/line-bookings': () => {
      const req = body as LineBookingLIFFUpdateRequest
      if (!EDITABLE.some((b) => b.bookingNo === req.bookingNo)) return failure(404, 'NOT_FOUND', 'Not found')
      console.info('[mock] PUT /liff/line-bookings', req)
      return json({ bookingNo: req.bookingNo, totalPrice: mockTotal(req), balancePending: req.balanceSlipImages.length > 0 })
    },
  }
  // GET /liff/line-bookings/:bookingNo — "26DEMO01" (deposit), "26FULL01" (paid in full) and
// "26WAIT01" (waiting for finance: warns on save) open; "26NEW001" answers LINE_BOOKING_NOT_EDITABLE.
  if (method === 'GET' && path.startsWith('/liff/line-bookings/') && path.endsWith('/events')) {
    await new Promise((r) => setTimeout(r, LATENCY_MS))
    return json({ items: MOCK_HISTORY })
  }
  if (method === 'GET' && path.startsWith('/liff/line-bookings/')) {
    const no = decodeURIComponent(path.slice('/liff/line-bookings/'.length)).toUpperCase()
    await new Promise((r) => setTimeout(r, LATENCY_MS))
    const found = EDITABLE.find((b) => b.bookingNo === no)
    if (found) return json(found)
    if (no === '26NEW001') return failure(409, 'LINE_BOOKING_NOT_EDITABLE', 'Only a paid booking can be edited')
    return failure(404, 'NOT_FOUND', 'Not found')
  }
  const handler = handlers[route]
  if (!handler) return null // not mocked: let client.ts call the real backend
  await new Promise((r) => setTimeout(r, LATENCY_MS))
  return handler()
}
