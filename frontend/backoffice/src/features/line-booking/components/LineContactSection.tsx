import { EditOutlined } from '@ant-design/icons';
import { Alert, App, Button, Input, Space, Tag, Typography } from 'antd';
import { useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { LineBooking, LineBookingContact } from '../../../types/lineBooking.ts';
import LineContactPicker, { ContactAvatar } from './LineContactPicker.tsx';
import OpenChatLink from './OpenChatLink.tsx';

type Props = {
  booking: LineBooking;
  editable: boolean;
  onLinked: (updated: LineBooking) => void;
};

const CHAT_URL = /^https:\/\/chat\.line\.biz\/\S+$/;

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
          <div className="flex flex-wrap items-center gap-3">
            <ContactAvatar pictureUrl={contact.pictureUrl} />
            <span className="font-medium">{contact.displayName || '(ไม่มีชื่อ)'}</span>
            {!contact.following && <Tag color="red" className="m-0">บล็อก OA แล้ว</Tag>}
            <OpenChatLink url={contact.chatUrl} />
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
        {contact && editable && <ChatUrlField booking={booking} contact={contact} onSaved={onLinked} />}
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

// The chat's link in LINE OA Manager, pasted by staff (OA Manager's chat ID can't be derived).
// Saved on the chat itself, so every booking with this chat can open it.
function ChatUrlField({ booking, contact, onSaved }: { booking: LineBooking; contact: LineBookingContact; onSaved: (updated: LineBooking) => void }) {
  const { message } = App.useApp();
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(contact.chatUrl);
  const [saving, setSaving] = useState(false);
  const url = value.trim();
  const invalid = url !== '' && !CHAT_URL.test(url);

  const save = async () => {
    setSaving(true);
    try {
      onSaved(await lineBookingsApi.updateLineContact({ id: booking.id, lineContactId: contact.id, chatUrl: url }));
      message.success('บันทึกลิงก์แชทแล้ว');
      setEditing(false);
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  if (!editing) {
    return (
      <Button
        size="small"
        type="link"
        icon={<EditOutlined />}
        className="self-start px-0"
        onClick={() => {
          setValue(contact.chatUrl);
          setEditing(true);
        }}
      >
        {contact.chatUrl ? 'แก้ไขลิงก์แชทใน LINE OA Manager' : 'เพิ่มลิงก์แชทใน LINE OA Manager'}
      </Button>
    );
  }

  return (
    <div className="flex flex-col gap-1">
      <Space.Compact className="w-full">
        <Input
          autoFocus
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onPressEnter={() => !invalid && url && void save()}
          status={invalid ? 'error' : undefined}
          placeholder="https://chat.line.biz/U…/chat/U…"
          autoComplete="off"
        />
        <Button type="primary" loading={saving} disabled={invalid || !url || url === contact.chatUrl} onClick={() => void save()}>
          บันทึก
        </Button>
        <Button disabled={saving} onClick={() => setEditing(false)}>
          ยกเลิก
        </Button>
      </Space.Compact>
      <Typography.Text type={invalid ? 'danger' : 'secondary'} className="text-xs">
        {invalid ? 'ลิงก์ต้องขึ้นต้นด้วย https://chat.line.biz/' : 'เปิดแชทของลูกค้าใน chat.line.biz แล้วคัดลอกลิงก์จากแถบที่อยู่มาวาง ใช้ได้กับทุกใบจองของแชทนี้'}
      </Typography.Text>
    </div>
  );
}
