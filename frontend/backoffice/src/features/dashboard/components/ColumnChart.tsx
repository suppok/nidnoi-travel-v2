import { useEffect, useRef, useState } from 'react';
import type { DashboardUnit } from '../../../types/dashboard.ts';
import { axisFormatter, CHART, niceTicks, SERIES } from '../chartTheme.ts';
import { bucketTick, bucketTitle } from '../lib/dates.ts';

export type ColumnPoint = { date: string; value: number };

type Props = {
  points: ColumnPoint[];
  unit: DashboardUnit;
  /** Names the measure in the tooltip and the screen-reader table ("ยอดขาย"). */
  label: string;
  format: (value: number) => string;
  color?: string;
  height?: number;
};

const PAD = { top: 12, right: 8, bottom: 28, left: 48 };
const MAX_COLUMN = 24;
const RADIUS = 4;

// Column path with a 4px rounded data-end and a square base.
function columnPath(x: number, y: number, w: number, h: number): string {
  const r = Math.min(RADIUS, w / 2, h);
  return `M${x},${y + h} V${y + r} Q${x},${y} ${x + r},${y} H${x + w - r} Q${x + w},${y} ${x + w},${y + r} V${y + h} Z`;
}

function useWidth<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  const [width, setWidth] = useState(0);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const ro = new ResizeObserver(([entry]) => setWidth(entry.contentRect.width));
    ro.observe(el);
    return () => ro.disconnect();
  }, []);
  return [ref, width] as const;
}

// One measure over time buckets: thin columns from a single baseline, recessive grid, hover tooltip.
export default function ColumnChart({ points, unit, label, format, color = SERIES[0], height = 240 }: Props) {
  const [ref, width] = useWidth<HTMLDivElement>();
  const [hover, setHover] = useState<number | null>(null);

  const max = Math.max(0, ...points.map((p) => p.value));
  const ticks = niceTicks(max);
  const top = ticks[ticks.length - 1];
  const tickText = axisFormatter(top);

  const plotW = Math.max(0, width - PAD.left - PAD.right);
  const plotH = height - PAD.top - PAD.bottom;
  const band = points.length ? plotW / points.length : 0;
  const colW = Math.max(2, Math.min(MAX_COLUMN, band * 0.62));
  const y = (v: number) => PAD.top + plotH - (top ? (v / top) * plotH : 0);

  // About six evenly spaced date labels, always including the last bucket.
  const every = Math.max(1, Math.ceil(points.length / Math.max(2, Math.floor(plotW / 72))));
  const showTick = (i: number) => (points.length - 1 - i) % every === 0;

  const hovered = hover !== null ? points[hover] : null;
  const tipLeft = hover !== null ? PAD.left + band * hover + band / 2 : 0;

  return (
    <div ref={ref} className="relative w-full" style={{ height }} onMouseLeave={() => setHover(null)}>
      {width > 0 && (
        <svg width={width} height={height} role="img" aria-label={`${label} ตามช่วงเวลา`}>
          {ticks.map((t) => (
            <g key={t}>
              <line x1={PAD.left} x2={width - PAD.right} y1={y(t)} y2={y(t)} stroke={t === 0 ? CHART.baseline : CHART.grid} strokeWidth={1} />
              <text x={PAD.left - 8} y={y(t)} dy="0.32em" textAnchor="end" fontSize={11} fill={CHART.axisText} style={{ fontVariantNumeric: 'tabular-nums' }}>
                {tickText(t)}
              </text>
            </g>
          ))}
          {points.map((p, i) => {
            const cx = PAD.left + band * i + band / 2;
            const h = Math.max(0, PAD.top + plotH - y(p.value));
            return (
              <g key={p.date}>
                {hover === i && <rect x={PAD.left + band * i} y={PAD.top} width={band} height={plotH} fill={CHART.hover} />}
                {p.value > 0 && (
                  <path d={columnPath(cx - colW / 2, y(p.value), colW, Math.max(h, 1))} fill={color} opacity={hover === null || hover === i ? 1 : 0.55} />
                )}
                {showTick(i) && (
                  <text
                    x={i === points.length - 1 ? Math.min(cx + band / 2, width - PAD.right) : cx}
                    y={height - 8}
                    textAnchor={i === points.length - 1 ? 'end' : 'middle'}
                    fontSize={11}
                    fill={CHART.axisText}
                  >
                    {bucketTick(p.date, unit)}
                  </text>
                )}
                {/* Hit target: the whole band, taller and wider than the column. */}
                <rect
                  x={PAD.left + band * i}
                  y={PAD.top}
                  width={band}
                  height={plotH}
                  fill="transparent"
                  onMouseEnter={() => setHover(i)}
                  onTouchStart={() => setHover(i)}
                />
              </g>
            );
          })}
        </svg>
      )}
      {hovered && (
        <div
          className="pointer-events-none absolute z-10 min-w-[140px] -translate-x-1/2 rounded-[10px] border border-line bg-white px-3 py-2 text-[13px] shadow-[0_6px_20px_rgba(29,43,79,0.12)]"
          style={{ left: Math.min(Math.max(tipLeft, 80), width - 80), top: 0 }}
        >
          <div className="text-ink-muted">{bucketTitle(hovered.date, unit)}</div>
          <div className="mt-0.5 flex items-center gap-2">
            <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: color }} />
            <span className="text-ink-soft">{label}</span>
            <span className="ml-auto font-semibold text-ink" style={{ fontVariantNumeric: 'tabular-nums' }}>
              {format(hovered.value)}
            </span>
          </div>
        </div>
      )}
      <table className="sr-only">
        <caption>{label}</caption>
        <tbody>
          {points.map((p) => (
            <tr key={p.date}>
              <th scope="row">{bucketTitle(p.date, unit)}</th>
              <td>{format(p.value)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
