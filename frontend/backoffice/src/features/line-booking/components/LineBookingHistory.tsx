import { SendOutlined } from '@ant-design/icons';
import { Alert, App, Button, Mentions, Skeleton, Timeline, Typography } from 'antd';
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

const escapeRegExp = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

// A comment's text with its @mentions highlighted.
function CommentText({ text, mentions }: { text: string; mentions: { name: string }[] }) {
  const names = mentions.map((m) => `@${m.name}`).sort((a, b) => b.length - a.length);
  if (names.length === 0) return <span className="whitespace-pre-wrap">{text}</span>;
  const parts = text.split(new RegExp(`(${names.map(escapeRegExp).join('|')})`, 'g'));
  return (
    <span className="whitespace-pre-wrap">
      {parts.map((p, i) =>
        names.includes(p) ? (
          <span key={i} className="rounded bg-brand-soft px-1 font-medium text-brand-dark">
            {p}
          </span>
        ) : (
          p
        ),
      )}
    </span>
  );
}

// The comment box: "@" opens the people who can see the booking.
function CommentBox({ bookingId, onPosted }: { bookingId: string; onPosted: (e: LineBookingEvent) => void }) {
  const { message } = App.useApp();
  const [text, setText] = useState('');
  const [users, setUsers] = useState<{ id: string; name: string }[]>([]);
  const [posting, setPosting] = useState(false);

  useEffect(() => {
    let cancelled = false;
    lineBookingsApi
      .mentionable(bookingId)
      .then((res) => !cancelled && setUsers(res.items))
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [bookingId]);

  const post = async () => {
    const body = text.trim();
    if (!body) return;
    // A user is mentioned when "@<their name>" is still in the text.
    const mentionUserIds = users.filter((u) => body.includes(`@${u.name}`)).map((u) => u.id);
    setPosting(true);
    try {
      onPosted(await lineBookingsApi.comment({ id: bookingId, text: body, mentionUserIds }));
      setText('');
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setPosting(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <Mentions
        value={text}
        onChange={setText}
        prefix="@"
        rows={2}
        maxLength={2000}
        placeholder="เขียนความคิดเห็น… พิมพ์ @ เพื่อแท็กผู้ใช้"
        options={users.map((u) => ({ value: u.name, label: u.name, key: u.id }))}
        disabled={posting}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) void post();
        }}
      />
      <div className="flex items-center justify-between">
        <span className="text-xs text-ink-muted">เจ้าของใบจองและฝ่ายการเงินจะได้รับแจ้งเตือน · Ctrl+Enter เพื่อส่ง</span>
        <Button type="primary" size="small" icon={<SendOutlined />} loading={posting} disabled={!text.trim()} onClick={() => void post()}>
          ส่ง
        </Button>
      </div>
    </div>
  );
}

// "ประวัติ" of a LINE booking: every action and comment, oldest first, with the fields an edit
// changed; a comment box at the bottom.
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
            content:
              e.action === LineBookingEvents.Comment ? (
                <div className="flex flex-col gap-1">
                  <span className="text-xs text-ink-muted">
                    <span className="font-medium text-ink">{e.user?.name ?? '-'}</span> · {formatDateTime(e.createDate)}
                  </span>
                  <div className="rounded-lg bg-page px-3 py-2 text-sm">
                    <CommentText text={e.note} mentions={e.mentions} />
                  </div>
                </div>
              ) : (
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
      {loaded && (
        <CommentBox
          bookingId={bookingId}
          onPosted={(e) => setLoaded((prev) => (prev ? { ...prev, events: [...(prev.events ?? []), e] } : prev))}
        />
      )}
    </section>
  );
}
