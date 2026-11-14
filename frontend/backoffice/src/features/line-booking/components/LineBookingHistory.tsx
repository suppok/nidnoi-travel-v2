import { Alert, Skeleton, Timeline, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import {
  LINE_BOOKING_EVENT_COLOR,
  LINE_BOOKING_FIELD_LABELS,
  LINE_BOOKING_MONEY_FIELDS,
  LineBookingEvents,
} from '../../../constants/lineBookingEvents.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { LineBookingEvent } from '../../../types/lineBooking.ts';

type Props = {
  bookingId: string;
  // Changes whenever the booking changes, so the history reloads.
  version: string;
};

function title(e: LineBookingEvent): string {
  switch (e.action) {
    case LineBookingEvents.Created:
      return 'ส่งใบจองจาก LINE';
    case LineBookingEvents.Confirmed:
      return `ยืนยันการชำระเงิน · ส่งให้ ${e.target || 'ฝ่ายการเงิน'}`;
    case LineBookingEvents.Cancelled:
      return 'ยกเลิกใบจอง';
    case LineBookingEvents.DocumentsSent:
      return `ฝ่ายการเงินส่งเอกสาร ${e.target}`;
    case LineBookingEvents.DocumentsUpdated:
      return `ฝ่ายการเงินเปลี่ยนไฟล์เอกสาร ${e.target}`;
    case LineBookingEvents.DocumentsApproved:
      return e.target ? `ยืนยันเอกสาร ${e.note} · ส่งให้ลูกค้าทาง LINE (${e.target})` : `ยืนยันเอกสาร ${e.note} (ไม่ได้ส่ง LINE)`;
    case LineBookingEvents.DocumentsRejected:
      return `ส่งเอกสาร ${e.target} กลับให้ฝ่ายการเงินแก้ไข`;
    case LineBookingEvents.LineContactSet:
      return 'เปลี่ยนแชท LINE ของลูกค้า';
    case LineBookingEvents.StaffNoteUpdated:
      return 'แก้บันทึกของเจ้าหน้าที่';
    case LineBookingEvents.Edited:
      return 'แก้ไขใบจองจาก LINE';
    case LineBookingEvents.BalanceSubmitted:
      return 'บันทึกการชำระส่วนที่เหลือ (แนบสลิป)';
    case LineBookingEvents.BalanceConfirmed:
      return `ยืนยันยอดคงเหลือ · ชำระครบแล้ว · ส่งให้ ${e.target || 'ฝ่ายการเงิน'} ออก RE`;
    case LineBookingEvents.Deleted:
      return 'ลบใบจอง';
    case LineBookingEvents.Restored:
      return 'กู้คืนใบจอง';
    default:
      return e.action;
  }
}

function value(field: string, v: string): string {
  if (v === '') return '–';
  if (LINE_BOOKING_MONEY_FIELDS.has(field) && /^-?\d+$/.test(v)) return Number(v).toLocaleString('th-TH');
  if (field === 'paymentMethod') return v === 'credit' ? 'บัตรเครดิต' : 'โอนเงิน';
  return v;
}

// "ประวัติ" of a LINE booking: every action, oldest first, with the fields an edit changed.
export default function LineBookingHistory({ bookingId, version }: Props) {
  const [loaded, setLoaded] = useState<{ key: string; events: LineBookingEvent[] | null; error: unknown } | null>(null);
  const key = `${bookingId}|${version}`;

  useEffect(() => {
    let cancelled = false;
    lineBookingsApi
      .events(bookingId)
      .then((res) => !cancelled && setLoaded({ key, events: res.items, error: null }))
      .catch((err: unknown) => !cancelled && setLoaded({ key, events: null, error: err }));
    return () => {
      cancelled = true;
    };
  }, [bookingId, key]);

  // Keep showing the previous history while a newer version loads.
  const events = loaded?.events ?? null;

  return (
    <section>
      <Typography.Title level={5}>ประวัติ</Typography.Title>
      {loaded?.error != null && <Alert type="error" showIcon title={errorMessage(loaded.error)} />}
      {!loaded && <Skeleton active paragraph={{ rows: 3 }} />}
      {events && events.length === 0 && (
        <Typography.Text type="secondary">ยังไม่มีประวัติ (ใบจองก่อนเปิดใช้ประวัติจะเริ่มบันทึกจากการเปลี่ยนแปลงครั้งถัดไป)</Typography.Text>
      )}
      {events && events.length > 0 && (
        <Timeline
          items={events.map((e) => ({
            key: e.id,
            color: LINE_BOOKING_EVENT_COLOR[e.action] ?? 'gray',
            content: (
              <div className="flex flex-col gap-1">
                <span className="font-medium">{title(e)}</span>
                <span className="text-xs text-ink-muted">
                  {formatDateTime(e.createDate)} · {e.user?.name ?? '-'}
                </span>
                {e.note && e.action === LineBookingEvents.DocumentsRejected && (
                  <span className="whitespace-pre-wrap text-sm text-red-700">{e.note}</span>
                )}
                {e.changes.length > 0 && (
                  <ul className="m-0 flex list-none flex-col gap-0.5 p-0 text-sm">
                    {e.changes.map((c) => (
                      <li key={c.field}>
                        <span className="text-ink-muted">{LINE_BOOKING_FIELD_LABELS[c.field] ?? c.field}:</span>{' '}
                        <span className="text-ink-muted line-through">{value(c.field, c.from)}</span> →{' '}
                        <span className="font-medium">{value(c.field, c.to)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ),
          }))}
        />
      )}
    </section>
  );
}
