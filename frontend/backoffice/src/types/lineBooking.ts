import type { LineBookingState } from '../constants/lineBookingStates.ts';
import type { Paging } from './api.ts';

// A booking submitted by sales staff through the LINE OA LIFF app.
// Tour, dates and prices are a snapshot taken at submission time.
export type LineBooking = {
  id: string;
  bookingNo: string;
  lineUserId: string;
  lineDisplayName: string;
  tourProgramId: string;
  tourProgramPeriodId: string;
  tourCode: string;
  tourName: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  // Travellers per price row and the unit price quoted. A null price means the period had no such row.
  adultTwinCount: number;
  adultTwinPrice: number;
  adultSingleCount: number;
  adultSinglePrice: number | null;
  childWithBedCount: number;
  childWithBedPrice: number | null;
  childNoBedCount: number;
  childNoBedPrice: number | null;
  totalPrice: number;
  customerName: string;
  customerPhone: string;
  slipImage: string;
  note: string;
  state: LineBookingState;
  staffNote: string;
  status: boolean;
  createDate: string;
  updateDate: string;
  // The customer's LINE OA chat; null until staff link it.
  lineContact: LineBookingContact | null;
};

export type LineBookingContact = {
  id: string;
  displayName: string;
  pictureUrl: string | null;
  // False once the customer blocked the OA: a LINE push would not reach them.
  following: boolean;
};

export type LineBookingListRequest = Paging & {
  keyword?: string;
  state?: LineBookingState;
  status?: boolean;
};

export type LineBookingUpdateRequest = {
  id: string;
  state: LineBookingState;
  staffNote: string;
};

export type LineBookingUpdateStatusRequest = { id: string; status: boolean };

// lineContactId null unlinks the chat.
export type LineBookingUpdateLineContactRequest = { id: string; lineContactId: string | null };

// A customer known from LINE OA chats (POST /line-contacts/list).
export type LineContact = {
  id: string;
  lineUserId: string;
  displayName: string;
  pictureUrl: string | null;
  following: boolean;
  lastMessageAt: string | null;
};

export type LineContactListRequest = Paging & {
  keyword?: string;
  following?: boolean;
};
