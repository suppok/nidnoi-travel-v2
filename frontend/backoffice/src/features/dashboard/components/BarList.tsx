import { Empty, Tooltip } from 'antd';
import type { ReactNode } from 'react';
import { Link } from 'react-router';
import { SERIES } from '../chartTheme.ts';

export type BarListRow = {
  key: string;
  label: ReactNode;
  value: number;
  /** Small line under the label (e.g. "12 ใบจอง · 31 คน"). */
  detail?: ReactNode;
  /** Tooltip on the bar. */
  hint?: ReactNode;
  to?: string;
};

type Props = {
  rows: BarListRow[];
  format: (value: number) => string;
  color?: string;
  empty?: string;
};

// Ranked horizontal bars: label + detail on the left, value at the bar's end. One series, so no legend.
export default function BarList({ rows, format, color = SERIES[0], empty = 'ยังไม่มีข้อมูลในช่วงนี้' }: Props) {
  if (rows.length === 0) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={empty} />;
  const max = Math.max(1, ...rows.map((r) => r.value));
  const total = rows.reduce((n, r) => n + r.value, 0);

  return (
    <ul className="m-0 flex list-none flex-col gap-3 p-0">
      {rows.map((r) => {
        const pct = (r.value / max) * 100;
        const share = total ? Math.round((r.value / total) * 100) : 0;
        const label = r.to ? (
          <Link to={r.to} className="text-ink hover:text-brand">
            {r.label}
          </Link>
        ) : (
          r.label
        );
        return (
          <li key={r.key} className="flex flex-col gap-1">
            <div className="flex items-baseline justify-between gap-3">
              <div className="min-w-0 flex-1 truncate text-[14px] text-ink">{label}</div>
              <div className="shrink-0 text-[14px] font-semibold text-ink" style={{ fontVariantNumeric: 'tabular-nums' }}>
                {format(r.value)}
              </div>
            </div>
            <Tooltip title={r.hint ?? `${share}% ของรายการที่แสดง`} placement="topLeft">
              <div className="h-2 w-full rounded-full bg-sky-soft">
                <div className="h-2 rounded-full" style={{ width: `${Math.max(pct, r.value > 0 ? 2 : 0)}%`, background: color }} />
              </div>
            </Tooltip>
            {r.detail && <div className="text-[12px] text-ink-muted">{r.detail}</div>}
          </li>
        );
      })}
    </ul>
  );
}
