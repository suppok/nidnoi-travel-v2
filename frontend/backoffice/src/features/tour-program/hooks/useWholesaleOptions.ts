import { useEffect, useState } from 'react';
import { wholesalesApi } from '../../../api/wholesales.ts';
import type { Wholesale } from '../../../types/wholesale.ts';

// Every wholesale (active and deleted) for Select options; a few hundred rows at most.
export function useWholesaleOptions() {
  const [wholesales, setWholesales] = useState<Wholesale[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    wholesalesApi
      .list({ page: 1, pageSize: 500 })
      .then((data) => !cancelled && setWholesales(data.items))
      .catch(() => !cancelled && setWholesales([]));
    return () => {
      cancelled = true;
    };
  }, []);

  return { wholesales: wholesales ?? [], loading: wholesales === null };
}
