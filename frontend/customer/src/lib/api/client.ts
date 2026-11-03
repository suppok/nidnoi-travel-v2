import { env } from '../../config/env';

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
    public fields?: Record<string, string>,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

type Envelope<T> =
  | { success: true; data: T }
  | { success: false; error: { code: string; message: string; fields?: Record<string, string> } };

// Public endpoints only — never attach an Authorization header.
async function request<T>(method: 'GET' | 'POST', path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
  const init: RequestInit = { method, headers: { 'Content-Type': 'application/json' }, signal };
  if (method === 'POST') init.body = JSON.stringify(body);
  const res = await fetch(`${env.API_BASE_URL}${path}`, init);
  const json = (await res.json().catch(() => null)) as Envelope<T> | null;
  if (!res.ok || !json || json.success === false) {
    const err = json && json.success === false ? json.error : undefined;
    throw new ApiError(err?.code ?? 'UNKNOWN', err?.message ?? res.statusText, res.status, err?.fields);
  }
  return json.data;
}

export const apiGet = <T>(path: string, signal?: AbortSignal) => request<T>('GET', path, undefined, signal);
export const apiPost = <T>(path: string, body: unknown) => request<T>('POST', path, body);
