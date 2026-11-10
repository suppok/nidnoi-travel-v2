import type { ListResponse } from '../types/api.ts';
import type { LineBooking, LineBookingListRequest, LineBookingUpdateRequest, LineBookingUpdateStatusRequest } from '../types/lineBooking.ts';
import { request } from './client.ts';

export const lineBookingsApi = {
  list: (body: LineBookingListRequest) => request<ListResponse<LineBooking>>('POST', '/line-bookings/list', body),
  get: (id: string) => request<LineBooking>('GET', `/line-bookings/${encodeURIComponent(id)}`),
  update: (body: LineBookingUpdateRequest) => request<LineBooking>('PUT', '/line-bookings', body),
  updateStatus: (body: LineBookingUpdateStatusRequest) => request<null>('PUT', '/line-bookings/status', body),
};
