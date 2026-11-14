// "ล็อกอินด้วย LINE": LINE Login (authorization code + PKCE). The browser goes to LINE's authorize
// page; LINE sends it back to LINE_CALLBACK_PATH with a code, which the backend exchanges (it holds
// the channel secret) and turns into a normal session — only for a LINE account a user linked.
//
// The state / nonce / PKCE verifier live in sessionStorage for the few seconds of the round trip and
// are removed on return. They are one-time values, not credentials: the session itself stays the
// in-memory access token + HttpOnly refresh cookie.

export const LINE_CALLBACK_PATH = '/login/line/callback';

const AUTHORIZE_URL = 'https://access.line.me/oauth2/v2.1/authorize';
const STORAGE_KEY = 'nidnoi.lineLogin';

type Pending = { state: string; nonce: string; verifier: string; rememberMe: boolean; returnTo: string };

function randomString(bytes = 32): string {
  return base64Url(crypto.getRandomValues(new Uint8Array(bytes)));
}

function base64Url(data: Uint8Array): string {
  let s = '';
  for (const b of data) s += String.fromCharCode(b);
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

async function codeChallenge(verifier: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
  return base64Url(new Uint8Array(digest));
}

export const lineRedirectUri = () => `${window.location.origin}${LINE_CALLBACK_PATH}`;

// Sends the browser to LINE. returnTo is where to land after signing in.
export async function startLineLogin(channelId: string, rememberMe: boolean, returnTo: string): Promise<void> {
  const pending: Pending = { state: randomString(), nonce: randomString(), verifier: randomString(48), rememberMe, returnTo };
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(pending));
  const params = new URLSearchParams({
    response_type: 'code',
    client_id: channelId,
    redirect_uri: lineRedirectUri(),
    state: pending.state,
    scope: 'profile openid',
    nonce: pending.nonce,
    code_challenge: await codeChallenge(pending.verifier),
    code_challenge_method: 'S256',
  });
  window.location.assign(`${AUTHORIZE_URL}?${params}`);
}

// The pending login for the state LINE returned; null when the state doesn't match (an old or
// forged callback). Doesn't remove it: call clearPendingLineLogin once the code is used.
export function peekPendingLineLogin(state: string | null): Pending | null {
  const raw = sessionStorage.getItem(STORAGE_KEY);
  if (!raw || !state) return null;
  try {
    const pending = JSON.parse(raw) as Pending;
    return pending.state === state ? pending : null;
  } catch {
    return null;
  }
}

export function clearPendingLineLogin(): void {
  sessionStorage.removeItem(STORAGE_KEY);
}
