import { DeleteOutlined, PhoneOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Descriptions, Drawer, Form, Image, Input, Segmented, Skeleton, Table, Tag, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { LINE_BOOKING_STATE_INFO, type LineBookingState, lineBookingStateInfo } from '../../../constants/lineBookingStates.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import { formatBaht, formatDateTime } from '../../../lib/format.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import LineContactSection from './LineContactSection.tsx';
import { lineBookingLines, type Line, travellerSummary, tripDates } from '../lib/lineBookingFormat.ts';

type Props = {
  bookingId: string | null;
  onClose: () => void;
  onChanged: () => void;
};

const DESCRIPTION_STYLES = { label: { width: 150, whiteSpace: 'nowrap' as const } };

type FormValues = { state: LineBookingState; staffNote: string };

export default function LineBookingDrawer({ bookingId, onClose, onChanged }: Props) {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const [form] = Form.useForm<FormValues>();
  const [loaded, setLoaded] = useState<{ id: string; booking: LineBooking | null; error: unknown } | null>(null);
  const current = loaded?.id === bookingId ? loaded : null;
  const booking = current?.booking ?? null;
  const loadError = current?.error ?? null;
  const setBooking = (b: LineBooking) => setLoaded({ id: b.id, booking: b, error: null });
  const [saving, setSaving] = useState(false);
  const editable = can(Permissions.LineBookingUpdate) && booking?.status === true;

  useEffect(() => {
    if (!bookingId) return;
    let cancelled = false;
    lineBookingsApi
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
      const updated = await lineBookingsApi.update({ id: booking.id, state: values.state, staffNote: values.staffNote ?? '' });
      setBooking(updated);
      message.success(`บันทึกใบจอง LINE ${updated.bookingNo} แล้ว`);
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
      title: remove ? `ลบใบจอง LINE ${booking.bookingNo}?` : `กู้คืนใบจอง LINE ${booking.bookingNo}?`,
      content: remove ? 'ใบจองจะถูกซ่อนจากรายการ กู้คืนได้ภายหลัง' : 'ใบจองจะกลับมาแสดงในรายการ',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await lineBookingsApi.updateStatus({ id: booking.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          setBooking({ ...booking, status: !remove });
          onChanged();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const state = booking ? lineBookingStateInfo(booking.state) : null;

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
          'ใบจอง LINE'
        )
      }
      extra={
        booking &&
        can(Permissions.LineBookingStatus) && (
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
          <section>
            <Typography.Title level={5}>ข้อมูลจาก LINE</Typography.Title>
            <Descriptions column={1} size="small" bordered styles={DESCRIPTION_STYLES}>
              <Descriptions.Item label="เซลล์เจ้าของ">{booking.salesUser?.name ?? <span className="text-ink-muted">-</span>}</Descriptions.Item>
              <Descriptions.Item label="ส่งโดย (LINE)">{booking.lineDisplayName}</Descriptions.Item>
              <Descriptions.Item label="ชื่อลูกค้า">{booking.customerName}</Descriptions.Item>
              <Descriptions.Item label="เบอร์โทร">
                <a href={`tel:${booking.customerPhone.replace(/[^\d+]/g, '')}`}>
                  <PhoneOutlined /> {booking.customerPhone}
                </a>
                <Typography.Text copyable={{ text: booking.customerPhone }} className="ml-2" />
              </Descriptions.Item>
              <Descriptions.Item label="หมายเหตุ">
                {booking.note ? <span className="whitespace-pre-wrap">{booking.note}</span> : <span className="text-ink-muted">-</span>}
              </Descriptions.Item>
              <Descriptions.Item label="ส่งเมื่อ">{formatDateTime(booking.createDate)}</Descriptions.Item>
            </Descriptions>
          </section>

          <LineContactSection
            booking={booking}
            editable={editable && can(Permissions.LineContactList)}
            onLinked={(updated) => {
              setBooking(updated);
              onChanged();
            }}
          />

          <section>
            <Typography.Title level={5}>สลิปโอนเงิน</Typography.Title>
            <Image src={booking.slipImage} alt="สลิปโอนเงิน" width={240} className="rounded-lg" />
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
                dataSource={lineBookingLines(booking)}
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
          </section>

          <section>
            <Typography.Title level={5}>การติดตาม</Typography.Title>
            <Form<FormValues> form={form} layout="vertical" onFinish={save} disabled={!editable}>
              <Form.Item name="state" label="สถานะ" rules={[{ required: true }]}>
                <Segmented<LineBookingState> options={LINE_BOOKING_STATE_INFO.map((s) => ({ value: s.value, label: s.label, title: s.hint }))} />
              </Form.Item>
              <Form.Item name="staffNote" label="บันทึกของเจ้าหน้าที่" rules={[{ max: 2000, message: 'ไม่เกิน 2,000 ตัวอักษร' }]}>
                <Input.TextArea rows={4} placeholder="เช่น ตรวจสอบสลิปแล้ว ยอดตรง / สลิปไม่ชัด ขอใหม่" showCount maxLength={2000} />
              </Form.Item>
            </Form>
          </section>
        </div>
      )}
    </Drawer>
  );
}
