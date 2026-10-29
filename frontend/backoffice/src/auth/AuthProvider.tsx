import { App } from 'antd';
import { type ReactNode, useCallback, useEffect, useMemo, useState } from 'react';
import { authApi } from '../api/auth.ts';
import type { Permission } from '../constants/permissions.ts';
import type { LoginRequest, Me } from '../types/auth.ts';
import { authBus, tokenStore } from './token.ts';
import { AuthContext, type AuthContextValue, type AuthStatus } from './useAuth.ts';

export default function AuthProvider({ children }: { children: ReactNode }) {
  const { message } = App.useApp();
  const [status, setStatus] = useState<AuthStatus>('loading');
  const [user, setUser] = useState<Me | null>(null);

  const reloadUser = useCallback(async () => {
    const me = await authApi.me();
    setUser(me);
    setStatus('authenticated');
    return me;
  }, []);

  // Restore the session from the refresh cookie on page load.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      const token = await authApi.refresh();
      if (cancelled) return;
      if (!token) {
        setStatus('anonymous');
        return;
      }
      try {
        const me = await authApi.me();
        if (cancelled) return;
        setUser(me);
        setStatus('authenticated');
      } catch {
        if (!cancelled) setStatus('anonymous');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  // The API client gave up on the session (refresh failed): drop to /login.
  const hadUser = user !== null;
  useEffect(
    () =>
      authBus.onUnauthenticated(() => {
        if (hadUser) message.warning('เซสชันหมดอายุ กรุณาเข้าสู่ระบบอีกครั้ง');
        setUser(null);
        setStatus('anonymous');
      }),
    [message, hadUser],
  );

  const login = useCallback(
    async (body: LoginRequest) => {
      const { accessToken } = await authApi.login(body);
      tokenStore.set(accessToken);
      return reloadUser();
    },
    [reloadUser],
  );

  const logout = useCallback(async () => {
    await authApi.logout().catch(() => undefined);
    tokenStore.clear();
    setUser(null);
    setStatus('anonymous');
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      status,
      user,
      login,
      logout,
      reloadUser,
      can: (permission: Permission) => user?.permissions.includes(permission) ?? false,
    }),
    [status, user, login, logout, reloadUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
