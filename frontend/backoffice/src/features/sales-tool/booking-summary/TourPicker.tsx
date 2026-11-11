import { App, Select, Spin } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { tourProgramsApi } from '../../../api/tourPrograms.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { TourProgramListItem } from '../../../types/tourProgram.ts';

const SEARCH_DEBOUNCE_MS = 300;
const PAGE_SIZE = 20;

type Props = {
  value: TourProgramListItem | null;
  onChange: (tour: TourProgramListItem | null) => void;
};

// Searches active tours by name, code or wholesale tour code while the dropdown is open.
export default function TourPicker({ value, onChange }: Props) {
  const { message } = App.useApp();
  const [open, setOpen] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [options, setOptions] = useState<TourProgramListItem[]>([]);
  const [searching, setSearching] = useState(false);
  const searchSeq = useRef(0);

  useEffect(() => {
    if (!open) return;
    const seq = ++searchSeq.current;
    const timer = setTimeout(() => {
      setSearching(true);
      tourProgramsApi
        .list({ keyword: keyword.trim() || undefined, status: true, page: 1, pageSize: PAGE_SIZE })
        .then((res) => seq === searchSeq.current && setOptions(res.items))
        .catch((err: unknown) => seq === searchSeq.current && message.error(errorMessage(err)))
        .finally(() => seq === searchSeq.current && setSearching(false));
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [keyword, open, message]);

  // Keep the chosen tour listed so the Select can show its label.
  const shown = value && !options.some((t) => t.id === value.id) ? [value, ...options] : options;

  return (
    <Select<string>
      className="w-full"
      placeholder="พิมพ์ชื่อ รหัสทัวร์ หรือรหัสโฮลเซลล์"
      value={value?.id ?? null}
      open={open}
      onOpenChange={setOpen}
      showSearch={{ filterOption: false, onSearch: setKeyword, searchValue: keyword }}
      allowClear
      loading={searching}
      notFoundContent={searching ? <Spin size="small" /> : 'ไม่พบโปรแกรมทัวร์'}
      onChange={(id) => {
        setOpen(false);
        setKeyword('');
        onChange(shown.find((t) => t.id === id) ?? null);
      }}
      options={shown.map((t) => ({
        value: t.id,
        label: (
          <span className="flex min-w-0 items-baseline gap-2">
            <span className="shrink-0 font-mono text-[13px] text-ink-muted">{t.code}</span>
            <span className="truncate">{t.name}</span>
          </span>
        ),
      }))}
    />
  );
}
