import type { ListResponse } from '../types/api.ts';
import type { LineContact, LineContactListRequest } from '../types/lineBooking.ts';
import { request } from './client.ts';

export const lineContactsApi = {
  list: (body: LineContactListRequest) => request<ListResponse<LineContact>>('POST', '/line-contacts/list', body),
};
