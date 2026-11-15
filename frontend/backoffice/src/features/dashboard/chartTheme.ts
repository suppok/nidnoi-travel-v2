// Chart colors. Series slots are the dataviz reference palette (validated on the white card surface,
// all three pairs CVD-safe); text never wears a series color. Keep the order fixed.
export const SERIES = ['#2a78d6', '#eb6834', '#1baf7a'] as const;

// A neutral for "cancelled" / "not set" parts of a whole: reads as "nothing", never as a series.
export const NEUTRAL = '#b8c2d1';

export const CHART = {
  grid: '#e3eaf2',
  baseline: '#c9d5e2',
  axisText: '#5a6785',
  hover: 'rgba(29, 43, 79, 0.05)',
  up: '#006300',
  down: '#d03b3b',
} as const;

const compact = new Intl.NumberFormat('th-TH', { notation: 'compact', maximumFractionDigits: 1 });
const full = new Intl.NumberFormat('th-TH');

export function formatCount(n: number): string {
  return full.format(n);
}

// Axis ticks: in full under 10,000, compact above (one style per axis is decided by the max).
export function axisFormatter(max: number): (n: number) => string {
  return max >= 10_000 ? (n) => compact.format(n) : (n) => full.format(n);
}

// Clean ticks from 0 to at least max, about `count` of them.
export function niceTicks(max: number, count = 4): number[] {
  if (max <= 0) return [0, 1];
  const raw = max / count;
  const mag = 10 ** Math.floor(Math.log10(raw));
  const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= raw) ?? 10 * mag;
  const ticks: number[] = [];
  for (let v = 0; v < max + step * 0.999; v += step) ticks.push(Math.round(v * 1e6) / 1e6);
  return ticks;
}
