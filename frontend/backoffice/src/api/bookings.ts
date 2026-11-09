import type { ListResponse } from '../types/api.ts';
import type { Booking, BookingListRequest, BookingUpdateRequest, BookingUpdateStatusRequest } from '../types/booking.ts';
import { request } from './client.ts';

export const bookingsApi = {
  list: (body: BookingListRequest) => request<ListResponse<Booking>>('POST', '/bookings/list', body),
  get: (id: string) => request<Booking>('GET', `/bookings/${encodeURIComponent(id)}`),
  update: (body: BookingUpdateRequest) => request<Booking>('PUT', '/bookings', body),
  updateStatus: (body: BookingUpdateStatusRequest) => request<null>('PUT', '/bookings/status', body),
};
