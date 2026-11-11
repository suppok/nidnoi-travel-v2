import type { ListResponse } from '../types/api.ts';
import type { BookingSource, BookingSourceCreateRequest, BookingSourceListRequest, BookingSourceStatusRequest, BookingSourceUpdateRequest } from '../types/bookingSource.ts';
import { request } from './client.ts';

export const bookingSourcesApi = {
  list: (body: BookingSourceListRequest) => request<ListResponse<BookingSource>>('POST', '/booking-sources/list', body),
  get: (id: string) => request<BookingSource>('GET', `/booking-sources/${encodeURIComponent(id)}`),
  create: (body: BookingSourceCreateRequest) => request<BookingSource>('POST', '/booking-sources', body),
  update: (body: BookingSourceUpdateRequest) => request<BookingSource>('PUT', '/booking-sources', body),
  updateStatus: (body: BookingSourceStatusRequest) => request<null>('PUT', '/booking-sources/status', body),
};
