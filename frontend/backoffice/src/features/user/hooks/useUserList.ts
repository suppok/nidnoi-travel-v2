import { useCallback, useEffect, useState } from 'react';
import { usersApi } from '../../../api/users.ts';
import type { ListResponse } from '../../../types/api.ts';
import type { User, UserListRequest } from '../../../types/user.ts';

type Result = { key: string; data: ListResponse<User> | null; error: unknown };

export function useUserList(filter: UserListRequest) {
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result | null>(null);

  const { keyword, role, status, page, pageSize } = filter;
  const key = JSON.stringify([keyword, role, status, page, pageSize, version]);

  useEffect(() => {
    let cancelled = false;
    usersApi
      .list({ keyword, role, status, page, pageSize })
      .then((data) => !cancelled && setResult({ key, data, error: null }))
      .catch((error: unknown) => !cancelled && setResult((prev) => ({ key, data: prev?.data ?? null, error })));
    return () => {
      cancelled = true;
    };
  }, [key, keyword, role, status, page, pageSize]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);

  return {
    // Keeps showing the previous page while the next one loads.
    data: result?.data ?? null,
    loading: result?.key !== key,
    error: result?.key === key ? result.error : null,
    reload,
  };
}
