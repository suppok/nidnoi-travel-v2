import type { DashboardDays, DashboardSummary } from '../types/dashboard.ts';
import { request } from './client.ts';

export const dashboardApi = {
  summary: (days: DashboardDays) => request<DashboardSummary>('POST', '/dashboard/summary', { days }),
};
