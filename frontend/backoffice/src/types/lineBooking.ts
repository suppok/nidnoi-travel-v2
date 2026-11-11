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
  // Deal details from the "สรุปยอดโอน" summary; defaults (0, [], 'transfer', null) on older bookings.
  // Unit prices above are the agreed ones; totalPrice = rows - discount + charges + cardFee.
  infantCount: number;
  infantPrice: number | null;
  adultPriceBeforeDiscount: number | null;
  discountPerPerson: number; // per traveller, infants included
  charges: LineBookingCharge[];
  paymentMethod: LineBookingPaymentMethod;
  creditFeePercent: number;
  cardFee: number;
  isDeposit: boolean;
  depositPerPerson: number;
  depositDueDate: string | null; // YYYY-MM-DD
  depositDueTime: string; // HH:mm or ''
  remainingDueDate: string | null;
  fullDueDate: string | null;
  fullDueTime: string;
  customerAddress: string;
  bookingSource: { id: string; name: string; shortName: string } | null;
  refName: string; // the customer's name on the booking source
  refBookingNo: string;
  // The sales person who owns the booking; null on bookings made before it was recorded.
  salesUser: { id: string; name: string } | null;
  // The customer's LINE OA chat; null until staff link it.
  lineContact: LineBookingContact | null;
};

// MUST match backend pkg/constants/linebookingpayment.
export type LineBookingPaymentMethod = 'transfer' | 'credit';
export type LineBookingChargeKind = 'visa' | 'foreignPassport' | 'joinLand' | 'fuelTax' | 'airportTax' | 'option';
export type LineBookingCharge = { kind: LineBookingChargeKind; name: string; count: number; price: number };

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
