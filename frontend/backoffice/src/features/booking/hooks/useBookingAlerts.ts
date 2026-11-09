import { createContext, useContext } from 'react';

export type BookingAlertsValue = {
  /** Active bookings still in state "new"; null until the first poll (or without booking:list). */
  newCount: number | null;
  /** Re-polls now, e.g. after a booking changed state. */
  refresh: () => void;
};

export const BookingAlertsContext = createContext<BookingAlertsValue>({ newCount: null, refresh: () => {} });

export function useBookingAlerts(): BookingAlertsValue {
  return useContext(BookingAlertsContext);
}
