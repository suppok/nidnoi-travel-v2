import { env } from '../config/env.ts'
import type {
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

async function send(method: string, path: string, body: unknown): Promise<Response> {
  const isForm = body instanceof FormData
  try {
    return await fetch(`${env.API_BASE_URL}${path}`, {
      method,
      headers: {
        ...(isForm ? {} : { 'Content-Type': 'application/json' }),
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
  const res = await send('POST', '/tour-programs/public/list', {
    keyword,
    page: 1,
    pageSize: 20,
  })
  const data = await parse<{ items: TourPublicItem[] }>(res)
  return data.items
}

export async function getTourPeriods(code: string): Promise<TourPeriodItem[]> {
  const res = await send('GET', `/tour-programs/public/${code}/periods`, undefined)
  return parse<TourPeriodItem[]>(res)
}

export async function uploadSlip(file: File, liffAccessToken: string): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  form.append('purpose', 'booking-slip')
  form.append('liffAccessToken', liffAccessToken)
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
