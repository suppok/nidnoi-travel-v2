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
// GET /auth/line/config — whether "ล็อกอินด้วย LINE" shows, and the LINE Login channel.
export type LineLoginConfig = { enabled: boolean; channelId: string };

// POST /auth/line/login — the code LINE redirected back with (PKCE).
export type LineLoginRequest = {
  code: string;
  redirectUri: string;
  codeVerifier: string;
  nonce: string;
  rememberMe: boolean;
};

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
  // The user's own LINE account ("ล็อกอินด้วย LINE"); null until linked.
  line: LineAccount | null;
};

export type LineAccount = {
  displayName: string;
  linkedAt: string;
};

// POST /auth/line-link — the LINE Login callback, like LineLoginRequest (no rememberMe).
export type LineLinkRequest = Omit<LineLoginRequest, 'rememberMe'>;
export type LineLinkResponse = { line: LineAccount };

// PUT /auth/password
export type ChangePasswordRequest = {
  currentPassword: string;
  newPassword: string;
};
