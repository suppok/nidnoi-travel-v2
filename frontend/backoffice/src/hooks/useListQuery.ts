import { useCallback, useEffect, useState } from 'react';
import type { ListResponse } from '../types/api.ts';

type Result<T> = { key: string; data: ListResponse<T> | null; error: unknown };

// Loads a list endpoint for `filter`, refetching when the filter (compared as JSON) or reload() changes.
export function useListQuery<T, F>(fetcher: (filter: F) => Promise<ListResponse<T>>, filter: F) {
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result<T> | null>(null);
  const filterKey = JSON.stringify(filter);
  const key = `${filterKey}#${version}`;

  useEffect(() => {
    let cancelled = false;
    fetcher(JSON.parse(filterKey) as F)
      .then((data) => !cancelled && setResult({ key, data, error: null }))
      .catch((error: unknown) => !cancelled && setResult((prev) => ({ key, data: prev?.data ?? null, error })));
    return () => {
      cancelled = true;
    };
  }, [fetcher, filterKey, key]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);

  return {
    // Keeps showing the previous page while the next one loads.
    data: result?.data ?? null,
    loading: result?.key !== key,
    error: result?.key === key ? result.error : null,
    reload,
  };
}
