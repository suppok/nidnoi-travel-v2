import type { Airline, AirlineCreateRequest, AirlineListRequest, AirlineStatusRequest, AirlineUpdateRequest } from '../types/airline.ts';
import type { ListResponse } from '../types/api.ts';
import { request } from './client.ts';

export const airlinesApi = {
  list: (body: AirlineListRequest) => request<ListResponse<Airline>>('POST', '/airlines/list', body),
  get: (id: string) => request<Airline>('GET', `/airlines/${encodeURIComponent(id)}`),
  create: (body: AirlineCreateRequest) => request<Airline>('POST', '/airlines', body),
  update: (body: AirlineUpdateRequest) => request<Airline>('PUT', '/airlines', body),
  updateStatus: (body: AirlineStatusRequest) => request<null>('PUT', '/airlines/status', body),
};
