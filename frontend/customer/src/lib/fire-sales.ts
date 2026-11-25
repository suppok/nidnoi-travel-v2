// โปรไฟไหม้ on the home page: departures staff flagged in the backoffice (per period, with an order).
import type { TourFireSalePublic, TourProgramPublicItem } from '../types/tour-program';
import { byNextDeparture, getPublicTour } from './api/tour-programs';
import { todayIso } from './tour-filter';

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

/**
 * Preview-only โปรไฟไหม้ for local/dev while staff haven't flagged any departure: the next open departure of
 * up to `limit` real tours (soonest first), with a made-up ~20 % discount when the period has none.
 * The prices are fake — callers must never use this on prd.
 * TODO(content): delete once staff flag real โปรไฟไหม้ departures on dev.
 */
export async function mockFireSales(tours: TourProgramPublicItem[], limit = FIRE_SALE_LIMIT): Promise<FireSaleEntry[]> {
  const today = todayIso();
  const entries: FireSaleEntry[] = [];
  for (const tour of [...tours].sort(byNextDeparture)) {
    if (entries.length >= limit) break;
    const period = (await getPublicTour(tour.code)).periods
      .filter((p) => !p.soldOut && p.price > 0 && p.startDate >= today)
      .sort((a, b) => a.startDate.localeCompare(b.startDate))[0];
    if (!period) continue;
    const before =
      period.priceBeforeDiscount && period.priceBeforeDiscount > period.price
        ? period.priceBeforeDiscount
        : Math.ceil((period.price * 1.25) / 100) * 100;
    entries.push({ tour, sale: { ...period, priceBeforeDiscount: before, fireSale: true, order: entries.length + 1 } });
  }
  return entries;
}
