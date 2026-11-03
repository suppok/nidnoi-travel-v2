import type { Paging } from './api.ts';

export type Airline = {
  id: string;
  refId: number | null; // legacy id from the old site
  code: string; // IATA (TG) or ICAO (THA), uppercase, unique
  name: string;
  logo: string | null; // URL from POST /media/upload (airline-logo)
  status: boolean; // false = soft-deleted
  createDate: string;
  updateDate: string;
};

// POST /airlines/list
export type AirlineListRequest = Paging & {
  keyword?: string; // matches name or code
  status?: boolean; // omitted = all
};

// POST /airlines
export type AirlineCreateRequest = {
  refId: number | null;
  code: string;
  name: string;
  logo: string | null;
};

// PUT /airlines
export type AirlineUpdateRequest = AirlineCreateRequest & { id: string };

// PUT /airlines/status
export type AirlineStatusRequest = {
  id: string;
  status: boolean;
};
