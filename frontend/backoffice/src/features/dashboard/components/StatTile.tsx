import { ArrowDownOutlined, ArrowUpOutlined } from '@ant-design/icons';
import { Card, Tooltip } from 'antd';
import type { ReactNode } from 'react';
import { CHART } from '../chartTheme.ts';

type Props = {
  title: string;
  value: ReactNode;
  icon?: ReactNode;
  /** Compares against the previous period of the same length. */
  delta?: { current: number; previous: number };
  footer?: ReactNode;
};

function Delta({ current, previous }: { current: number; previous: number }) {
  if (previous === 0) {
    return <span className="text-ink-muted">{current > 0 ? 'ช่วงก่อนหน้าไม่มีข้อมูล' : 'เท่ากับช่วงก่อนหน้า'}</span>;
  }
  const change = ((current - previous) / previous) * 100;
  const rounded = Math.round(change);
  if (rounded === 0) return <span className="text-ink-muted">ใกล้เคียงช่วงก่อนหน้า</span>;
  const up = change > 0;
  return (
    <span className="inline-flex items-center gap-1 font-medium" style={{ color: up ? CHART.up : CHART.down }}>
      {up ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
      {up ? 'เพิ่มขึ้น' : 'ลดลง'} {Math.abs(rounded)}%
      <span className="font-normal text-ink-muted">จากช่วงก่อนหน้า</span>
    </span>
  );
}

// A headline number. The delta is text + arrow (never color alone).
export default function StatTile({ title, value, icon, delta, footer }: Props) {
  return (
    <Card className="h-full" styles={{ body: { padding: 20, height: '100%' } }}>
      <div className="flex h-full flex-col gap-2">
        <div className="flex items-center justify-between gap-2 text-[14px] text-ink-muted">
          <span>{title}</span>
          {icon && <span className="grid h-8 w-8 place-items-center rounded-full bg-sky text-[15px] text-ocean">{icon}</span>}
        </div>
        <div className="font-display text-[28px] leading-tight text-ink">{value}</div>
        <div className="mt-auto text-[12.5px]">
          {delta ? (
            <Tooltip title={`ช่วงก่อนหน้า: ${delta.previous.toLocaleString('th-TH')}`}>
              <span>
                <Delta {...delta} />
              </span>
            </Tooltip>
          ) : (
            footer && <span className="text-ink-muted">{footer}</span>
          )}
        </div>
      </div>
    </Card>
  );
}
