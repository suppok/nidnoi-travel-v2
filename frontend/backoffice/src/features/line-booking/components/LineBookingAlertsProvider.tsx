import { App, Button } from 'antd';
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { LineBookingStates } from '../../../constants/lineBookingStates.ts';
import { Permissions } from '../../../constants/permissions.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import { LineBookingAlertsContext } from '../hooks/useLineBookingAlerts.ts';

const POLL_MS = 60_000;
const NOTIFICATION_KEY = 'new-line-booking';

export default function LineBookingAlertsProvider({ children }: { children: ReactNode }) {
  const { can } = useAuth();
  const { notification } = App.useApp();
  const navigate = useNavigate();
  const enabled = can(Permissions.LineBookingList);
  const [newCount, setNewCount] = useState<number | null>(null);
  const seenRef = useRef<number | null>(null);

  const notify = useCallback(
    (newest: LineBooking, total: number) =>
      notification.info({
        key: NOTIFICATION_KEY,
        title: 'มีใบจอง LINE ใหม่',
        description: `${newest.bookingNo} · ${newest.lineDisplayName} · ${newest.tourCode}${total > 1 ? ` (รอตรวจสอบทั้งหมด ${total} รายการ)` : ''}`,
        duration: 0,
        placement: 'bottomRight',
        actions: (
          <Button
            type="primary"
            size="small"
            onClick={() => {
              notification.destroy(NOTIFICATION_KEY);
              navigate(`/line-bookings?state=${LineBookingStates.New}&open=${newest.id}`);
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
    lineBookingsApi
      .list({ state: LineBookingStates.New, status: true, page: 1, pageSize: 1 })
      .then((res) => {
        setNewCount(res.total);
        const newest = res.items[0];
        const at = newest ? Date.parse(newest.createDate) : 0;
        const seen = seenRef.current;
        seenRef.current = Math.max(seen ?? 0, at);
        if (newest && seen !== null && at > seen) notify(newest, res.total);
      })
      .catch(() => {});
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

  const value = useMemo(() => ({ newCount: enabled ? newCount : null, refresh: poll }), [enabled, newCount, poll]);
  return <LineBookingAlertsContext.Provider value={value}>{children}</LineBookingAlertsContext.Provider>;
}
