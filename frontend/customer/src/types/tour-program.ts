// Public tour DTOs. Mirrors the planned backend responses for
// GET /tour-programs/public, /tour-programs/public/:code and /tour-programs/public/:code/periods.
import type { CityRef, TaxonomyRef } from './taxonomy';

export type AirlinePublic = {
  code: string;
  name: string;
  logo: string | null;
};

/** One departure. Prices are whole Thai baht. Dates are ISO `YYYY-MM-DD`. */
export type TourPeriodPublic = {
  id: string;
  startDate: string;
  endDate: string;
  price: number;
  priceBeforeDiscount: number | null;
  childWithBedPrice: number | null;
  childNoBedPrice: number | null;
  singleSupplement: number | null;
  deposit: number | null;
  seatsAvailable: number;
  soldOut: boolean;
};

/** One day of the short timeline: places in visiting order ("โรงแรม" last when the night is at a hotel). */
export type TourItineraryDay = { day: number; places: string[] };

/** Card / listing shape. */
export type TourProgramPublicItem = {
  code: string;
  name: string;
  /** Short summary for cards and the meta description. */
  highlight: string;
  days: number;
  nights: number;
  airline: AirlinePublic | null;
  cardImage: string | null;
  countries: TaxonomyRef[];
  cities: CityRef[];
  /** Lowest active period price ("ราคาเริ่มต้น"); null when there is no priced period. */
  startingPrice: number | null;
  /** Struck-through price of the period that sets startingPrice. */
  startingPriceBeforeDiscount: number | null;
  nextDepartureDate: string | null;
  periodCount: number;
  updateDate: string;
};

export type TourProgramPublicDetail = TourProgramPublicItem & {
  description: string;
  bannerImage: string | null;
  images: string[];
  pdfFile: string | null;
  periods: TourPeriodPublic[];
  /** ลงร้านรัฐบาล: true = stops at government shops, false = it doesn't, null = not specified. */
  governmentShopStops: boolean | null;
  /** What the price includes / doesn't, in display order (may be empty). */
  includedItems: string[];
  excludedItems: string[];
  /** Day-by-day places (may be empty). */
  itinerary: TourItineraryDay[];
};
