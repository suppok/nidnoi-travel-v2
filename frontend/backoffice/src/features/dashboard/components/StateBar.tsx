import { Empty, Tooltip } from 'antd';
import { formatCount } from '../chartTheme.ts';

export type StateSegment = { key: string | number; label: string; count: number; color: string };

// Part-to-whole of one total: a single segmented bar (2px surface gaps) with a legend that carries
// the label, count and share, so identity never depends on color alone.
export default function StateBar({ segments, unit = 'รายการ' }: { segments: StateSegment[]; unit?: string }) {
  const total = segments.reduce((n, s) => n + s.count, 0);
  if (total === 0) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="ยังไม่มีข้อมูลในช่วงนี้" />;
  const shown = segments.filter((s) => s.count > 0);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex h-3 w-full gap-[2px] overflow-hidden rounded-full" role="img" aria-label="สัดส่วนตามสถานะ">
        {shown.map((s) => (
          <Tooltip key={s.key} title={`${s.label}: ${formatCount(s.count)} ${unit} (${Math.round((s.count / total) * 100)}%)`}>
            <div className="h-full min-w-[3px]" style={{ flexGrow: s.count, flexBasis: 0, background: s.color }} />
          </Tooltip>
        ))}
      </div>
      <ul className="m-0 flex list-none flex-col gap-2 p-0">
        {segments.map((s) => (
          <li key={s.key} className="flex items-center gap-2 text-[13px]">
            <span className="inline-block h-2.5 w-2.5 shrink-0 rounded-full" style={{ background: s.color }} />
            <span className="truncate text-ink-soft">{s.label}</span>
            <span className="ml-auto font-semibold text-ink" style={{ fontVariantNumeric: 'tabular-nums' }}>
              {formatCount(s.count)}
            </span>
            <span className="w-9 text-right text-ink-muted" style={{ fontVariantNumeric: 'tabular-nums' }}>
              {total ? Math.round((s.count / total) * 100) : 0}%
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
