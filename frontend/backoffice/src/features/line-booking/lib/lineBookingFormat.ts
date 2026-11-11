import type { LineBooking, LineBookingChargeKind, LineBookingDocumentKey } from '../../../types/lineBooking.ts';

// One priced line of a booking; a negative price is a discount.
export type Line = { key: string; label: string; count: number; price: number };

const CHARGE_LABELS: Record<LineBookingChargeKind, string> = {
  visa: 'ค่า visa',
  foreignPassport: 'ค่า passport ต่างชาติ',
  joinLand: 'ค่า join land',
  fuelTax: 'ค่าภาษีน้ำมัน',
  airportTax: 'ค่าภาษีสนามบิน',
  option: 'Option',
};

export function travellerCount(b: LineBooking): number {
  return b.adultTwinCount + b.adultSingleCount + b.childWithBedCount + b.childNoBedCount + b.infantCount;
}

export function lineBookingLines(b: LineBooking): Line[] {
  const rows: (Line | null)[] = [
    { key: 'twin', label: 'ผู้ใหญ่ (พักห้องคู่)', count: b.adultTwinCount, price: b.adultTwinPrice },
    b.adultSinglePrice !== null ? { key: 'single', label: 'ผู้ใหญ่ (พักห้องเดี่ยว)', count: b.adultSingleCount, price: b.adultSinglePrice } : null,
    b.childWithBedPrice !== null ? { key: 'childBed', label: 'เด็ก (มีเตียง)', count: b.childWithBedCount, price: b.childWithBedPrice } : null,
    b.childNoBedPrice !== null ? { key: 'childNoBed', label: 'เด็ก (ไม่มีเตียง)', count: b.childNoBedCount, price: b.childNoBedPrice } : null,
    b.infantPrice !== null ? { key: 'infant', label: 'เด็กทารก', count: b.infantCount, price: b.infantPrice } : null,
    ...b.charges.map((c, i) => ({
      key: `charge-${i}`,
      label: c.kind === 'option' ? c.name : CHARGE_LABELS[c.kind],
      count: c.count,
      price: c.price,
    })),
    b.discountPerPerson > 0 ? { key: 'discount', label: 'ส่วนลด', count: travellerCount(b), price: -b.discountPerPerson } : null,
    b.cardFee > 0 ? { key: 'fee', label: `ค่าธรรมเนียมบัตรเครดิต (${b.creditFeePercent}%)`, count: 1, price: b.cardFee } : null,
  ];
  return rows.filter((r): r is Line => r !== null && r.count > 0);
}

export function travellerSummary(b: LineBooking): string {
  const adults = b.adultTwinCount + b.adultSingleCount;
  const children = b.childWithBedCount + b.childNoBedCount;
  const parts = [`${adults} ผู้ใหญ่`];
  if (children > 0) parts.push(`${children} เด็ก`);
  if (b.infantCount > 0) parts.push(`${b.infantCount} ทารก`);
  return parts.join(', ');
}

const thaiDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'Asia/Bangkok' });
const d = (iso: string) => thaiDate.format(new Date(`${iso}T00:00:00+07:00`));

export function tripDates(b: LineBooking): string {
  return `${d(b.startDate)} – ${d(b.endDate)}`;
}

// "15 ม.ค. 70 ก่อน 18:00 น." or "-"
export function dueDate(date: string | null, time = ''): string {
  if (!date) return '-';
  return time ? `${d(date)} ก่อน ${time} น.` : d(date);
}

export type PaymentPlan = { deposit: number; depositToPay: number; remaining: number };

// The deposit split, as told to the customer: the card fee is on the deposit only.
export type RequiredDocument = { key: LineBookingDocumentKey; code: string; label: string };

// The finance documents a confirmed booking needs: deposit → AI + IV, full payment → RE.
export function requiredDocuments(b: Pick<LineBooking, 'isDeposit'>): RequiredDocument[] {
  return b.isDeposit
    ? [
        { key: 'aiFile', code: 'AI', label: 'ใบรับเงินมัดจำ' },
        { key: 'ivFile', code: 'IV', label: 'ใบแจ้งหนี้' },
      ]
    : [{ key: 'reFile', code: 'RE', label: 'ใบเสร็จรับเงิน' }];
}

export function paymentPlan(b: LineBooking): PaymentPlan | null {
  if (!b.isDeposit) return null;
  const deposit = b.depositPerPerson * travellerCount(b);
  return { deposit, depositToPay: deposit + b.cardFee, remaining: b.totalPrice - b.cardFee - deposit };
}
