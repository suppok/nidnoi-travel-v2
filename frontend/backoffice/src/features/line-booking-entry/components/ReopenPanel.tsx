import { Alert, Input, List, Tag, Typography } from 'antd';
import { useRef, useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { LineBookingStates } from '../../../constants/lineBookingStates.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { LineBooking, LineBookingContact } from '../../../types/lineBooking.ts';
import LineContactPicker from '../../line-booking/components/LineContactPicker.tsx';
import { tripRange } from '../lib/format.ts';

function stateTag(b: LineBooking) {
  if (b.state === LineBookingStates.PaidInFull) return <Tag color="green">ชำระครบแล้ว</Tag>;
  if (b.balanceSubmittedAt && !b.balanceConfirmedAt) return <Tag color="gold">รอยืนยันยอดคงเหลือ</Tag>;
  return <Tag color="blue">ชำระมัดจำแล้ว</Tag>;
}

// Reopens an earlier paid booking, to change the deal or record the balance the customer paid:
// by its booking no, or from the customer's chat when the booking was linked to it.
export default function ReopenPanel({ onOpen }: { onOpen: (booking: LineBooking) => void }) {
  const [opening, setOpening] = useState(false);
  const [error, setError] = useState('');
  const [contact, setContact] = useState<LineBookingContact | null>(null);
  const [chatBookings, setChatBookings] = useState<LineBooking[] | null>(null);
  const [chatLoading, setChatLoading] = useState(false);
  const [chatError, setChatError] = useState('');
  const chatSeq = useRef(0);

  const openByNo = async (raw: string) => {
    const no = raw.trim().toUpperCase();
    if (!no) return;
    setError('');
    setOpening(true);
    try {
      onOpen(await lineBookingsApi.getEditable(no));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setOpening(false);
    }
  };

  const pickContact = async (c: LineBookingContact) => {
    const seq = ++chatSeq.current;
    setContact(c);
    setChatBookings(null);
    setChatError('');
    setChatLoading(true);
    try {
      const res = await lineBookingsApi.listEditable(c.id);
      if (seq === chatSeq.current) setChatBookings(res.items);
    } catch (err) {
      if (seq === chatSeq.current) setChatError(errorMessage(err));
    } finally {
      if (seq === chatSeq.current) setChatLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <Typography.Text strong>เลขที่จอง</Typography.Text>
        <Input.Search
          placeholder="เช่น 26X8K9AQ"
          enterButton="เปิดใบจอง"
          loading={opening}
          maxLength={32}
          autoComplete="off"
          className="font-mono"
          onSearch={(v) => void openByNo(v)}
        />
        {error && <Alert type="error" showIcon title={error} />}
      </div>

      <div className="flex flex-col gap-1.5">
        <Typography.Text strong>หรือค้นหาจากแชทลูกค้า</Typography.Text>
        <LineContactPicker value={contact} onPick={(c) => void pickContact(c)} />
        <Typography.Text type="secondary" className="text-xs">
          แสดงเฉพาะใบจองที่ชำระเงินแล้วและผูกแชทนี้ไว้
        </Typography.Text>
        {chatError && <Alert type="error" showIcon title={chatError} />}
        {(chatLoading || chatBookings) && (
          <List<LineBooking>
            bordered
            size="small"
            loading={chatLoading}
            dataSource={chatBookings ?? []}
            locale={{ emptyText: 'ไม่พบใบจองที่ชำระเงินแล้วของแชทนี้ — ลองกรอกเลขที่จองแทน' }}
            renderItem={(b) => (
              <List.Item className="cursor-pointer hover:bg-sky-soft" onClick={() => onOpen(b)}>
                <div className="flex w-full items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="font-mono font-medium">{b.bookingNo}</div>
                    <div className="truncate text-xs">
                      {b.tourCode} {b.tourName}
                    </div>
                    <div className="text-xs text-ink-muted">
                      {tripRange(b.startDate, b.endDate)} · {b.customerName}
                    </div>
                  </div>
                  <div className="flex shrink-0 flex-col items-end gap-1">
                    <span className="text-xs">{formatBaht(b.totalPrice)}</span>
                    {stateTag(b)}
                  </div>
                </div>
              </List.Item>
            )}
          />
        )}
      </div>
    </div>
  );
}
