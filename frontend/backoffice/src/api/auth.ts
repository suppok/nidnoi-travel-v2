import type { ChangePasswordRequest, LineLinkRequest, LineLinkResponse, LineLoginConfig, LineLoginRequest, LoginRequest, Me, TokenResponse } from '../types/auth.ts';
import { refreshAccessToken, request } from './client.ts';

export const authApi = {
  login: (body: LoginRequest) => request<TokenResponse>('POST', '/auth/login', body, { auth: false }),
  lineConfig: () => request<LineLoginConfig>('GET', '/auth/line/config', undefined, { auth: false }),
  // Only a LINE account linked to an active user signs in; otherwise LINE_NOT_LINKED.
  lineLogin: (body: LineLoginRequest) => request<TokenResponse>('POST', '/auth/line/login', body, { auth: false }),
  refresh: refreshAccessToken,
  // Revokes the refresh token and clears its cookie. Works even when the access token has expired.
  logout: () => request<null>('POST', '/auth/logout', undefined, { auth: false }),
  me: () => request<Me>('GET', '/auth/me'),
  // Links the caller's LINE account with the code LINE Login sent back (LINE_ALREADY_LINKED when another user has it).
  lineLink: (body: LineLinkRequest) => request<LineLinkResponse>('POST', '/auth/line-link', body),
  lineUnlink: () => request<null>('POST', '/auth/line-unlink'),
  changePassword: (body: ChangePasswordRequest) => request<null>('PUT', '/auth/password', body),
};
