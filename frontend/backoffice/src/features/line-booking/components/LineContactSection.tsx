import { Alert, App, Button, Tag, Typography } from 'antd';
import { useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import LineContactPicker, { ContactAvatar } from './LineContactPicker.tsx';

type Props = {
  booking: LineBooking;
  editable: boolean;
  onLinked: (updated: LineBooking) => void;
};

// The customer's LINE OA chat for a booking. The receipt is pushed to this chat later, so staff pick
// the chat the customer used; contacts appear here once they have messaged or followed the OA.
export default function LineContactSection({ booking, editable, onLinked }: Props) {
  const { message } = App.useApp();
  const [saving, setSaving] = useState(false);
  const contact = booking.lineContact;

  const link = async (lineContactId: string | null) => {
    setSaving(true);
    try {
      const updated = await lineBookingsApi.updateLineContact({ id: booking.id, lineContactId });
      onLinked(updated);
      message.success(lineContactId ? 'เชื่อมแชท LINE แล้ว' : 'ยกเลิกการเชื่อมแชท LINE แล้ว');
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <section>
      <Typography.Title level={5}>แชท LINE ของลูกค้า</Typography.Title>
      <div className="flex flex-col gap-3">
        {contact ? (
          <div className="flex items-center gap-3">
            <ContactAvatar pictureUrl={contact.pictureUrl} />
            <span className="font-medium">{contact.displayName || '(ไม่มีชื่อ)'}</span>
            {!contact.following && <Tag color="red" className="m-0">บล็อก OA แล้ว</Tag>}
            {editable && (
              <Button size="small" type="link" danger disabled={saving} onClick={() => void link(null)}>
                ยกเลิกการเชื่อม
              </Button>
            )}
          </div>
        ) : (
          <Typography.Text type="secondary">
            ไม่ได้เชื่อมแชท (ไม่บังคับ) — เชื่อมเฉพาะเมื่อลูกค้าคุยกับ OA เพื่อให้ระบบส่งใบเสร็จทาง LINE ได้ ถ้าคุยผ่าน LINE ส่วนตัวไม่ต้องเชื่อม
          </Typography.Text>
        )}
        {contact && !contact.following && (
          <Alert type="warning" showIcon title="ลูกค้าบล็อก OA หรือยังไม่ได้เพิ่มเพื่อน ระบบจะส่งข้อความทาง LINE ไม่ได้" />
        )}
        {editable && (
          <LineContactPicker
            placeholder={contact ? 'เปลี่ยนแชท — พิมพ์ชื่อ LINE ให้ตรงทั้งชื่อ' : 'พิมพ์ชื่อ LINE ของลูกค้าให้ตรงทั้งชื่อ (ไม่บังคับ)'}
            loading={saving}
            disabled={saving}
            onPick={(c) => {
              if (c.id !== contact?.id) void link(c.id);
            }}
          />
        )}
      </div>
    </section>
  );
}
