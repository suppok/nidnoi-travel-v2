import type { Role } from '../constants/roles.ts';
import type { Paging } from './api.ts';

export type User = {
  id: string;
  email: string;
  name: string;
  phone: string; // digits only, '' = none
  roles: Role[]; // at least one
  status: boolean; // false = deactivated (soft delete)
  mustChangePassword: boolean;
  lastLoginAt: string | null;
  lineDisplayName: string | null; // LINE account linked for the LIFF app
  createDate: string;
  updateDate: string;
};

// POST /users/list
export type UserListRequest = Paging & {
  keyword?: string; // matches name or email
  role?: Role;
  status?: boolean; // omitted = all
};

// POST /users
export type UserCreateRequest = {
  email: string;
  name: string;
  phone: string; // optional ('')
  roles: Role[];
  temporaryPassword: string; // user must change it at first login
};

// PUT /users
export type UserUpdateRequest = {
  id: string;
  email: string;
  name: string;
  phone: string;
  roles: Role[];
};

// PUT /users/status
export type UserStatusRequest = {
  id: string;
  status: boolean;
};

// PUT /users/password — admin reset; signs the user out everywhere.
export type UserPasswordResetRequest = {
  id: string;
  temporaryPassword: string;
};

// GET /users/sales — sales person options for the sales tools.
export type SalesUser = { id: string; name: string; phone: string };
export type SalesUserListResponse = {
  items: SalesUser[]; // active sales users, by name
  defaultId: string; // the caller when they are one of them, else ''
};
