import { ErrorCodes } from '../../constants/errorCodes.ts';
import type { Permission } from '../../constants/permissions.ts';
import { type MockUser, users } from './db.ts';
import { permissionsFor } from './policy.ts';

export type MockRequest = {
  method: string;
  path: string;
  params: Record<string, string>;
  body: unknown;
  token: string | null;
};

export type MockHandler = (req: MockRequest) => Response;

export class MockFail extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(status: number, code: string, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

export function ok(data: unknown, status = 200): Response {
  return Response.json({ success: true, data }, { status });
}

export function fail(err: MockFail): Response {
  return Response.json(
    { success: false, error: { code: err.code, message: err.message, ...(err.fields ? { fields: err.fields } : {}) } },
    { status: err.status },
  );
}

export function validation(fields: Record<string, string>): MockFail {
  return new MockFail(400, ErrorCodes.Validation, 'Validation failed', fields);
}

// --- Tokens -------------------------------------------------------------------------------------
// Mock access token: "mock.<userId>.<sessionVersion>.<expiresAtMs>". Not a JWT; never leaves the mock.

export const ACCESS_TTL_SECONDS = 15 * 60;

export function issueAccessToken(user: MockUser): string {
  return `mock.${user.id}.${user.sessionVersion}.${Date.now() + ACCESS_TTL_SECONDS * 1000}`;
}

// Routes a user who must change their password may still call.
const PASSWORD_CHANGE_ALLOWED = new Set(['GET /auth/me', 'PUT /auth/password']);

export function authenticate(req: MockRequest, route: string): MockUser {
  const [prefix, id, version, exp] = (req.token ?? '').split('.');
  const user = users.find((u) => u.id === id);
  if (prefix !== 'mock' || !user || !user.status || Number(version) !== user.sessionVersion || Number(exp) < Date.now()) {
    throw new MockFail(401, ErrorCodes.Unauthenticated, 'Authentication required');
  }
  if (user.mustChangePassword && !PASSWORD_CHANGE_ALLOWED.has(route)) {
    throw new MockFail(403, ErrorCodes.PasswordChangeRequired, 'Password change required');
  }
  return user;
}

export function authorize(req: MockRequest, route: string, permission: Permission): MockUser {
  const user = authenticate(req, route);
  if (!permissionsFor(user.roles).includes(permission)) {
    throw new MockFail(403, ErrorCodes.PermissionDenied, 'Permission denied');
  }
  return user;
}

// --- Refresh "cookie" ---------------------------------------------------------------------------
// The real refresh token is an HttpOnly cookie. The mock imitates it with web storage:
// localStorage for "remember me" (persistent cookie), sessionStorage otherwise (session cookie).

const COOKIE_KEY = 'nidnoi.mock.refresh';

type RefreshCookie = { userId: string; sessionVersion: number; persistent: boolean };

export function setRefreshCookie(user: MockUser, persistent: boolean) {
  clearRefreshCookie();
  const value: RefreshCookie = { userId: user.id, sessionVersion: user.sessionVersion, persistent };
  (persistent ? localStorage : sessionStorage).setItem(COOKIE_KEY, JSON.stringify(value));
}

export function readRefreshCookie(): RefreshCookie | null {
  const raw = sessionStorage.getItem(COOKIE_KEY) ?? localStorage.getItem(COOKIE_KEY);
  try {
    return raw ? (JSON.parse(raw) as RefreshCookie) : null;
  } catch {
    return null;
  }
}

export function clearRefreshCookie() {
  sessionStorage.removeItem(COOKIE_KEY);
  localStorage.removeItem(COOKIE_KEY);
}

// --- Validation helpers -------------------------------------------------------------------------

export function asRecord(body: unknown): Record<string, unknown> {
  return body && typeof body === 'object' ? (body as Record<string, unknown>) : {};
}

export function str(v: unknown): string {
  return typeof v === 'string' ? v.trim() : '';
}

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
