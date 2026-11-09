import { DeleteOutlined, MailOutlined, PhoneOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Descriptions, Drawer, Form, Input, Segmented, Skeleton, Table, Tag, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { bookingsApi } from '../../../api/bookings.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { BOOKING_STATE_INFO, type BookingState, bookingStateInfo } from '../../../constants/bookingStates.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import { formatBaht, formatDateTime } from '../../../lib/format.ts';
import type { Booking } from '../../../types/booking.ts';
import { bookingLines, type Line, travellerSummary, tripDates } from '../lib/bookingFormat.ts';

type Props = {
  bookingId: string | null;
  onClose: () => void;
  /** Called after a save or delete/restore so the list and the badge refresh. */
  onChanged: () => void;
};

// Same label column in every details table; labels never wrap ("วันเดินทาง").
const DESCRIPTION_STYLES = { label: { width: 150, whiteSpace: 'nowrap' as const } };

type FormValues = { state: BookingState; staffNote: string };

/** Booking details plus the sales workflow form (state + staff note). */
export default function BookingDrawer({ bookingId, onClose, onChanged }: Props) {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const [form] = Form.useForm<FormValues>();
  // Keyed by id so a stale result never shows for another booking.
  const [loaded, setLoaded] = useState<{ id: string; booking: Booking | null; error: unknown } | null>(null);
  const current = loaded?.id === bookingId ? loaded : null;
  const booking = current?.booking ?? null;
  const loadError = current?.error ?? null;
  const setBooking = (b: Booking) => setLoaded({ id: b.id, booking: b, error: null });
  const [saving, setSaving] = useState(false);
  const editable = can(Permissions.BookingUpdate) && booking?.status === true;

  useEffect(() => {
    if (!bookingId) return;
    let cancelled = false;
    bookingsApi
      .get(bookingId)
      .then((b) => {
        if (cancelled) return;
        setLoaded({ id: bookingId, booking: b, error: null });
        form.setFieldsValue({ state: b.state, staffNote: b.staffNote });
      })
      .catch((err: unknown) => !cancelled && setLoaded({ id: bookingId, booking: null, error: err }));
    return () => {
      cancelled = true;
    };
  }, [bookingId, form]);

  const save = async (values: FormValues) => {
    if (!booking) return;
    setSaving(true);
    try {
      const updated = await bookingsApi.update({ id: booking.id, state: values.state, staffNote: values.staffNote ?? '' });
      setBooking(updated);
      message.success(`บันทึกใบจอง ${updated.bookingNo} แล้ว`);
      onChanged();
      onClose();
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const confirmStatus = () => {
    if (!booking) return;
    const remove = booking.status;
    modal.confirm({
      title: remove ? `ลบใบจอง ${booking.bookingNo}?` : `กู้คืนใบจอง ${booking.bookingNo}?`,
      content: remove ? 'ใช้กับใบจองสแปมหรือซ้ำ ใบจองจะถูกซ่อนจากรายการ กู้คืนได้ภายหลัง' : 'ใบจองจะกลับมาแสดงในรายการ',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await bookingsApi.updateStatus({ id: booking.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          setBooking({ ...booking, status: !remove });
          onChanged();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const state = booking ? bookingStateInfo(booking.state) : null;

  return (
    <Drawer
      open={bookingId !== null}
      onClose={onClose}
      size={640}
      title={
        booking ? (
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono">{booking.bookingNo}</span>
            {state && <Tag color={state.color} className="m-0">{state.label}</Tag>}
            {!booking.status && <Tag className="m-0">ลบแล้ว</Tag>}
          </div>
        ) : (
          'ใบจอง'
        )
      }
      extra={
        booking &&
        can(Permissions.BookingStatus) && (
          <Button danger={booking.status} icon={booking.status ? <DeleteOutlined /> : <UndoOutlined />} onClick={confirmStatus}>
            {booking.status ? 'ลบ' : 'กู้คืน'}
          </Button>
        )
      }
      footer={
        editable && (
          <div className="flex justify-end gap-2">
            <Button onClick={onClose}>ปิด</Button>
            <Button type="primary" loading={saving} onClick={() => form.submit()}>
              บันทึก
            </Button>
          </div>
        )
      }
      destroyOnHidden
    >
      {loadError !== null && <Alert type="error" showIcon title={errorMessage(loadError)} />}
      {!booking && loadError === null && <Skeleton active paragraph={{ rows: 10 }} />}
      {booking && (
        <div className="flex flex-col gap-6">
          <Alert
            type="info"
            showIcon
            title="ใบจองจากหน้าเว็บยังไม่ได้กันที่นั่งและยังไม่มีการชำระเงิน ติดต่อลูกค้าเพื่อยืนยันที่นั่งและราคา แล้วปรับที่นั่งในหน้ารอบเดินทางเอง"
          />

          <section>
            <Typography.Title level={5}>ลูกค้า</Typography.Title>
            <Descriptions column={1} size="small" bordered styles={DESCRIPTION_STYLES}>
              <Descriptions.Item label="ชื่อ">{[booking.firstName, booking.lastName].filter(Boolean).join(' ')}</Descriptions.Item>
              <Descriptions.Item label="มือถือ">
                <a href={`tel:${booking.phone.replace(/[^\d+]/g, '')}`}>
                  <PhoneOutlined /> {booking.phone}
                </a>
                <Typography.Text copyable={{ text: booking.phone }} className="ml-2" />
              </Descriptions.Item>
              <Descriptions.Item label="อีเมล">
                <a href={`mailto:${booking.email}?subject=${encodeURIComponent(`ใบจอง ${booking.bookingNo} ทัวร์ ${booking.tourCode}`)}`}>
                  <MailOutlined /> {booking.email}
                </a>
                <Typography.Text copyable={{ text: booking.email }} className="ml-2" />
              </Descriptions.Item>
              <Descriptions.Item label="ความต้องการพิเศษ">
                {booking.customerNote ? <span className="whitespace-pre-wrap">{booking.customerNote}</span> : <span className="text-ink-muted">—</span>}
              </Descriptions.Item>
              <Descriptions.Item label="ส่งเมื่อ">{formatDateTime(booking.createDate)}</Descriptions.Item>
            </Descriptions>
          </section>

          <section>
            <Typography.Title level={5}>ทัวร์และผู้เดินทาง</Typography.Title>
            <div className="flex flex-col gap-4">
              <Descriptions column={1} size="small" bordered styles={DESCRIPTION_STYLES}>
                <Descriptions.Item label="ทัวร์">
                  {can(Permissions.TourProgramGet) ? (
                    <Link to={`/tour-programs/${booking.tourProgramId}/edit#periods`} className="font-mono">
                      {booking.tourCode}
                    </Link>
                  ) : (
                    <span className="font-mono">{booking.tourCode}</span>
                  )}{' '}
                  {booking.tourName}
                </Descriptions.Item>
                <Descriptions.Item label="วันเดินทาง">{tripDates(booking)}</Descriptions.Item>
              </Descriptions>
              <Table<Line>
                rowKey="key"
                size="small"
                pagination={false}
                dataSource={bookingLines(booking)}
                columns={[
                  { title: 'ห้องพัก', dataIndex: 'label' },
                  { title: 'ราคา/ท่าน', dataIndex: 'price', align: 'right', render: (v: number) => formatBaht(v) },
                  { title: 'จำนวน', dataIndex: 'count', align: 'center' },
                  { title: 'รวม', key: 'sum', align: 'right', render: (_, l) => formatBaht(l.price * l.count) },
                ]}
                summary={() => (
                  <Table.Summary.Row>
                    <Table.Summary.Cell index={0} colSpan={2}>
                      <strong>รวม</strong>
                    </Table.Summary.Cell>
                    <Table.Summary.Cell index={2} align="center">
                      {travellerSummary(booking)}
                    </Table.Summary.Cell>
                    <Table.Summary.Cell index={3} align="right">
                      <strong>{formatBaht(booking.totalPrice)}</strong>
                    </Table.Summary.Cell>
                  </Table.Summary.Row>
                )}
              />
            </div>
            <p className="mt-2 mb-0 text-xs text-ink-muted">ราคา ณ เวลาที่ลูกค้าส่งใบจอง ตรวจสอบราคาปัจจุบันในหน้ารอบเดินทางก่อนยืนยันกับลูกค้า</p>
          </section>

          <section>
            <Typography.Title level={5}>การติดตาม</Typography.Title>
            <Form<FormValues> form={form} layout="vertical" onFinish={save} disabled={!editable}>
              <Form.Item name="state" label="สถานะ" rules={[{ required: true }]}>
                <Segmented<BookingState> options={BOOKING_STATE_INFO.map((s) => ({ value: s.value, label: s.label, title: s.hint }))} />
              </Form.Item>
              <Form.Item name="staffNote" label="บันทึกของเจ้าหน้าที่" rules={[{ max: 2000, message: 'ไม่เกิน 2,000 ตัวอักษร' }]}>
                <Input.TextArea rows={4} placeholder="เช่น โทรแล้ว ลูกค้าขอเลื่อนรอบ / โอนมัดจำแล้ว" showCount maxLength={2000} />
              </Form.Item>
            </Form>
          </section>
        </div>
      )}
    </Drawer>
  );
}
