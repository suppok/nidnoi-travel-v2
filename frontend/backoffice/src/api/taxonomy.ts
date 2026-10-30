import type { ListResponse } from '../types/api.ts';
import type {
  City,
  CityCreateRequest,
  CityListRequest,
  CityUpdateRequest,
  Country,
  CountryCreateRequest,
  CountryListRequest,
  CountryUpdateRequest,
  TaxonomyStatusRequest,
} from '../types/taxonomy.ts';
import { request } from './client.ts';

export const countriesApi = {
  list: (body: CountryListRequest) => request<ListResponse<Country>>('POST', '/countries/list', body),
  get: (id: string) => request<Country>('GET', `/countries/${encodeURIComponent(id)}`),
  create: (body: CountryCreateRequest) => request<Country>('POST', '/countries', body),
  update: (body: CountryUpdateRequest) => request<Country>('PUT', '/countries', body),
  updateStatus: (body: TaxonomyStatusRequest) => request<null>('PUT', '/countries/status', body),
};

export const citiesApi = {
  list: (body: CityListRequest) => request<ListResponse<City>>('POST', '/cities/list', body),
  get: (id: string) => request<City>('GET', `/cities/${encodeURIComponent(id)}`),
  create: (body: CityCreateRequest) => request<City>('POST', '/cities', body),
  update: (body: CityUpdateRequest) => request<City>('PUT', '/cities', body),
  updateStatus: (body: TaxonomyStatusRequest) => request<null>('PUT', '/cities/status', body),
};
