import { useEffect, useState } from 'react';
import { citiesApi } from '../../../api/taxonomy.ts';
import type { City } from '../../../types/taxonomy.ts';

// Cities (active and deleted) of the given countries, fetched once per country and cached.
export function useCityOptions(countryIds: readonly string[]) {
  const [byCountry, setByCountry] = useState<Record<string, City[]>>({});
  const missing = countryIds.filter((id) => !(id in byCountry));
  const missingKey = missing.join(',');

  useEffect(() => {
    if (!missingKey) return;
    let cancelled = false;
    for (const countryId of missingKey.split(',')) {
      citiesApi
        .list({ countryId, page: 1, pageSize: 500 })
        .then((data) => !cancelled && setByCountry((prev) => ({ ...prev, [countryId]: data.items })))
        .catch(() => !cancelled && setByCountry((prev) => ({ ...prev, [countryId]: [] })));
    }
    return () => {
      cancelled = true;
    };
  }, [missingKey]);

  return {
    cities: countryIds.flatMap((id) => byCountry[id] ?? []),
    loading: missing.length > 0,
  };
}
