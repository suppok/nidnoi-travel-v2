import { UserOutlined } from '@ant-design/icons';
import { Alert, App, Avatar, Button, Select, Spin, Tag, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { lineContactsApi } from '../../../api/lineContacts.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { LineBooking, LineContact } from '../../../types/lineBooking.ts';

type Props = {
  booking: LineBooking;
  editable: boolean;
  onLinked: (updated: LineBooking) => void;
};

const SEARCH_DEBOUNCE_MS = 300;
const PAGE_SIZE = 20;

function ContactAvatar({ pictureUrl, size = 'default' }: { pictureUrl: string | null; size?: 'small' | 'default' }) {
  return <Avatar size={size} src={pictureUrl ?? undefined} icon={<UserOutlined />} />;
}

// The customer's LINE OA chat for a booking. The receipt is pushed to this chat later, so staff pick
// the chat the customer used; contacts appear here once they have messaged or followed the OA.
export default function LineContactSection({ booking, editable, onLinked }: Props) {
  const { message } = App.useApp();
  const [keyword, setKeyword] = useState('');
  const [options, setOptions] = useState<LineContact[]>([]);
  const [searching, setSearching] = useState(false);
  const [saving, setSaving] = useState(false);
  const [open, setOpen] = useState(false);
  const searchSeq = useRef(0);
  const contact = booking.lineContact;

  // Search only while the dropdown is open, debounced; newest chat first.
  useEffect(() => {
    if (!open) return;
    const seq = ++searchSeq.current;
    const timer = setTimeout(() => {
      setSearching(true);
      lineContactsApi
        .list({ keyword: keyword.trim() || undefined, page: 1, pageSize: PAGE_SIZE })
        .then((res) => seq === searchSeq.current && setOptions(res.items))
        .catch((err: unknown) => seq === searchSeq.current && message.error(errorMessage(err)))
        .finally(() => seq === searchSeq.current && setSearching(false));
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [keyword, open, message]);

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
          <Select<string>
            className="w-full"
            placeholder={contact ? 'เปลี่ยนแชท — พิมพ์ชื่อ LINE เพื่อค้นหา' : 'พิมพ์ชื่อ LINE ของลูกค้าเพื่อค้นหา (ไม่บังคับ)'}
            value={null}
            open={open}
            onOpenChange={setOpen}
            showSearch={{ filterOption: false, onSearch: setKeyword, searchValue: keyword }}
            loading={searching || saving}
            disabled={saving}
            notFoundContent={searching ? <Spin size="small" /> : 'ไม่พบแชท — ลูกค้าต้องทักแชท OA ก่อน'}
            onChange={(id) => {
              setOpen(false);
              setKeyword('');
              if (id !== contact?.id) void link(id);
            }}
            options={options.map((c) => ({
              value: c.id,
              label: (
                <div className="flex items-center gap-2">
                  <ContactAvatar pictureUrl={c.pictureUrl} size="small" />
                  <span className="truncate">{c.displayName || '(ไม่มีชื่อ)'}</span>
                  {!c.following && <Tag color="red" className="m-0">บล็อก</Tag>}
                  {c.lastMessageAt && <span className="ml-auto text-xs text-ink-muted">{formatDateTime(c.lastMessageAt)}</span>}
                </div>
              ),
            }))}
          />
        )}
      </div>
    </section>
  );
}
