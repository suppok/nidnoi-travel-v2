import type { BookingState } from '../constants/bookingStates.ts';
import type { Paging } from './api.ts';

// A booking request from the customer site (a sales lead: no seats reserved, no payment).
// Tour, dates and prices are a snapshot taken when the customer sent it.
export type Booking = {
  id: string;
  bookingNo: string;
  tourProgramId: string;
  tourProgramPeriodId: string;
  tourCode: string;
  tourName: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  adultTwinCount: number;
  adultTwinPrice: number;
  adultSingleCount: number;
  adultSinglePrice: number | null;
  childWithBedCount: number;
  childWithBedPrice: number | null;
  childNoBedCount: number;
  childNoBedPrice: number | null;
  totalPrice: number;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  customerNote: string;
  state: BookingState;
  /** The staff user who first took the booking out of "new". */
  assignee: { id: string; name: string } | null;
  staffNote: string;
  status: boolean;
  createDate: string;
  updateDate: string;
};

export type BookingListRequest = Paging & {
  keyword?: string;
  state?: BookingState;
  status?: boolean;
};

export type BookingUpdateRequest = {
  id: string;
  state: BookingState;
  staffNote: string;
};

export type BookingUpdateStatusRequest = { id: string; status: boolean };
