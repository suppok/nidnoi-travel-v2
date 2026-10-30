import { useCallback, useEffect, useState } from 'react';
import { countriesApi } from '../../../api/taxonomy.ts';
import type { Country } from '../../../types/taxonomy.ts';

type Result = { key: string; data: Country | null; error: unknown };

// Loads one country by id; reload() refetches after an edit or status change.
export function useCountry(id: string) {
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result | null>(null);
  const key = `${id}#${version}`;

  useEffect(() => {
    let cancelled = false;
    countriesApi
      .get(id)
      .then((data) => !cancelled && setResult({ key, data, error: null }))
      .catch((error: unknown) => !cancelled && setResult({ key, data: null, error }));
    return () => {
      cancelled = true;
    };
  }, [id, key]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);

  return {
    // Keeps showing the loaded country while it reloads.
    country: result?.data ?? null,
    loading: result?.key !== key,
    error: result?.key === key ? result.error : null,
    reload,
  };
}
