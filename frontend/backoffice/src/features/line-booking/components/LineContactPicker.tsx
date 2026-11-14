import { UserOutlined } from '@ant-design/icons';
import { App, Avatar, Select, Spin, Tag } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { lineContactsApi } from '../../../api/lineContacts.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { LineBookingContact, LineContact } from '../../../types/lineBooking.ts';

type Props = {
  value?: LineBookingContact | null; // shown as selected; null/undefined = nothing picked
  onPick: (contact: LineBookingContact) => void;
  placeholder?: string;
  disabled?: boolean;
  loading?: boolean;
  defaultOpen?: boolean;
  onClose?: () => void; // the dropdown closed without (or after) a pick
};

const SEARCH_DEBOUNCE_MS = 300;
const PAGE_SIZE = 20;

export function ContactAvatar({ pictureUrl, size = 'default' }: { pictureUrl: string | null; size?: 'small' | 'default' }) {
  return <Avatar size={size} src={pictureUrl ?? undefined} icon={<UserOutlined />} />;
}

// Picks a LINE OA chat. An empty search lists the newest chats; a typed name must match the whole
// display name (case-insensitive) so a partial name can't surface another customer's chat.
export default function LineContactPicker({ value, onPick, placeholder, disabled, loading, defaultOpen, onClose }: Props) {
  const { message } = App.useApp();
  const [keyword, setKeyword] = useState('');
  const [options, setOptions] = useState<LineContact[]>([]);
  const [searching, setSearching] = useState(false);
  const [open, setOpen] = useState(defaultOpen ?? false);
  const searchSeq = useRef(0);

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

  const changeOpen = (next: boolean) => {
    setOpen(next);
    if (!next) {
      setKeyword('');
      onClose?.();
    }
  };

  // Keep the selected contact renderable even when it isn't in the current results.
  const items: (LineBookingContact & Partial<LineContact>)[] = value && !options.some((c) => c.id === value.id) ? [value, ...options] : options;

  return (
    <Select<string>
      className="w-full"
      placeholder={placeholder ?? 'พิมพ์ชื่อ LINE ของลูกค้าให้ตรงทั้งชื่อเพื่อค้นหา'}
      value={value?.id ?? null}
      open={open}
      autoFocus={defaultOpen}
      onOpenChange={changeOpen}
      showSearch={{ filterOption: false, onSearch: setKeyword, searchValue: keyword }}
      loading={searching || loading}
      disabled={disabled}
      notFoundContent={
        searching ? <Spin size="small" /> : keyword.trim() ? 'ไม่พบแชทชื่อนี้ — พิมพ์ชื่อ LINE ให้ตรงทั้งชื่อ' : 'ไม่พบแชท — ลูกค้าต้องทักแชท OA ก่อน'
      }
      onChange={(id) => {
        const picked = items.find((c) => c.id === id);
        changeOpen(false);
        if (picked) onPick(picked);
      }}
      options={items.map((c) => ({
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
  );
}
