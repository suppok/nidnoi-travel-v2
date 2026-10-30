import type { ListResponse } from '../types/api.ts';
import type { Wholesale, WholesaleCreateRequest, WholesaleListRequest, WholesaleStatusRequest, WholesaleUpdateRequest } from '../types/wholesale.ts';
import { request } from './client.ts';

export const wholesalesApi = {
  list: (body: WholesaleListRequest) => request<ListResponse<Wholesale>>('POST', '/wholesales/list', body),
  get: (id: string) => request<Wholesale>('GET', `/wholesales/${encodeURIComponent(id)}`),
  create: (body: WholesaleCreateRequest) => request<Wholesale>('POST', '/wholesales', body),
  update: (body: WholesaleUpdateRequest) => request<Wholesale>('PUT', '/wholesales', body),
  updateStatus: (body: WholesaleStatusRequest) => request<null>('PUT', '/wholesales/status', body),
};
