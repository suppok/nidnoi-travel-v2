import { createContext, useContext } from 'react';
import type { Permission } from '../constants/permissions.ts';
import type { LoginRequest, Me } from '../types/auth.ts';

export type AuthStatus = 'loading' | 'authenticated' | 'anonymous';

export type AuthContextValue = {
  status: AuthStatus;
  user: Me | null;
  login: (body: LoginRequest) => Promise<Me>;
  logout: () => Promise<void>;
  reloadUser: () => Promise<Me>;
  // UX only — the backend re-checks every request.
  can: (permission: Permission) => boolean;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>');
  return ctx;
}
