import type { ChangePasswordRequest, LoginRequest, Me, TokenResponse } from '../types/auth.ts';
import { refreshAccessToken, request } from './client.ts';

export const authApi = {
  login: (body: LoginRequest) => request<TokenResponse>('POST', '/auth/login', body, { auth: false }),
  refresh: refreshAccessToken,
  // Revokes the refresh token and clears its cookie. Works even when the access token has expired.
  logout: () => request<null>('POST', '/auth/logout', undefined, { auth: false }),
  me: () => request<Me>('GET', '/auth/me'),
  changePassword: (body: ChangePasswordRequest) => request<null>('PUT', '/auth/password', body),
};
