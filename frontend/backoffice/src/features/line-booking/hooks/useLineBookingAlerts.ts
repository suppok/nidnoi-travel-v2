import { createContext, useContext } from 'react';

export type LineBookingAlertsValue = {
  /** Active LINE bookings still in state "new"; null until the first poll (or without line-booking:list). */
  newCount: number | null;
  /** Finance only: confirmed bookings sent to me that still wait for my documents; null otherwise. */
  waitingDocumentsCount: number | null;
  /** Re-polls now, e.g. after a booking changed state. */
  refresh: () => void;
};

export const LineBookingAlertsContext = createContext<LineBookingAlertsValue>({ newCount: null, waitingDocumentsCount: null, refresh: () => {} });

export function useLineBookingAlerts(): LineBookingAlertsValue {
  return useContext(LineBookingAlertsContext);
}
