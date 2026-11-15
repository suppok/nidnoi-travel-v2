import type { ListResponse } from '../types/api.ts';
import type {
  LineBooking,
  LineBookingConfirmRequest,
  LineBookingCreateRequest,
  LineBookingCreateResponse,
  LineBookingDealUpdateRequest,
  LineBookingDealUpdateResponse,
  LineBookingDocumentsRequest,
  LineBookingEvent,
  LineBookingNotification,
  LineBookingFinanceUsers,
  LineBookingListRequest,
  LineBookingUpdateLineContactRequest,
  LineBookingUpdateRequest,
  LineBookingUpdateStatusRequest,
} from '../types/lineBooking.ts';
import type { MediaUploadResponse } from '../types/media.ts';
import { request } from './client.ts';

export const lineBookingsApi = {
  // --- booking entry (บันทึกใบจอง LINE) ---
  create: (body: LineBookingCreateRequest) => request<LineBookingCreateResponse>('POST', '/line-bookings', body),
  uploadSlip: (file: File) => {
    const body = new FormData();
    body.append('file', file);
    return request<MediaUploadResponse>('POST', '/media/upload/line-booking-slip', body);
  },
  // Reopen a paid booking by its booking no, or list a customer chat's; any sales user may.
  getEditable: (bookingNo: string) => request<LineBooking>('GET', `/line-bookings/editable/${encodeURIComponent(bookingNo)}`),
  listEditable: (lineContactId: string) => request<{ items: LineBooking[] }>('POST', '/line-bookings/editable/list', { lineContactId }),
  editableEvents: (bookingNo: string) =>
    request<{ items: LineBookingEvent[] }>('GET', `/line-bookings/editable/${encodeURIComponent(bookingNo)}/events`),
  updateDeal: (body: LineBookingDealUpdateRequest) => request<LineBookingDealUpdateResponse>('PUT', '/line-bookings/editable', body),

  list: (body: LineBookingListRequest) => request<ListResponse<LineBooking>>('POST', '/line-bookings/list', body),
  get: (id: string) => request<LineBooking>('GET', `/line-bookings/${encodeURIComponent(id)}`),
  // Comments: anyone who can see the booking; mentionUserIds from mentionable(id).
  comment: (body: { id: string; text: string; mentionUserIds: string[] }) => request<LineBookingEvent>('POST', '/line-bookings/comments', body),
  mentionable: (id: string) => request<{ items: { id: string; name: string }[] }>('GET', `/line-bookings/${encodeURIComponent(id)}/mentionable`),
  notifications: (body: { unread?: boolean; page: number; pageSize: number }) =>
    request<ListResponse<LineBookingNotification>>('POST', '/line-booking-notifications/list', body),
  // Opening a booking marks the caller's notifications on it read.
  readNotifications: (lineBookingId: string) => request<null>('PUT', '/line-booking-notifications/read', { lineBookingId }),
  events: (id: string) => request<{ items: LineBookingEvent[] }>('GET', `/line-bookings/${encodeURIComponent(id)}/events`),
  update: (body: LineBookingUpdateRequest) => request<LineBooking>('PUT', '/line-bookings', body),
  updateLineContact: (body: LineBookingUpdateLineContactRequest) => request<LineBooking>('PUT', '/line-bookings/line-contact', body),
  confirm: (body: LineBookingConfirmRequest) => request<LineBooking>('PUT', '/line-bookings/confirm', body),
  // The sales owner confirms the balance sales recorded: paid in full, back to finance for the RE.
  confirmBalance: (id: string) => request<LineBooking>('PUT', '/line-bookings/confirm-balance', { id }),
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
