import { ErrorCodes } from '../../constants/errorCodes.ts';
import { passwordProblem } from '../../lib/password.ts';
import type { Me, TokenResponse } from '../../types/auth.ts';
import { save, users } from './db.ts';
import {
  ACCESS_TTL_SECONDS,
  asRecord,
  authenticate,
  clearRefreshCookie,
  issueAccessToken,
  type MockHandler,
  MockFail,
  ok,
  readRefreshCookie,
  setRefreshCookie,
  str,
  validation,
} from './http.ts';
import { permissionsFor } from './policy.ts';

const login: MockHandler = (req) => {
  const body = asRecord(req.body);
  const email = str(body.email).toLowerCase();
  const password = typeof body.password === 'string' ? body.password : '';
  const fields: Record<string, string> = {};
  if (!email) fields.email = 'required';
  if (!password) fields.password = 'required';
  if (Object.keys(fields).length) throw validation(fields);

  const user = users.find((u) => u.email === email);
  // Same error for unknown email, wrong password and deactivated user — don't reveal which.
  if (!user || !user.status || user.password !== password) {
    throw new MockFail(401, ErrorCodes.InvalidCredentials, 'Invalid email or password');
  }

  user.lastLoginAt = new Date().toISOString();
  save();
  setRefreshCookie(user, body.rememberMe === true);
  return ok({ accessToken: issueAccessToken(user), expiresIn: ACCESS_TTL_SECONDS } satisfies TokenResponse);
};

const refresh: MockHandler = () => {
  const cookie = readRefreshCookie();
  const user = cookie && users.find((u) => u.id === cookie.userId);
  if (!cookie || !user || !user.status || user.sessionVersion !== cookie.sessionVersion) {
    clearRefreshCookie();
    throw new MockFail(401, ErrorCodes.Unauthenticated, 'Refresh token invalid or expired');
  }
  setRefreshCookie(user, cookie.persistent); // rotation
  return ok({ accessToken: issueAccessToken(user), expiresIn: ACCESS_TTL_SECONDS } satisfies TokenResponse);
};

const logout: MockHandler = () => {
  clearRefreshCookie();
  return ok(null);
};

const me: MockHandler = (req) => {
  const user = authenticate(req, 'GET /auth/me');
  return ok({
    id: user.id,
    email: user.email,
    name: user.name,
    roles: user.roles,
    permissions: permissionsFor(user.roles),
    mustChangePassword: user.mustChangePassword,
  } satisfies Me);
};

const changePassword: MockHandler = (req) => {
  const user = authenticate(req, 'PUT /auth/password');
  const body = asRecord(req.body);
  const currentPassword = typeof body.currentPassword === 'string' ? body.currentPassword : '';
  const newPassword = typeof body.newPassword === 'string' ? body.newPassword : '';

  if (currentPassword !== user.password) {
    throw new MockFail(400, ErrorCodes.InvalidPassword, 'Current password is incorrect');
  }
  const problem = passwordProblem(newPassword);
  if (problem) throw validation({ newPassword: problem });
  if (newPassword === currentPassword) throw validation({ newPassword: 'ต้องไม่ซ้ำกับรหัสผ่านเดิม' });

  user.password = newPassword;
  user.mustChangePassword = false;
  user.updateDate = new Date().toISOString();
  save();
  return ok(null);
};

export const authRoutes: Record<string, MockHandler> = {
  'POST /auth/login': login,
  'POST /auth/refresh': refresh,
  'POST /auth/logout': logout,
  'GET /auth/me': me,
  'PUT /auth/password': changePassword,
};
