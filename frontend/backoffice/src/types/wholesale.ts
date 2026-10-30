import type { WholesaleState } from '../constants/wholesaleStates.ts';
import type { Paging } from './api.ts';

export type WholesaleContact = {
  id: string;
  name: string;
  phone: string;
};

export type Wholesale = {
  id: string;
  refId: number | null; // legacy id from the old site
  code: string; // uppercase, unique
  name: string;
  state: WholesaleState;
  contacts: WholesaleContact[]; // display order
  status: boolean; // false = soft-deleted
  createDate: string;
  updateDate: string;
};

// POST /wholesales/list
export type WholesaleListRequest = Paging & {
  keyword?: string; // matches name, code, or a contact's name/phone
  state?: WholesaleState;
  status?: boolean; // omitted = all
};

export type WholesaleContactInput = Pick<WholesaleContact, 'name' | 'phone'>;

// POST /wholesales — state defaults to on request
export type WholesaleCreateRequest = {
  refId: number | null;
  code: string;
  name: string;
  state: WholesaleState;
  contacts: WholesaleContactInput[]; // max 20, replaces the whole list
};

// PUT /wholesales
export type WholesaleUpdateRequest = WholesaleCreateRequest & { id: string };

// PUT /wholesales/status
export type WholesaleStatusRequest = {
  id: string;
  status: boolean;
};
