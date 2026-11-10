import type { LineBooking } from '../../../types/lineBooking.ts';

export type Line = { key: string; label: string; count: number; price: number };

export function lineBookingLines(b: LineBooking): Line[] {
  const rows: (Line | null)[] = [
    { key: 'twin', label: 'ผู้ใหญ่ (พักห้องคู่)', count: b.adultTwin, price: b.price },
    b.singleSupplement !== null ? { key: 'single', label: 'ผู้ใหญ่ (พักห้องเดี่ยว)', count: b.adultSingle, price: b.price + b.singleSupplement } : null,
    b.childWithBedPrice !== null ? { key: 'childBed', label: 'เด็ก (มีเตียง)', count: b.childWithBed, price: b.childWithBedPrice } : null,
    b.childNoBedPrice !== null ? { key: 'childNoBed', label: 'เด็ก (ไม่มีเตียง)', count: b.childNoBed, price: b.childNoBedPrice } : null,
  ];
  return rows.filter((r): r is Line => r !== null && r.count > 0);
}

export function travellerSummary(b: LineBooking): string {
  const adults = b.adultTwin + b.adultSingle;
  const children = b.childWithBed + b.childNoBed;
  return children > 0 ? `${adults} ผู้ใหญ่, ${children} เด็ก` : `${adults} ผู้ใหญ่`;
}

const thaiDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'Asia/Bangkok' });

export function tripDates(b: LineBooking): string {
  const d = (iso: string) => thaiDate.format(new Date(`${iso}T00:00:00+07:00`));
  return `${d(b.startDate)} – ${d(b.endDate)}`;
}
