import dayjs from 'dayjs';

const shortDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit' });

// "30 พ.ย. 69 – 2 ธ.ค. 69 (3 วัน)" from YYYY-MM-DD dates.
export function tripRange(start: string, end: string): string {
  const s = dayjs(start.slice(0, 10));
  const e = dayjs(end.slice(0, 10));
  return `${shortDate.format(s.toDate())} – ${shortDate.format(e.toDate())} (${e.diff(s, 'day') + 1} วัน)`;
}

export function shortThaiDate(date: string | null): string {
  return date ? shortDate.format(dayjs(date).toDate()) : '-';
}
