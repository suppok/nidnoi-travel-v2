import type { LineBooking } from '../../../types/lineBooking.ts';

export type Line = { key: string; label: string; count: number; price: number };

export function lineBookingLines(b: LineBooking): Line[] {
  const rows: (Line | null)[] = [
    { key: 'twin', label: 'ผู้ใหญ่ (พักห้องคู่)', count: b.adultTwinCount, price: b.adultTwinPrice },
    b.adultSinglePrice !== null ? { key: 'single', label: 'ผู้ใหญ่ (พักห้องเดี่ยว)', count: b.adultSingleCount, price: b.adultSinglePrice } : null,
    b.childWithBedPrice !== null ? { key: 'childBed', label: 'เด็ก (มีเตียง)', count: b.childWithBedCount, price: b.childWithBedPrice } : null,
    b.childNoBedPrice !== null ? { key: 'childNoBed', label: 'เด็ก (ไม่มีเตียง)', count: b.childNoBedCount, price: b.childNoBedPrice } : null,
  ];
  return rows.filter((r): r is Line => r !== null && r.count > 0);
}

export function travellerSummary(b: LineBooking): string {
  const adults = b.adultTwinCount + b.adultSingleCount;
  const children = b.childWithBedCount + b.childNoBedCount;
  return children > 0 ? `${adults} ผู้ใหญ่, ${children} เด็ก` : `${adults} ผู้ใหญ่`;
}

const thaiDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'Asia/Bangkok' });

export function tripDates(b: LineBooking): string {
  const d = (iso: string) => thaiDate.format(new Date(`${iso}T00:00:00+07:00`));
  return `${d(b.startDate)} – ${d(b.endDate)}`;
}
