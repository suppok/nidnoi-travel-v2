import type { SalesFinanceSettings, SalesFinanceUpdateRequest } from '../types/salesFinance.ts';
import { request } from './client.ts';

export const salesFinancesApi = {
  get: () => request<SalesFinanceSettings>('GET', '/sales-finances'),
  update: (body: SalesFinanceUpdateRequest) => request<SalesFinanceSettings>('PUT', '/sales-finances', body),
};
