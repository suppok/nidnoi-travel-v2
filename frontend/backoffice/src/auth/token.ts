// Access token lives in memory only. The refresh token is an HttpOnly cookie owned by the backend,
// so a reload restores the session through POST /auth/refresh without JS ever seeing it.
let accessToken: string | null = null;

export const tokenStore = {
  get: () => accessToken,
  set: (token: string) => {
    accessToken = token;
  },
  clear: () => {
    accessToken = null;
  },
};

type Listener = () => void;
const unauthenticatedListeners = new Set<Listener>();

// Fired by the API client when the session can't be recovered (refresh failed).
export const authBus = {
  onUnauthenticated(listener: Listener): () => void {
    unauthenticatedListeners.add(listener);
    return () => {
      unauthenticatedListeners.delete(listener);
    };
  },
  emitUnauthenticated() {
    unauthenticatedListeners.forEach((l) => l());
  },
};
