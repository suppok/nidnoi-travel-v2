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
  // The finance user who issues the receipt, and when the sales owner confirmed; null while new.
  financeUser: { id: string; name: string } | null;
  confirmedAt: string | null;
  // Finance documents (PDF URLs, '' = none): deposit → aiFile + ivFile, full payment → reFile.
  aiFile: string;
  ivFile: string;
  reFile: string;
  // When finance sent them; null on a confirmed booking = waiting for finance.
  documentsAt: string | null;
  // The sales owner approved the documents; customerSentAt = the card was pushed to the customer's
  // LINE chat (null = approved without sending).
  documentsApprovedAt: string | null;
  customerSentAt: string | null;
  // The latest fix request from sales to finance ('' = none).
  fixNote: string;
  fixRequestedAt: string | null;
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
  paid?: boolean; // confirmed only (deposit paid / paid in full), newest confirmation first
  assignedToMe?: boolean; // only bookings sent to the caller as finance
  ownedByMe?: boolean; // only bookings the caller owns as sales
  documentsDone?: boolean; // true = finance sent the documents (newest first), false = waiting
  documentsApproved?: boolean; // true = the sales owner approved them, false = not yet
};

// The state changes only through confirm / cancel.
export type LineBookingUpdateRequest = { id: string; staffNote: string };

// financeUserId is one of GET /line-bookings/:id/finance-users.
export type LineBookingConfirmRequest = { id: string; financeUserId: string };

export type LineBookingDocumentKey = 'aiFile' | 'ivFile' | 'reFile';

// PDF URLs from POST /media/upload/line-booking; '' for documents the plan doesn't use.
export type LineBookingDocumentsRequest = { id: string } & Record<LineBookingDocumentKey, string>;

export type LineBookingFinanceUsers = { items: { id: string; name: string }[]; defaultId: string };

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
