import { useEffect, useState } from 'react';
import { airlinesApi } from '../../../api/airlines.ts';
import type { Airline } from '../../../types/airline.ts';

// Every airline (active and deleted) for Select options; a few hundred rows at most.
export function useAirlineOptions() {
  const [airlines, setAirlines] = useState<Airline[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    airlinesApi
      .list({ page: 1, pageSize: 500 })
      .then((data) => !cancelled && setAirlines(data.items))
      .catch(() => !cancelled && setAirlines([]));
    return () => {
      cancelled = true;
    };
  }, []);

  return { airlines: airlines ?? [], loading: airlines === null };
}
