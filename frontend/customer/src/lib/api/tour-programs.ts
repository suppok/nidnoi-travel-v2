// Build-time tour reads. Browser code must use ./periods.ts instead (no fixtures in the bundle).
import type { TourProgramPublicDetail, TourProgramPublicItem } from '../../types/tour-program';
import { apiGet } from './client';
import { USE_MOCK_DATA } from './mock-mode';

const PAGE_SIZE = 100;

type Page<T> = { items: T[]; total: number };

let toursOnce: Promise<TourProgramPublicItem[]> | undefined;

/** Every public tour (paged through). Fetched once per build. */
export function listPublicTours(): Promise<TourProgramPublicItem[]> {
  toursOnce ??= fetchAllTours();
  return toursOnce;
}

async function fetchAllTours(): Promise<TourProgramPublicItem[]> {
  if (USE_MOCK_DATA) {
    const { mockTours } = await import('./mock-data');
    return mockTours;
  }
  const all: TourProgramPublicItem[] = [];
  for (let page = 1; ; page++) {
    const res = await apiGet<Page<TourProgramPublicItem>>(`/tour-programs/public?page=${page}&pageSize=${PAGE_SIZE}`);
    all.push(...res.items);
    if (all.length >= res.total || res.items.length === 0) return all;
  }
}

export async function getPublicTour(code: string): Promise<TourProgramPublicDetail> {
  if (USE_MOCK_DATA) {
    const { mockTours } = await import('./mock-data');
    const tour = mockTours.find((t) => t.code === code);
    if (!tour) throw new Error(`mock: tour ${code} not found`);
    return tour;
  }
  return apiGet<TourProgramPublicDetail>(`/tour-programs/public/${encodeURIComponent(code)}`);
}

/** Tours sorted for display: soonest next departure first, undated last. */
export function byNextDeparture(a: TourProgramPublicItem, b: TourProgramPublicItem): number {
  return (a.nextDepartureDate ?? '9999').localeCompare(b.nextDepartureDate ?? '9999');
}
