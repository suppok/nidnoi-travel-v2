import type { DashboardUnit } from '../../../types/dashboard.ts';

// Dashboard dates are plain YYYY-MM-DD (Thailand dates); format them as UTC so the day never shifts.
const dayMonth = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', timeZone: 'UTC' });
const dayMonthYear = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'UTC' });
const monthYear = new Intl.DateTimeFormat('th-TH', { month: 'short', year: '2-digit', timeZone: 'UTC' });
const monthLong = new Intl.DateTimeFormat('th-TH', { month: 'long', year: 'numeric', timeZone: 'UTC' });
const weekday = new Intl.DateTimeFormat('th-TH', { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'UTC' });

export function parseDate(date: string): Date {
  return new Date(`${date}T00:00:00Z`);
}

export function addDays(date: string, days: number): string {
  const d = parseDate(date);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

export function daysBetween(from: string, to: string): number {
  return Math.round((parseDate(to).getTime() - parseDate(from).getTime()) / 86_400_000);
}

// Short label under a column.
export function bucketTick(date: string, unit: DashboardUnit): string {
  return unit === 'month' ? monthYear.format(parseDate(date)) : dayMonth.format(parseDate(date));
}

// Full label in the tooltip.
export function bucketTitle(date: string, unit: DashboardUnit): string {
  const d = parseDate(date);
  if (unit === 'month') return monthLong.format(d);
  if (unit === 'week') return `สัปดาห์ ${dayMonth.format(d)} – ${dayMonth.format(parseDate(addDays(date, 6)))}`;
  return weekday.format(d);
}

export function formatShortDate(date: string): string {
  return dayMonthYear.format(parseDate(date));
}

// "15–19 พ.ย. 69" style range for departures.
export function formatDateRange(start: string, end: string): string {
  return `${dayMonth.format(parseDate(start))} – ${dayMonthYear.format(parseDate(end))}`;
}
