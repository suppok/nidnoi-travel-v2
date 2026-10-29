// In-browser mock of the backend API, enabled with VITE_API_MOCK=true. Handlers follow the real
// contracts (paths, status codes, error codes), so switching to the backend changes only client.ts.
import { env } from '../../config/env.ts';
import { ErrorCodes } from '../../constants/errorCodes.ts';
import { authRoutes } from './auth.ts';
import { fail, type MockHandler, MockFail } from './http.ts';
import { userRoutes } from './users.ts';

const routes: Record<string, MockHandler> = { ...authRoutes, ...userRoutes };

const LATENCY_MS = 300;

function match(method: string, path: string): { handler: MockHandler; params: Record<string, string> } | null {
  for (const [key, handler] of Object.entries(routes)) {
    const [m, pattern] = key.split(' ');
    if (m !== method) continue;
    const names: string[] = [];
    const re = new RegExp(
      `^${pattern.replace(/:(\w+)/g, (_, name: string) => {
        names.push(name);
        return '([^/]+)';
      })}$`,
    );
    const found = re.exec(path);
    if (found) {
      return { handler, params: Object.fromEntries(names.map((n, i) => [n, decodeURIComponent(found[i + 1])])) };
    }
  }
  return null;
}

export async function mockFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  await new Promise((r) => setTimeout(r, LATENCY_MS));

  const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
  const method = (init?.method ?? 'GET').toUpperCase();
  const path = url.slice(env.API_BASE_URL.length).split('?')[0];
  const headers = new Headers(init?.headers);
  const token = headers.get('Authorization')?.replace(/^Bearer /, '') ?? null;
  const body = typeof init?.body === 'string' ? JSON.parse(init.body) : undefined;

  const found = match(method, path);
  if (!found) return fail(new MockFail(404, ErrorCodes.NotFound, `Mock has no route for ${method} ${path}`));

  try {
    return found.handler({ method, path, params: found.params, body, token });
  } catch (err) {
    if (err instanceof MockFail) return fail(err);
    throw err;
  }
}
