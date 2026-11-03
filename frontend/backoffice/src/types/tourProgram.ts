import type { TourProgramState } from '../constants/tourProgramStates.ts';
import type { Paging } from './api.ts';

export type TourProgramWholesale = { id: string; code: string; name: string; status: boolean };
export type TourProgramAirline = { id: string; code: string; name: string; logo: string | null; status: boolean };
export type TourProgramCountry = { id: string; slug: string; nameTh: string; nameEn: string; status: boolean };
export type TourProgramCity = TourProgramCountry & { countryId: string };
export type TourProgramImage = { id: string; url: string };

export type TourProgram = {
  id: string;
  refId: number | null; // legacy id from the old site
  code: string; // public, used in /tour/<code>/; never changes
  name: string;
  wholesale: TourProgramWholesale;
  wholesaleTourCode: string; // the wholesale's own code (staff only)
  state: TourProgramState;
  days: number;
  nights: number;
  airline: TourProgramAirline | null; // null = no airline
  highlight: string;
  description: string;
  bannerImage: string | null; // tour detail page hero
  cardImage: string | null; // tour cards on home/listing pages
  pdfFile: string | null; // public itinerary
  docxFile: string | null; // staff only
  internalNote: string;
  countries: TourProgramCountry[]; // display order; first = main country
  cities: TourProgramCity[];
  images: TourProgramImage[]; // gallery, display order
  status: boolean; // false = soft-deleted
  createDate: string;
  updateDate: string;
};

export type TourProgramListItem = Pick<
  TourProgram,
  'id' | 'refId' | 'code' | 'name' | 'wholesale' | 'wholesaleTourCode' | 'state' | 'days' | 'nights' | 'bannerImage' | 'cardImage' | 'countries' | 'status' | 'createDate' | 'updateDate'
>;

// POST /tour-programs/list
export type TourProgramListRequest = Paging & {
  keyword?: string; // name, code or wholesale tour code
  countryId?: string;
  cityId?: string;
  wholesaleId?: string;
  state?: TourProgramState;
  status?: boolean; // omitted = all
};

// POST /tour-programs — countries/cities/images replace the whole lists
export type TourProgramCreateRequest = {
  refId: number | null;
  name: string;
  wholesaleId: string;
  wholesaleTourCode: string;
  state: TourProgramState;
  days: number;
  nights: number;
  airlineId: string | null; // airline master id, optional
  highlight: string;
  description: string;
  bannerImage: string | null;
  cardImage: string | null;
  pdfFile: string | null;
  docxFile: string | null;
  internalNote: string;
  countryIds: string[];
  cityIds: string[];
  images: string[]; // gallery URLs in order
};

// PUT /tour-programs
export type TourProgramUpdateRequest = TourProgramCreateRequest & { id: string };

// PUT /tour-programs/status, PUT /tour-program-periods/status
export type StatusRequest = { id: string; status: boolean };

export type TourProgramPeriod = {
  id: string;
  tourProgramId: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  price: number; // adult, whole baht
  priceBeforeDiscount: number | null;
  childWithBedPrice: number | null;
  childNoBedPrice: number | null;
  singleSupplement: number | null;
  deposit: number | null;
  seatsTotal: number;
  seatsAvailable: number; // 0 = sold out
  note: string; // staff only
  status: boolean;
  createDate: string;
  updateDate: string;
};

export type TourProgramPeriodFields = Omit<TourProgramPeriod, 'id' | 'tourProgramId' | 'status' | 'createDate' | 'updateDate'>;

// POST /tour-program-periods
export type PeriodCreateRequest = TourProgramPeriodFields & { tourProgramId: string };

// PUT /tour-program-periods
export type PeriodUpdateRequest = TourProgramPeriodFields & { id: string };

// POST /tour-program-periods/list
export type PeriodListRequest = { tourProgramId: string; status?: boolean };
