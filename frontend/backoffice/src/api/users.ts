import type { ListResponse } from '../types/api.ts';
import type {
  User,
  UserCreateRequest,
  UserListRequest,
  UserPasswordResetRequest,
  UserStatusRequest,
  UserUpdateRequest,
} from '../types/user.ts';
import { request } from './client.ts';

export const usersApi = {
  list: (body: UserListRequest) => request<ListResponse<User>>('POST', '/users/list', body),
  get: (id: string) => request<User>('GET', `/users/${encodeURIComponent(id)}`),
  create: (body: UserCreateRequest) => request<User>('POST', '/users', body),
  update: (body: UserUpdateRequest) => request<User>('PUT', '/users', body),
  updateStatus: (body: UserStatusRequest) => request<null>('PUT', '/users/status', body),
  resetPassword: (body: UserPasswordResetRequest) => request<null>('PUT', '/users/password', body),
};
