import type { ListResponse } from '../types/api.ts';
import type {
  LineBooking,
  LineBookingConfirmRequest,
  LineBookingDocumentsRequest,
  LineBookingFinanceUsers,
  LineBookingListRequest,
  LineBookingUpdateLineContactRequest,
  LineBookingUpdateRequest,
  LineBookingUpdateStatusRequest,
} from '../types/lineBooking.ts';
import type { MediaUploadResponse } from '../types/media.ts';
import { request } from './client.ts';

export const lineBookingsApi = {
  list: (body: LineBookingListRequest) => request<ListResponse<LineBooking>>('POST', '/line-bookings/list', body),
  get: (id: string) => request<LineBooking>('GET', `/line-bookings/${encodeURIComponent(id)}`),
  update: (body: LineBookingUpdateRequest) => request<LineBooking>('PUT', '/line-bookings', body),
  updateLineContact: (body: LineBookingUpdateLineContactRequest) => request<LineBooking>('PUT', '/line-bookings/line-contact', body),
  confirm: (body: LineBookingConfirmRequest) => request<LineBooking>('PUT', '/line-bookings/confirm', body),
  cancel: (id: string) => request<LineBooking>('PUT', '/line-bookings/cancel', { id }),
  documents: (body: LineBookingDocumentsRequest) => request<LineBooking>('PUT', '/line-bookings/documents', body),
  // Sales review: approve (pushes the card to the customer's LINE chat when linked) or send back.
  approveDocuments: (id: string) => request<LineBooking>('PUT', '/line-bookings/documents/approve', { id }),
  rejectDocuments: (id: string, note: string) => request<LineBooking>('PUT', '/line-bookings/documents/reject', { id, note }),
  // A finance PDF (AI / IV / RE); the purpose is fixed by the endpoint.
  uploadDocument: (file: File) => {
    const body = new FormData();
    body.append('file', file);
    return request<MediaUploadResponse>('POST', '/media/upload/line-booking', body);
  },
  financeUsers: (id: string) => request<LineBookingFinanceUsers>('GET', `/line-bookings/${encodeURIComponent(id)}/finance-users`),
  updateStatus: (body: LineBookingUpdateStatusRequest) => request<null>('PUT', '/line-bookings/status', body),
};
