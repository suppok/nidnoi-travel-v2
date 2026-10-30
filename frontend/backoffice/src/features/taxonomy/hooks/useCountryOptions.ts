import { useEffect, useState } from 'react';
import { countriesApi } from '../../../api/taxonomy.ts';
import type { Country } from '../../../types/taxonomy.ts';

// Every country (active and deleted) for Select options; ~130 rows, so one request is enough.
export function useCountryOptions() {
  const [countries, setCountries] = useState<Country[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    countriesApi
      .list({ page: 1, pageSize: 500 })
      .then((data) => !cancelled && setCountries(data.items))
      .catch(() => !cancelled && setCountries([]));
    return () => {
      cancelled = true;
    };
  }, []);

  return { countries: countries ?? [], loading: countries === null };
}
