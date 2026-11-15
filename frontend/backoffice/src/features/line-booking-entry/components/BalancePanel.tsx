import { Alert, App, Button, Card, Form, Input, Statistic } from 'antd';
import { useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import { dealFromBooking, dealRequest, dealTotals } from '../lib/deal.ts';
import { shortThaiDate } from '../lib/format.ts';
import type { SavedResult } from './EntryForm.tsx';
import ReopenedSummary from './ReopenedSummary.tsx';
import SlipUploadField from './SlipUploadField.tsx';

type Props = {
  booking: LineBooking;
  onSaved: (result: SavedResult) => void;
  onCancel: () => void;
  onChangePurpose: () => void;
};

// The customer paid the rest of a deposit booking: the balance slips (and a note); the deal is sent
// back as it is. The sales owner then confirms it in the booking drawer and finance issues the RE.
export default function BalancePanel({ booking, onSaved, onCancel, onChangePurpose }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<{ slips: string[]; note: string }>();
  const [submitting, setSubmitting] = useState(false);
  const deal = dealFromBooking(booking);
  const remaining = dealTotals(deal).remaining;
  const owner = booking.salesUser?.name ?? 'เซลล์เจ้าของการจอง';

  const submit = async () => {
    let v: { slips: string[]; note: string };
    try {
      v = await form.validateFields();
    } catch {
      return;
    }
    setSubmitting(true);
    try {
      const res = await lineBookingsApi.updateDeal({
        bookingNo: booking.bookingNo,
        lineContactId: booking.lineContact?.id ?? null,
        bookingSourceId: booking.bookingSource?.id ?? null,
        refName: booking.refName,
        refBookingNo: booking.refBookingNo,
        ...dealRequest(deal, booking.adultPriceBeforeDiscount),
        customerName: booking.customerName,
        customerPhone: booking.customerPhone,
        customerAddress: booking.customerAddress,
        note: (v.note ?? '').trim(),
        balanceSlipImages: v.slips,
      });
      onSaved({
        bookingNo: res.bookingNo,
        totalPrice: res.totalPrice,
        message: `บันทึกยอดคงเหลือแล้ว รอ ${owner} ยืนยันในหน้าใบจอง LINE แล้วฝ่ายการเงินจะออกใบเสร็จ`,
      });
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mx-auto flex w-full max-w-[720px] flex-col gap-4">
      <ReopenedSummary booking={booking} onCancel={onCancel} onChangePurpose={onChangePurpose} />
      <Card size="small" className="border-green-200 bg-green-50">
        <Statistic title="ยอดคงเหลือที่ลูกค้าต้องชำระ" value={remaining} suffix="บาท" />
        {booking.remainingDueDate && <p className="mt-1 mb-0 text-[13px] text-ink-muted">กำหนดชำระภายใน {shortThaiDate(booking.remainingDueDate)}</p>}
      </Card>
      <Card title="สลิปชำระส่วนที่เหลือ" size="small">
        <Form form={form} layout="vertical" requiredMark={false} initialValues={{ slips: [], note: booking.note }}>
          {booking.balanceSubmittedAt && (
            <Alert type="warning" showIcon className="mb-3" title="มีสลิปส่วนที่เหลือรอเซลล์ยืนยันอยู่แล้ว อัปโหลดใหม่เพื่อแทนที่ทั้งชุด" />
          )}
          <Form.Item
            name="slips"
            extra={`${owner} จะตรวจสลิปและยืนยันในหน้าใบจอง LINE`}
            rules={[{ validator: (_, v: string[] | undefined) => (v?.length ? Promise.resolve() : Promise.reject(new Error('กรุณาแนบสลิปชำระส่วนที่เหลือ'))) }]}
          >
            <SlipUploadField />
          </Form.Item>
          <Form.Item label="หมายเหตุ" name="note" className="mb-0">
            <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} maxLength={1000} />
          </Form.Item>
        </Form>
      </Card>
      <Button type="primary" size="large" block loading={submitting} onClick={() => void submit()}>
        ส่งยอดคงเหลือให้เซลล์ยืนยัน
      </Button>
    </div>
  );
}
