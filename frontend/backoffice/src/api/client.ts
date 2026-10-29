import { authBus, tokenStore } from '../auth/token.ts';
import { env } from '../config/env.ts';
import { ErrorCodes } from '../constants/errorCodes.ts';
import type { TokenResponse } from '../types/auth.ts';

export class ApiError extends Error {
  code: string;
  status: number;
  fields?: Record<string, string>;

  constructor(code: string, message: string, status: number, fields?: Record<string, string>) {
    super(message);
    this.code = code;
    this.status = status;
    this.fields = fields;
  }
}

type RequestOptions = {
  // false for endpoints that must not attach the access token or trigger a refresh (login, refresh).
  auth?: boolean;
};

// The mock is loaded lazily so it is never downloaded when VITE_API_MOCK is off.
const doFetch: typeof fetch = env.API_MOCK
  ? (...args) => import('./mock/index.ts').then((m) => m.mockFetch(...args))
  : (...args) => fetch(...args);

async function send(method: string, path: string, body: unknown, token: string | null): Promise<Response> {
  const isForm = body instanceof FormData;
  try {
    return await doFetch(`${env.API_BASE_URL}${path}`, {
      method,
      // The refresh token cookie is sent to api.* from admin.* (same site, cross origin).
      credentials: 'include',
      headers: {
        ...(isForm ? {} : { 'Content-Type': 'application/json' }),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body === undefined ? undefined : isForm ? body : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(ErrorCodes.Network, 'Cannot reach the server', 0);
  }
}

async function parse<T>(res: Response): Promise<T> {
  const json = await res.json().catch(() => null);
  if (!res.ok || json?.success === false) {
    throw new ApiError(
      json?.error?.code ?? ErrorCodes.Unknown,
      json?.error?.message ?? res.statusText,
      res.status,
      json?.error?.fields,
    );
  }
  return json.data as T;
}

let refreshing: Promise<string | null> | null = null;

// Exchanges the refresh cookie for a new access token. Concurrent callers share one request,
// because the backend rotates the refresh token on every use.
export function refreshAccessToken(): Promise<string | null> {
  refreshing ??= (async () => {
    try {
      const res = await send('POST', '/auth/refresh', undefined, null);
      const data = await parse<TokenResponse>(res);
      tokenStore.set(data.accessToken);
      return data.accessToken;
    } catch {
      tokenStore.clear();
      return null;
    } finally {
      refreshing = null;
    }
  })();
  return refreshing;
}

export async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  { auth = true }: RequestOptions = {},
): Promise<T> {
  let res = await send(method, path, body, auth ? tokenStore.get() : null);

  if (res.status === 401 && auth) {
    const token = await refreshAccessToken();
    if (token) {
      res = await send(method, path, body, token);
    }
    if (!token || res.status === 401) {
      tokenStore.clear();
      authBus.emitUnauthenticated();
      throw new ApiError(ErrorCodes.Unauthenticated, 'Session expired', 401);
    }
  }

  return parse<T>(res);
}
