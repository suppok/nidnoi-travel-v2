// Browser call used by the PeriodTable island. Cached ~30 s at the edge.
import type { TourPeriodPublic } from '../../types/tour-program';
import { apiGet } from './client';
import { USE_MOCK_DATA } from './mock-mode';

/** Returns null in mock mode so the island keeps its build-time rows. */
export async function fetchPublicPeriods(code: string, signal?: AbortSignal): Promise<TourPeriodPublic[] | null> {
  if (USE_MOCK_DATA) return null;
  return apiGet<TourPeriodPublic[]>(`/tour-programs/public/${encodeURIComponent(code)}/periods`, signal);
}
