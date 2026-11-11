import { App, Button } from 'antd';
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { FINANCE_LINE_BOOKINGS_PATH, LineBookingStates, lineBookingStateInfo } from '../../../constants/lineBookingStates.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { Roles } from '../../../constants/roles.ts';
import type { LineBooking, LineBookingListRequest } from '../../../types/lineBooking.ts';
import { requiredDocuments } from '../lib/lineBookingFormat.ts';
import { LineBookingAlertsContext } from '../hooks/useLineBookingAlerts.ts';

const POLL_MS = 60_000;

// One alert stream: polls the newest matching booking, pops a notification when a newer one shows up
// (by its timestamp field). The first poll only sets the baseline.
type Stream = {
  key: string;
  request: LineBookingListRequest;
  at: (b: LineBooking) => string | null;
  title: string;
  describe: (b: LineBooking, total: number) => string;
  path: (b: LineBooking) => string;
};

const STREAMS = {
  // Sales: a new booking from LIFF (owner scope applies).
  newBooking: {
    key: 'new-line-booking',
    request: { state: LineBookingStates.New, status: true, page: 1, pageSize: 1 },
    at: (b) => b.createDate,
    title: 'มีใบจอง LINE ใหม่',
    describe: (b, total) => `${b.bookingNo} · ${b.lineDisplayName} · ${b.tourCode}${total > 1 ? ` (รอตรวจสอบทั้งหมด ${total} รายการ)` : ''}`,
    path: (b) => `/line-bookings?state=${LineBookingStates.New}&open=${b.id}`,
  },
  // Finance: a sales owner confirmed a booking and sent it to me; waiting for my documents.
  waitingDocuments: {
    key: 'line-booking-documents-waiting',
    request: { paid: true, assignedToMe: true, documentsDone: false, status: true, page: 1, pageSize: 1 },
    // A fix request from sales brings the booking back, newer than its confirmation.
    at: (b) => latest(b.confirmedAt, b.fixRequestedAt),
    title: 'มีใบจอง LINE รอออกเอกสาร',
    describe: (b) =>
      b.fixRequestedAt && b.fixNote
        ? `${b.bookingNo} · ${b.salesUser?.name ?? 'เซลล์'} ขอแก้ไข: ${b.fixNote}`
        : `${b.bookingNo} · ${lineBookingStateInfo(b.state).label} · ${b.customerName} · ต้องออก ${requiredDocuments(b)
            .map((d) => d.code)
            .join(' + ')}`,
    path: (b) => `${FINANCE_LINE_BOOKINGS_PATH}?open=${b.id}`,
  },
  // Sales: finance sent the documents of my booking.
  documentsReady: {
    key: 'line-booking-documents-ready',
    request: { ownedByMe: true, documentsDone: true, documentsApproved: false, status: true, page: 1, pageSize: 1 },
    at: (b) => b.documentsAt,
    title: 'เอกสารการเงินพร้อมแล้ว',
    describe: (b) =>
      `${b.bookingNo} · ${b.customerName} · ${requiredDocuments(b)
        .map((d) => d.code)
        .join(' + ')} จาก ${b.financeUser?.name ?? 'ฝ่ายการเงิน'} — เปิดเพื่อตรวจและส่งให้ลูกค้า`,
    path: (b) => `/line-bookings?open=${b.id}`,
  },
} satisfies Record<string, Stream>;

type StreamName = keyof typeof STREAMS;

function latest(a: string | null, b: string | null): string | null {
  if (!a || !b) return a ?? b;
  return Date.parse(a) >= Date.parse(b) ? a : b;
}

export default function LineBookingAlertsProvider({ children }: { children: ReactNode }) {
  const { can, user } = useAuth();
  const { notification } = App.useApp();
  const navigate = useNavigate();
  const enabled = can(Permissions.LineBookingList);
  const roles = user?.roles ?? [];
  const isSales = enabled && roles.includes(Roles.Sales);
  const isFinance = enabled && roles.includes(Roles.Finance);
  const [newCount, setNewCount] = useState<number | null>(null);
  const [waitingCount, setWaitingCount] = useState<number | null>(null);
  const seen = useRef<Partial<Record<StreamName, number>>>({});

  const check = useCallback(
    (name: StreamName, onTotal?: (total: number) => void) => {
      const s: Stream = STREAMS[name];
      lineBookingsApi
        .list(s.request)
        .then((res) => {
          onTotal?.(res.total);
          const newest = res.items[0];
          const raw = newest ? s.at(newest) : null;
          const at = raw ? Date.parse(raw) : 0;
          const before = seen.current[name];
          seen.current[name] = Math.max(before ?? 0, at);
          if (!newest || before === undefined || at <= before) return;
          notification.info({
            key: s.key,
            title: s.title,
            description: s.describe(newest, res.total),
            duration: 0,
            placement: 'bottomRight',
            actions: (
              <Button
                type="primary"
                size="small"
                onClick={() => {
                  notification.destroy(s.key);
                  navigate(s.path(newest));
                }}
              >
                ดูใบจอง
              </Button>
            ),
          });
        })
        .catch(() => {});
    },
    [navigate, notification],
  );

  const poll = useCallback(() => {
    if (!enabled) return;
    check('newBooking', setNewCount);
    if (isFinance) check('waitingDocuments', setWaitingCount);
    if (isSales) check('documentsReady');
  }, [check, enabled, isFinance, isSales]);

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

  const value = useMemo(
    () => ({ newCount: enabled ? newCount : null, waitingDocumentsCount: isFinance ? waitingCount : null, refresh: poll }),
    [enabled, isFinance, newCount, waitingCount, poll],
  );
  return <LineBookingAlertsContext.Provider value={value}>{children}</LineBookingAlertsContext.Provider>;
}
