// Mirrors backend POST /bookings/public (pkg/modules/booking/models/booking_public_create_model.go).

/** Travellers per price row. A row the period doesn't price must stay 0. */
export type BookingCounts = {
  adultTwin: number;
  adultSingle: number;
  childWithBed: number;
  childNoBed: number;
};

export type BookingPublicCreateRequest = BookingCounts & {
  tourCode: string;
  periodId: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  note: string;
  /** Honeypot: hidden from people, so it is always "" from a real visitor. */
  website: string;
};

export type BookingPublicCreateResponse = {
  bookingNo: string;
  /** Whole baht, quoted by the backend. */
  totalPrice: number;
};
