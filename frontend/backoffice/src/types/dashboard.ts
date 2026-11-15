// POST /dashboard/summary — mirrors backend pkg/modules/dashboard/models.

export const DASHBOARD_DAYS = [7, 30, 90, 365] as const;
export type DashboardDays = (typeof DASHBOARD_DAYS)[number];

export type DashboardUnit = 'day' | 'week' | 'month';

export type DashboardRange = {
  from: string; // YYYY-MM-DD, start of the first bucket
  to: string; // today
  days: number;
  unit: DashboardUnit;
  previousFrom: string;
  previousTo: string;
};

export type DashboardStateCount = { state: number; count: number };

export type DashboardBucket = {
  date: string; // first day of the bucket
  count: number;
  revenue: number; // LINE bookings only
};

export type DashboardTours = {
  byState: DashboardStateCount[];
  byCountry: { countryId: string; nameTh: string; count: number }[];
  seats: {
    days: number;
    lowSeatsThreshold: number;
    periods: number;
    seatsTotal: number;
    seatsAvailable: number;
    soldOut: number;
    lowSeats: number;
  };
  lowSeatPeriods: DashboardPeriod[];
};

export type DashboardPeriod = {
  periodId: string;
  tourProgramId: string;
  tourCode: string;
  tourName: string;
  startDate: string;
  endDate: string;
  price: number;
  seatsTotal: number;
  seatsAvailable: number;
};

export type DashboardBookings = {
  total: number;
  previousTotal: number;
  newCount: number;
  byState: DashboardStateCount[];
  buckets: DashboardBucket[];
};

export type DashboardLineTotals = {
  count: number;
  paidCount: number;
  cancelled: number;
  revenue: number;
  travellers: number;
  averageValue: number;
};

export type DashboardLineGroup = {
  key: string; // '' = not set
  label: string;
  count: number;
  paidCount: number;
  revenue: number;
  travellers: number;
};

export type DashboardDeparture = {
  id: string;
  bookingNo: string;
  tourCode: string;
  tourName: string;
  startDate: string;
  endDate: string;
  customerName: string;
  travellers: number;
  state: number;
  totalPrice: number;
  salesUser: { id: string; name: string } | null;
};

export type DashboardLineBookings = {
  scopeAll: boolean;
  totals: DashboardLineTotals;
  previous: DashboardLineTotals;
  buckets: DashboardBucket[];
  byState: DashboardStateCount[];
  bySource: DashboardLineGroup[];
  byTour: DashboardLineGroup[];
  bySales: DashboardLineGroup[];
  pipeline: { waitingConfirm: number; waitingDocuments: number; waitingReview: number; balancePending: number };
  departuresDays: number;
  departures: DashboardDeparture[];
};

export type DashboardSummary = {
  range: DashboardRange;
  generatedAt: string;
  tours: DashboardTours | null;
  bookings: DashboardBookings | null;
  lineBookings: DashboardLineBookings | null;
};
