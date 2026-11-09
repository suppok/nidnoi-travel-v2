// Browser call used by the booking form. Rate-limited by the backend (RATE_LIMITED).
import type { BookingPublicCreateRequest, BookingPublicCreateResponse } from '../../types/booking';
import { apiPost } from './client';
import { USE_MOCK_DATA } from './mock-mode';

export async function submitBooking(body: BookingPublicCreateRequest): Promise<BookingPublicCreateResponse> {
  if (USE_MOCK_DATA) {
    await new Promise((r) => setTimeout(r, 600));
    return { bookingNo: '26MOCK42', totalPrice: 0 };
  }
  return apiPost<BookingPublicCreateResponse>('/bookings/public', body);
}
