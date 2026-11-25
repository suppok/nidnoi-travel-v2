// Departure-month and price filters shared by the home hero (builds the link) and the landing tour list
// (filters the cards client-side from ?month=YYYY-MM&price=min-max). Dependency-free: also runs in the browser.
import type { TourPeriodPublic } from '../types/tour-program';
import { formatPrice } from './format';

export const PRICE_RANGES = [
  { value: '0-20000', label: `ไม่เกิน ${formatPrice(20000)} บาท` },
  { value: '20000-40000', label: `${formatPrice(20000)} – ${formatPrice(40000)} บาท` },
  { value: '40000-70000', label: `${formatPrice(40000)} – ${formatPrice(70000)} บาท` },
  { value: '70000-', label: `${formatPrice(70000)} บาทขึ้นไป` },
] as const;

/** "20000-40000" → [20000, 40000]; "70000-" → [70000, Infinity]; anything else → null (no price filter). */
export function parsePriceRange(value: string | null): [number, number] | null {
  const m = /^(\d+)-(\d*)$/.exec(value ?? '');
  return m ? [Number(m[1]), m[2] ? Number(m[2]) : Infinity] : null;
}

export function priceRangeLabel(value: string): string | null {
  return PRICE_RANGES.find((r) => r.value === value)?.label ?? null;
}

/** Today in Bangkok as `YYYY-MM-DD`. */
export function todayIso(): string {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Bangkok' }).format(new Date());
}

/** Departures a customer can still pick: not sold out and not yet departed (as of the build). */
function openPeriods(periods: TourPeriodPublic[], today: string): TourPeriodPublic[] {
  return periods.filter((p) => !p.soldOut && p.startDate >= today);
}

/** Months (`YYYY-MM`) with an open departure, ascending. */
export function departureMonths(periods: TourPeriodPublic[], today = todayIso()): string[] {
  return [...new Set(openPeriods(periods, today).map((p) => p.startDate.slice(0, 7)))].sort();
}

/** Open departures encoded for a data attribute: "2026-12:9499,12999 2027-01:15900" (month: prices). */
export function encodeDepartures(periods: TourPeriodPublic[], today = todayIso()): string {
  const byMonth = new Map<string, Set<number>>();
  for (const p of openPeriods(periods, today)) {
    const key = p.startDate.slice(0, 7);
    const prices = byMonth.get(key) ?? new Set<number>();
    if (p.price > 0) prices.add(p.price);
    byMonth.set(key, prices);
  }
  return [...byMonth]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, prices]) => `${key}:${[...prices].join(',')}`)
    .join(' ');
}

/** True when one departure matches both the month (if any) and the price range (if any). */
export function matchesDepartures(encoded: string, month: string, range: [number, number] | null): boolean {
  return encoded
    .split(' ')
    .filter(Boolean)
    .some((token) => {
      const [key, list] = token.split(':');
      if (month && key !== month) return false;
      if (!range) return true;
      return list
        .split(',')
        .filter(Boolean)
        .some((p) => Number(p) >= range[0] && Number(p) <= range[1]);
    });
}
