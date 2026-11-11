import { env } from '../config/env.ts'
import type {
  BookingSource,
  LIFFMe,
  LineContact,
  SalesUserList,
  LineBookingCreateRequest,
  LineBookingCreateResponse,
  MediaUploadResponse,
  TourPeriodItem,
  TourPublicItem,
} from '../types/index.ts'

export class ApiError extends Error {
  code: string
  status: number
  fields?: Record<string, string>

  constructor(code: string, message: string, status: number, fields?: Record<string, string>) {
    super(message)
    this.code = code
    this.status = status
    this.fields = fields
  }
}

// The LIFF access token authenticates every staff call ("Authorization: Bearer"). The backend
// maps it to the backoffice user who linked this LINE account.
let liffToken = ''

export function setLiffToken(token: string) {
  liffToken = token
}

async function send(method: string, path: string, body: unknown): Promise<Response> {
  if (env.MOCK) {
    const { mockSend } = await import('./mock.ts')
    const mocked = await mockSend(method, path, body)
    if (mocked) return mocked
  }
  const isForm = body instanceof FormData
  try {
    return await fetch(`${env.API_BASE_URL}${path}`, {
      method,
      headers: {
        ...(isForm || body === undefined ? {} : { 'Content-Type': 'application/json' }),
        ...(liffToken ? { Authorization: `Bearer ${liffToken}` } : {}),
      },
      body: body === undefined ? undefined : isForm ? body : JSON.stringify(body),
    })
  } catch {
    throw new ApiError('NETWORK', 'Cannot reach the server', 0)
  }
}

async function parse<T>(res: Response): Promise<T> {
  const json = await res.json().catch(() => null)
  if (!res.ok || json?.success === false) {
    throw new ApiError(
      json?.error?.code ?? 'UNKNOWN',
      json?.error?.message ?? res.statusText,
      res.status,
      json?.error?.fields,
    )
  }
  return json.data as T
}

export async function searchTours(keyword: string): Promise<TourPublicItem[]> {
  const query = new URLSearchParams({ keyword, page: '1', pageSize: '20' })
  const res = await send('GET', `/tour-programs/public?${query}`, undefined)
  const data = await parse<{ items: TourPublicItem[] }>(res)
  return data.items
}

export async function getTourPeriods(code: string): Promise<TourPeriodItem[]> {
  const res = await send('GET', `/tour-programs/public/${encodeURIComponent(code)}/periods`, undefined)
  const data = await parse<{ items: TourPeriodItem[] }>(res)
  return data.items
}

// The staff user linked to this LINE account. Throws LINE_NOT_LINKED, PERMISSION_DENIED (e.g.
// finance) or PASSWORD_CHANGE_REQUIRED when the app may not be used.
export async function getLiffMe(): Promise<LIFFMe> {
  return parse<LIFFMe>(await send('GET', '/liff/me', undefined))
}

// Finishes linking this LINE account with the one-time code from the backoffice QR code.
export async function linkLineAccount(code: string): Promise<{ name: string; email: string }> {
  return parse<{ name: string; email: string }>(await send('POST', '/liff/line-link', { code }))
}

// Customers who chatted with the OA, most recent chat first; matches the LINE display name.
export async function searchLineContacts(keyword: string): Promise<LineContact[]> {
  const res = await send('POST', '/liff/line-contacts/list', { keyword, page: 1, pageSize: 20 })
  return (await parse<{ items: LineContact[] }>(res)).items
}

export async function getSalesUsers(): Promise<SalesUserList> {
  return parse<SalesUserList>(await send('GET', '/liff/sales-users', undefined))
}

export async function getBookingSources(): Promise<BookingSource[]> {
  return (await parse<{ items: BookingSource[] }>(await send('GET', '/liff/booking-sources', undefined))).items
}

export async function uploadSlip(file: File): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  form.append('purpose', 'booking-slip')
  const res = await send('POST', '/media/upload/liff', form)
  const data = await parse<MediaUploadResponse>(res)
  return data.url
}

export async function createLineBooking(
  body: LineBookingCreateRequest,
): Promise<LineBookingCreateResponse> {
  const res = await send('POST', '/line-bookings', body)
  return parse<LineBookingCreateResponse>(res)
}
