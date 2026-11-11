import type { Paging } from './api.ts';

export type BookingSource = {
  id: string;
  name: string; // unique, case-insensitive
  shortName: string; // label for the source's own reference name ("ชื่อ <shortName>"); '' = none
  sortOrder: number;
  status: boolean; // false = soft-deleted
  createDate: string;
  updateDate: string;
};

// POST /booking-sources/list — ordered by sortOrder, then name
export type BookingSourceListRequest = Paging & {
  keyword?: string; // matches name or short name
  status?: boolean; // omitted = all
};

// POST /booking-sources
export type BookingSourceCreateRequest = {
  name: string;
  shortName: string;
  sortOrder: number;
};

// PUT /booking-sources
export type BookingSourceUpdateRequest = BookingSourceCreateRequest & { id: string };

// PUT /booking-sources/status
export type BookingSourceStatusRequest = {
  id: string;
  status: boolean;
};
