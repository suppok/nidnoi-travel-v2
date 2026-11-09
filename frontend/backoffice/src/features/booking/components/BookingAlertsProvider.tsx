import { App, Button } from 'antd';
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { bookingsApi } from '../../../api/bookings.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { BookingStates } from '../../../constants/bookingStates.ts';
import { Permissions } from '../../../constants/permissions.ts';
import type { Booking } from '../../../types/booking.ts';
import { BookingAlertsContext } from '../hooks/useBookingAlerts.ts';

// Bookings arrive from the customer site; staff learn about them by polling (no push channel yet).
const POLL_MS = 60_000;
const NOTIFICATION_KEY = 'new-booking';

function customerName(b: { firstName: string; lastName: string }) {
  return [b.firstName, b.lastName].filter(Boolean).join(' ');
}

/**
 * Polls the newest "new" booking for staff with booking:list: drives the menu/header badge, prefixes
 * the tab title with the count and pops a notification when a booking arrives while the app is open.
 */
export default function BookingAlertsProvider({ children }: { children: ReactNode }) {
  const { can } = useAuth();
  const { notification } = App.useApp();
  const navigate = useNavigate();
  const enabled = can(Permissions.BookingList);
  const [newCount, setNewCount] = useState<number | null>(null);
  // Newest createDate (ms) already seen; null until the first poll sets the baseline.
  const seenRef = useRef<number | null>(null);

  const notify = useCallback(
    (newest: Booking, total: number) =>
      notification.info({
        key: NOTIFICATION_KEY,
        title: 'มีใบจองใหม่จากหน้าเว็บ',
        description: `${newest.bookingNo} · ${customerName(newest)} · ${newest.tourCode}${total > 1 ? ` (รอติดต่อทั้งหมด ${total} รายการ)` : ''}`,
        duration: 0,
        placement: 'bottomRight',
        actions: (
          <Button
            type="primary"
            size="small"
            onClick={() => {
              notification.destroy(NOTIFICATION_KEY);
              navigate(`/bookings?state=${BookingStates.New}&open=${newest.id}`);
            }}
          >
            ดูใบจอง
          </Button>
        ),
      }),
    [navigate, notification],
  );

  const poll = useCallback(() => {
    if (!enabled) return;
    bookingsApi
      .list({ state: BookingStates.New, status: true, page: 1, pageSize: 1 })
      .then((res) => {
        setNewCount(res.total);
        const newest = res.items[0];
        const at = newest ? Date.parse(newest.createDate) : 0;
        const seen = seenRef.current;
        seenRef.current = Math.max(seen ?? 0, at);
        // The first poll only sets the baseline: bookings already waiting show in the badge, not as a popup.
        if (newest && seen !== null && at > seen) notify(newest, res.total);
      })
      .catch(() => {
        // Keep the last count; the next poll tries again.
      });
  }, [enabled, notify]);

  useEffect(() => {
    if (!enabled) return;
    poll();
    const timer = setInterval(() => {
      if (document.visibilityState === 'visible') poll();
    }, POLL_MS);
    const onVisible = () => document.visibilityState === 'visible' && poll();
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisible);
    };
  }, [enabled, poll]);

  // "(3) nidnoi Back-Office": visible from another tab.
  useEffect(() => {
    const base = document.title.replace(/^\(\d+\) /, '');
    document.title = newCount ? `(${newCount}) ${base}` : base;
  }, [newCount]);

  const value = useMemo(() => ({ newCount: enabled ? newCount : null, refresh: poll }), [enabled, newCount, poll]);
  return <BookingAlertsContext.Provider value={value}>{children}</BookingAlertsContext.Provider>;
}
