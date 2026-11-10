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
  // The user's own LINE account, linked for the LIFF booking app; null until linked.
  line: LineAccount | null;
};

export type LineAccount = {
  displayName: string;
  linkedAt: string;
};

// POST /auth/line-link — a one-time LIFF URL shown as a QR code.
export type LineLinkResponse = {
  url: string;
  expiresAt: string;
};

// PUT /auth/password
export type ChangePasswordRequest = {
  currentPassword: string;
  newPassword: string;
};
