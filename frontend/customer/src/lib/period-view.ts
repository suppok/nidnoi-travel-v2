// Markup for the period table, shared by the build-time render (PeriodTable.astro) and the browser
// refresh (its <script>), so the static and live HTML are identical (no CLS, no drift).
import { contact, lineMessageUrl } from '../config/site';
import { discountPercent, formatDateRange, formatPrice } from './format';
import type { TourPeriodPublic } from '../types/tour-program';

const LOW_SEATS = 5;

const escapeHtml = (s: string) =>
  s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c] as string);

/** First bookable period, else the first one; null when there are none. */
export function firstOpenId(periods: TourPeriodPublic[]): string | null {
  return (periods.find((p) => !p.soldOut) ?? periods[0])?.id ?? null;
}

function seatLabel(p: TourPeriodPublic): { text: string; tone: 'danger' | 'muted' | 'ok' } {
  if (p.soldOut || p.seatsAvailable === 0) return { text: 'เต็มแล้ว', tone: 'muted' };
  if (p.seatsAvailable <= LOW_SEATS) return { text: `เหลือ ${p.seatsAvailable} ที่นั่ง`, tone: 'danger' };
  return { text: `ว่าง ${p.seatsAvailable} ที่นั่ง`, tone: 'ok' };
}

export function legendText(count: number): string {
  return `เลือกรอบเดินทาง (${count} รอบ)`;
}

export function rowsHtml(periods: TourPeriodPublic[], selectedId: string | null): string {
  return periods
    .map((p) => {
      const seat = seatLabel(p);
      const pct = discountPercent(p.price, p.priceBeforeDiscount);
      return `<li><button type="button" class="period-row" data-id="${escapeHtml(p.id)}" aria-pressed="${p.id === selectedId}"${p.soldOut ? ' disabled' : ''}>`
        + `<span class="period-row-main"><span class="period-date">${formatDateRange(p.startDate, p.endDate)}</span>`
        + `<span class="period-seat ${seat.tone}">${seat.text}</span></span>`
        + `<span class="period-row-price"><span class="period-amount">${formatPrice(p.price)} <small>บาท</small></span>`
        + (pct > 0 ? `<span class="period-pct">ลด ${pct}%</span>` : '')
        + `</span></button></li>`;
    })
    .join('');
}

export function priceHtml(p: TourPeriodPublic | null): string {
  if (!p) return '';
  const pct = discountPercent(p.price, p.priceBeforeDiscount);
  return `<div class="period-price-head"><span>ราคารอบที่เลือก / ท่าน</span>`
    + (pct > 0 ? `<span class="chip chip-accent">ลด ${pct}%</span>` : '')
    + `</div><div class="period-price-value"><span class="amount">${formatPrice(p.price)}</span><span class="unit">บาท</span></div>`
    + (pct > 0 && p.priceBeforeDiscount ? `<div class="period-price-before">${formatPrice(p.priceBeforeDiscount)} บาท</div>` : '');
}

export function detailsHtml(p: TourPeriodPublic | null): string {
  if (!p) return '';
  const rows: [string, number | null][] = [
    ['เด็ก (มีเตียง)', p.childWithBedPrice],
    ['เด็ก (ไม่มีเตียง)', p.childNoBedPrice],
    ['พักเดี่ยว เพิ่ม', p.singleSupplement],
    ['มัดจำ / ท่าน', p.deposit],
  ];
  const shown = rows.filter((r): r is [string, number] => r[1] !== null);
  if (shown.length === 0) return '';
  return `<dl class="period-details">${shown
    .map(([label, value]) => `<div><dt>${label}</dt><dd>${formatPrice(value)} บาท</dd></div>`)
    .join('')}</dl>`;
}

/** Booking form for the chosen departure; the tour's booking page when none is bookable. */
export function bookingHref(code: string, p: TourPeriodPublic | null): string {
  const base = `/tour/${encodeURIComponent(code)}/booking/`;
  return p && !p.soldOut ? `${base}?period=${encodeURIComponent(p.id)}` : base;
}

/** LINE chat prefilled with the tour code and the chosen departure. */
export function lineHref(code: string, tourName: string, p: TourPeriodPublic | null): string {
  const when = p ? ` รอบ ${formatDateRange(p.startDate, p.endDate)}` : '';
  return lineMessageUrl(`สนใจทัวร์ ${code.toUpperCase()} ${tourName}${when}`);
}

export const primaryPhone = contact.phones[0];
