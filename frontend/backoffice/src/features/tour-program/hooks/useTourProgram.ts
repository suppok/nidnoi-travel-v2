import { useCallback, useEffect, useState } from 'react';
import { tourProgramsApi } from '../../../api/tourPrograms.ts';
import type { TourProgram } from '../../../types/tourProgram.ts';

type Result = { key: string; data: TourProgram | null; error: unknown };

// Loads one tour program by id; reload() refetches after a status change.
export function useTourProgram(id: string) {
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result | null>(null);
  const key = `${id}#${version}`;

  useEffect(() => {
    let cancelled = false;
    tourProgramsApi
      .get(id)
      .then((data) => !cancelled && setResult({ key, data, error: null }))
      .catch((error: unknown) => !cancelled && setResult({ key, data: null, error }));
    return () => {
      cancelled = true;
    };
  }, [id, key]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);
  // Replaces the loaded tour with a save response without refetching.
  const replace = useCallback((data: TourProgram) => setResult((prev) => (prev ? { ...prev, data } : prev)), []);

  return {
    tour: result?.data ?? null,
    loading: result?.key !== key,
    error: result?.key === key ? result.error : null,
    reload,
    replace,
  };
}
