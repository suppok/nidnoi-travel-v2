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
  adultTwin: number;
  adultSingle: number;
  childWithBed: number;
  childNoBed: number;
  price: number;
  singleSupplement: number | null;
  childWithBedPrice: number | null;
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
