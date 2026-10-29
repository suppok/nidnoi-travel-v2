import type { Permission } from '../constants/permissions.ts';
import type { Role } from '../constants/roles.ts';

// POST /auth/login
export type LoginRequest = {
  email: string;
  password: string;
  // true: refresh cookie persists across browser restarts; false: session cookie.
  rememberMe: boolean;
};

// POST /auth/login, POST /auth/refresh. The refresh token itself is an HttpOnly cookie.
export type TokenResponse = {
  accessToken: string;
  expiresIn: number; // seconds
};

// GET /auth/me
export type Me = {
  id: string;
  email: string;
  name: string;
  roles: Role[];
  permissions: Permission[];
  mustChangePassword: boolean;
};

// PUT /auth/password
export type ChangePasswordRequest = {
  currentPassword: string;
  newPassword: string;
};
