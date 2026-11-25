// โปรไฟไหม้ on the home page: departures staff flagged in the backoffice (per period, with an order).
import type { TourFireSalePublic, TourProgramPublicItem } from '../types/tour-program';

export const FIRE_SALE_LIMIT = 4;

export type FireSaleEntry = { tour: TourProgramPublicItem; sale: TourFireSalePublic };

/** Every tour's fire sales, by staff order, then departure date; at most `limit`. */
export function pickFireSales(tours: TourProgramPublicItem[], limit = FIRE_SALE_LIMIT): FireSaleEntry[] {
  return tours
    .flatMap((tour) => (tour.fireSales ?? []).map((sale) => ({ tour, sale })))
    .sort(
      (a, b) =>
        a.sale.order - b.sale.order ||
        a.sale.startDate.localeCompare(b.sale.startDate) ||
        a.tour.code.localeCompare(b.tour.code),
    )
    .slice(0, limit);
}
