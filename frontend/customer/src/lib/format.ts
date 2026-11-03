// Display formatting. Runs at build time and in islands, so keep it dependency-free.

const priceFormat = new Intl.NumberFormat('th-TH');

export function formatPrice(baht: number): string {
  return priceFormat.format(baht);
}

// th-TH uses the Buddhist calendar; '2-digit' gives the short year the design uses ("22 พ.ย. 69").
// timeZone is fixed so build machines (UTC) and visitors render the same day.
const shortDate = new Intl.DateTimeFormat('th-TH', {
  day: '2-digit',
  month: 'short',
  year: '2-digit',
  timeZone: 'Asia/Bangkok',
});

const dayMonth = new Intl.DateTimeFormat('th-TH', {
  day: '2-digit',
  month: 'short',
  timeZone: 'Asia/Bangkok',
});

/** Parses an ISO date (`YYYY-MM-DD`) as midnight in Bangkok. */
function parseDate(iso: string): Date {
  return new Date(`${iso}T00:00:00+07:00`);
}

/** "22 พ.ย. 69" */
export function formatDate(iso: string): string {
  return shortDate.format(parseDate(iso));
}

/** "22 – 27 พ.ย. 69" or "28 พ.ย. – 04 ธ.ค. 69", or with both years when they differ. */
export function formatDateRange(startIso: string, endIso: string): string {
  const start = parseDate(startIso);
  const end = parseDate(endIso);
  if (start.getUTCFullYear() !== end.getUTCFullYear()) {
    return `${shortDate.format(start)} – ${shortDate.format(end)}`;
  }
  if (startIso.slice(0, 7) === endIso.slice(0, 7)) {
    return `${startIso.slice(8, 10)} – ${shortDate.format(end)}`;
  }
  return `${dayMonth.format(start)} – ${shortDate.format(end)}`;
}

/** "6 วัน 4 คืน" */
export function formatDuration(days: number, nights: number): string {
  return nights > 0 ? `${days} วัน ${nights} คืน` : `${days} วัน`;
}

/** Whole-percent discount, or 0 when there is none. */
export function discountPercent(price: number, before: number | null): number {
  if (!before || before <= price) return 0;
  return Math.round(((before - price) / before) * 100);
}

/** Splits plain-text descriptions into paragraphs on blank lines. */
export function paragraphs(text: string): string[] {
  return text
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean);
}
