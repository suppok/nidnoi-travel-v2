import { Alert, App, Form, Modal, Select, Skeleton } from 'antd';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { LineBookingStates, lineBookingStateInfo } from '../../../constants/lineBookingStates.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { LineBooking, LineBookingFinanceUsers } from '../../../types/lineBooking.ts';

type Props = {
  booking: LineBooking | null; // open when set
  // The viewer owns the booking: the empty state links to their own finance settings.
  isOwner: boolean;
  onClose: () => void;
  onConfirmed: (updated: LineBooking) => void;
};

type FormValues = { financeUserId: string };

// "ยืนยันการชำระเงิน": pick the finance user who issues the receipt. The new state follows the
// booking's payment plan (deposit → ชำระมัดจำแล้ว, full → ชำระครบแล้ว).
export default function ConfirmLineBookingModal({ booking, isOwner, onClose, onConfirmed }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [loaded, setLoaded] = useState<{ id: string; data: LineBookingFinanceUsers | null; error: unknown } | null>(null);
  const [saving, setSaving] = useState(false);
  const current = booking && loaded?.id === booking.id ? loaded : null;
  const options = current?.data?.items ?? [];

  useEffect(() => {
    if (!booking) return;
    let cancelled = false;
    lineBookingsApi
      .financeUsers(booking.id)
      .then((data) => {
        if (cancelled) return;
        setLoaded({ id: booking.id, data, error: null });
        const preset = data.items.some((f) => f.id === data.defaultId) ? data.defaultId : data.items.length === 1 ? data.items[0].id : undefined;
        form.setFieldsValue({ financeUserId: preset });
      })
      .catch((err: unknown) => !cancelled && setLoaded({ id: booking.id, data: null, error: err }));
    return () => {
      cancelled = true;
    };
  }, [booking, form]);

  const submit = async ({ financeUserId }: FormValues) => {
    if (!booking) return;
    setSaving(true);
    try {
      const updated = await lineBookingsApi.confirm({ id: booking.id, financeUserId });
      message.success(`ยืนยันใบจอง ${updated.bookingNo} แล้ว — ส่งให้ ${updated.financeUser?.name ?? 'ฝ่ายการเงิน'} ออกใบเสร็จ`);
      onConfirmed(updated);
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const next = booking ? lineBookingStateInfo(booking.isDeposit ? LineBookingStates.DepositPaid : LineBookingStates.PaidInFull) : null;

  return (
    <Modal
      open={booking !== null}
      title={booking ? `ยืนยันการชำระเงิน ${booking.bookingNo}` : ''}
      okText="ยืนยัน"
      cancelText="ปิด"
      onOk={() => form.submit()}
      okButtonProps={{ loading: saving, disabled: options.length === 0 }}
      onCancel={onClose}
      destroyOnHidden
    >
      {!current && <Skeleton active paragraph={{ rows: 2 }} />}
      {current?.error != null && <Alert type="error" showIcon title={errorMessage(current.error)} />}
      {current?.data && options.length === 0 && (
        <Alert
          type="warning"
          showIcon
          title="ยังไม่มีฝ่ายการเงินให้เลือก"
          description={
            isOwner ? (
              <>
                เลือกฝ่ายการเงินของคุณก่อนที่ <Link to="/settings/finance">ฝ่ายการเงินของฉัน</Link>
              </>
            ) : (
              'เซลล์เจ้าของใบจองยังไม่ได้เลือกฝ่ายการเงิน'
            )
          }
        />
      )}
      {current?.data && options.length > 0 && (
        <Form<FormValues> form={form} layout="vertical" onFinish={submit} requiredMark={false}>
          <p className="mb-4 text-ink-muted">
            สถานะจะเปลี่ยนเป็น <strong className="text-ink">{next?.label}</strong> และแจ้งฝ่ายการเงินให้ออกใบเสร็จ
          </p>
          <Form.Item name="financeUserId" label="ฝ่ายการเงิน" rules={[{ required: true, message: 'กรุณาเลือกฝ่ายการเงิน' }]}>
            <Select placeholder="เลือกฝ่ายการเงิน" options={options.map((f) => ({ value: f.id, label: f.name }))} />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}
