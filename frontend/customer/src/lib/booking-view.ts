// Markup and pricing for the booking page, shared by the build-time render (BookingForm.astro) and
// its browser script, so the static and live HTML are identical. The backend re-quotes every price;
// these totals are for display only.
import type { BookingCounts } from '../types/booking';
import type { TourPeriodPublic } from '../types/tour-program';
import { formatDateRange, formatPrice } from './format';

export type RowKey = keyof BookingCounts;

export type PriceRow = { key: RowKey; label: string; room: string; price: number; adult: boolean };

export const MAX_TRAVELLERS = 50;
export const DEFAULT_COUNTS: BookingCounts = { adultTwin: 2, adultSingle: 0, childWithBed: 0, childNoBed: 0 };

const escapeHtml = (s: string) =>
  s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c] as string);

/** The price rows a period offers: adults twin always, the others only when the period prices them. */
export function priceRows(p: TourPeriodPublic): PriceRow[] {
  const rows: PriceRow[] = [{ key: 'adultTwin', label: 'ผู้ใหญ่', room: 'พักห้องคู่', price: p.price, adult: true }];
  if (p.singleSupplement !== null) {
    rows.push({ key: 'adultSingle', label: 'ผู้ใหญ่', room: 'พักห้องเดี่ยว', price: p.price + p.singleSupplement, adult: true });
  }
  if (p.childWithBedPrice !== null) {
    rows.push({ key: 'childWithBed', label: 'เด็ก', room: 'มีเตียง', price: p.childWithBedPrice, adult: false });
  }
  if (p.childNoBedPrice !== null) {
    rows.push({ key: 'childNoBed', label: 'เด็ก', room: 'ไม่มีเตียง', price: p.childNoBedPrice, adult: false });
  }
  return rows;
}

/** Drops counts for rows the period doesn't offer. */
export function fitCounts(rows: PriceRow[], counts: BookingCounts): BookingCounts {
  const offered = new Set(rows.map((r) => r.key));
  const next = { ...counts };
  for (const key of Object.keys(next) as RowKey[]) if (!offered.has(key)) next[key] = 0;
  return next;
}

export function travellers(counts: BookingCounts): { adults: number; children: number; total: number } {
  const adults = counts.adultTwin + counts.adultSingle;
  const children = counts.childWithBed + counts.childNoBed;
  return { adults, children, total: adults + children };
}

export function totalPrice(rows: PriceRow[], counts: BookingCounts): number {
  return rows.reduce((sum, r) => sum + r.price * counts[r.key], 0);
}

export function periodOptionLabel(p: TourPeriodPublic): string {
  const seats = p.soldOut || p.seatsAvailable === 0 ? 'เต็มแล้ว' : `ว่าง ${p.seatsAvailable} ที่นั่ง`;
  return `${formatDateRange(p.startDate, p.endDate)} · ${formatPrice(p.price)} บาท · ${seats}`;
}

/** First bookable period: the requested one when it is open, else the first open one. */
export function pickPeriod(periods: TourPeriodPublic[], wantedId: string | null): TourPeriodPublic | null {
  const open = periods.filter((p) => !p.soldOut && p.seatsAvailable > 0);
  return open.find((p) => p.id === wantedId) ?? open[0] ?? null;
}

/** Travellers table body: one row per price with a − / + stepper. */
export function stepperRowsHtml(rows: PriceRow[], counts: BookingCounts): string {
  const total = travellers(counts).total;
  return rows
    .map((r) => {
      const n = counts[r.key];
      const min = r.key === 'adultTwin' && counts.adultSingle === 0 ? 1 : 0;
      const id = `count-${r.key}`;
      return `<tr>`
        + `<th scope="row"><span class="bk-row-label">${r.label}</span> <span class="bk-row-room">(${r.room})</span>`
        + `<span class="bk-row-unit">${formatPrice(r.price)} บาท / ท่าน</span></th>`
        + `<td class="bk-col-price">${formatPrice(r.price)}</td>`
        + `<td><div class="bk-stepper">`
        + `<button type="button" data-step="-1" data-key="${r.key}" aria-label="ลด${r.label} ${r.room}"${n <= min ? ' disabled' : ''}>−</button>`
        + `<input id="${id}" type="number" inputmode="numeric" min="${min}" max="${MAX_TRAVELLERS}" value="${n}" data-key="${r.key}" aria-label="จำนวน${r.label} ${r.room}" />`
        + `<button type="button" data-step="1" data-key="${r.key}" aria-label="เพิ่ม${r.label} ${r.room}"${total >= MAX_TRAVELLERS ? ' disabled' : ''}>+</button>`
        + `</div></td>`
        + `<td class="bk-col-sum">${formatPrice(r.price * n)}</td>`
        + `</tr>`;
    })
    .join('');
}

export function travellersText(counts: BookingCounts): string {
  const t = travellers(counts);
  return `${t.adults} ผู้ใหญ่, ${t.children} เด็ก`;
}

/** "รายละเอียดราคา" list in the summary card. */
export function summaryHtml(rows: PriceRow[], counts: BookingCounts): string {
  const lines = rows
    .filter((r) => counts[r.key] > 0)
    .map(
      (r) =>
        `<li><span>${counts[r.key]} × <em>${r.label} (${r.room})</em><small>(${formatPrice(r.price)} บ. / คน)</small></span>`
        + `<span>${formatPrice(r.price * counts[r.key])}</span></li>`,
    )
    .join('');
  return `<ul class="bk-summary-lines">${lines}`
    + `<li class="bk-summary-fee"><span>ค่าธรรมเนียมการจอง</span><span>ฟรี</span></li></ul>`
    + `<div class="bk-summary-total"><span>รวม</span><span>${formatPrice(totalPrice(rows, counts))} บาท</span></div>`;
}

export function periodDatesHtml(p: TourPeriodPublic | null, duration: string): string {
  if (!p) return '<span class="text-ink-muted">ยังไม่ได้เลือกรอบเดินทาง</span>';
  return `${escapeHtml(formatDateRange(p.startDate, p.endDate))} (${escapeHtml(duration)})`;
}

/** Shown when more travellers are requested than seats left: still allowed, staff check it. */
export function seatsHint(p: TourPeriodPublic | null, counts: BookingCounts): string {
  if (!p || travellers(counts).total <= p.seatsAvailable) return '';
  return `รอบนี้เหลือ ${p.seatsAvailable} ที่นั่ง เจ้าหน้าที่จะตรวจสอบที่นั่งให้อีกครั้งเมื่อติดต่อกลับ`;
}
