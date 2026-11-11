import dayjs from 'dayjs';
import { formatPhone } from '../../../lib/format.ts';

// Pure logic of the booking summary tool: totals and the text sales paste into LINE OA / Facebook.
// The LIFF app parses this text back (frontend/liff/src/lib/parseSummary.ts): keep the
// "• <label> : <value>" lines and their labels stable, or update the parser too.

export type PaymentMethod = 'transfer' | 'credit';

export type PriceLine = { count: number; price: number };
// A count of null follows the number of travellers (taxes are usually per person).
export type PerTravellerLine = { count: number | null; price: number };
export type OptionLine = { name: string; qty: number; price: number };

export type SummaryInput = {
  source: { name: string; shortName: string } | null;
  refName: string; // the customer's name on the source (e.g. their LINE name)
  refBookingNo: string;

  tour: { code: string; name: string; days: number; nights: number; pdfFile: string | null; airlineName: string | null; wholesaleName: string } | null;
  period: { startDate: string; endDate: string } | null; // YYYY-MM-DD

  adult: PriceLine;
  adultPriceBeforeDiscount: number | null; // shown as "(จาก …)" when higher than the price
  child: PriceLine;
  infant: PriceLine;
  visa: PriceLine;
  foreignPassport: PriceLine;
  singleRoom: PriceLine; // count = rooms
  joinLand: PriceLine;
  fuelTax: PerTravellerLine;
  airportTax: PerTravellerLine;
  options: OptionLine[];
  discountPerPerson: number; // off the adult, child and infant prices

  paymentMethod: PaymentMethod;
  creditFeePercent: number; // card fee, charged on what is paid now
  isDeposit: boolean;
  depositPerPerson: number;
  depositDueDate: string | null; // YYYY-MM-DD
  depositDueTime: string | null; // HH:mm
  remainingDueDate: string | null;
  fullDueDate: string | null;
  fullDueTime: string | null;

  customerName: string;
  customerPhone: string;
  customerAddress: string;
  sales: { name: string; phone: string } | null;
};

export type SummaryTotals = {
  travellers: number;
  subtotal: number; // everything before the card fee
  fee: number; // card fee, 0 for transfer
  deposit: number; // deposit before the card fee
  depositToPay: number; // deposit including the card fee
  remaining: number; // balance after the deposit (no card fee)
  grandTotal: number; // what the customer pays in total
};

const n = (v: number | null | undefined) => (Number.isFinite(v) ? Number(v) : 0);
const lineTotal = (l: PriceLine) => n(l.count) * n(l.price);

export function resolvedCount(line: PerTravellerLine, travellers: number): number {
  return line.count === null ? travellers : n(line.count);
}

// Card fee on an amount, rounded up to the next baht. Computed in basis points so float error never
// adds a baht (15,000 x 2.6% = 390); the backend's LINE booking quote does the same.
function cardFee(amount: number, percent: number): number {
  return Math.ceil((amount * Math.round(percent * 100)) / 10000);
}

export function computeTotals(input: SummaryInput): SummaryTotals {
  const travellers = n(input.adult.count) + n(input.child.count) + n(input.infant.count);
  const discount = n(input.discountPerPerson);

  const subtotal =
    n(input.adult.count) * (n(input.adult.price) - discount) +
    n(input.child.count) * (n(input.child.price) - discount) +
    n(input.infant.count) * (n(input.infant.price) - discount) +
    lineTotal(input.visa) +
    lineTotal(input.foreignPassport) +
    lineTotal(input.singleRoom) +
    lineTotal(input.joinLand) +
    resolvedCount(input.fuelTax, travellers) * n(input.fuelTax.price) +
    resolvedCount(input.airportTax, travellers) * n(input.airportTax.price) +
    input.options.reduce((sum, o) => sum + n(o.qty) * n(o.price), 0);

  const credit = input.paymentMethod === 'credit';
  const percent = n(input.creditFeePercent);
  const deposit = input.isDeposit ? n(input.depositPerPerson) * travellers : 0;

  // With a deposit the card fee applies to the deposit only; the balance is paid later without it.
  const fee = credit ? cardFee(input.isDeposit ? deposit : subtotal, percent) : 0;
  const depositToPay = input.isDeposit ? deposit + fee : 0;
  const remaining = input.isDeposit ? subtotal - deposit : 0;

  return { travellers, subtotal, fee, deposit, depositToPay, remaining, grandTotal: subtotal + fee };
}

// Default balance due date: 30 days before departure.
export function defaultRemainingDueDate(startDate: string): string {
  return dayjs(startDate).subtract(30, 'day').format('YYYY-MM-DD');
}

// A deposit is the default when departure is more than 30 days away.
export function defaultIsDeposit(startDate: string, today = dayjs()): boolean {
  return dayjs(startDate).diff(today.startOf('day'), 'day') > 30;
}

const longDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'long', year: 'numeric' });

export function thaiDate(date: string | null): string {
  return date ? longDate.format(dayjs(date).toDate()) : '-';
}

const money = (v: number) => `${v.toLocaleString('th-TH')} บาท`;
const orDash = (v: string | null | undefined) => (v && v.trim() ? v.trim() : '-');

export function refNameLabel(source: SummaryInput['source']): string {
  return source?.shortName ? `ชื่อ ${source.shortName}` : 'ชื่อ';
}

export function salesLabel(sales: SummaryInput['sales']): string {
  if (!sales) return '-';
  return sales.phone ? `${sales.name} ${formatPhone(sales.phone)}` : sales.name;
}

function deadline(date: string | null, time: string | null): string {
  return `ภายในวันที่ ${thaiDate(date)}${time ? ` (ก่อน ${time} น.)` : ''}`;
}

const SEAT_CUT = '(ขอสงวนสิทธิ์ตัดที่นั่งอัตโนมัติ หากไม่ชำระเงินภายในเวลาที่กำหนด)';

export function buildSummaryText(input: SummaryInput, totals: SummaryTotals = computeTotals(input)): string {
  const { tour, period } = input;
  const lines: string[] = [];
  const push = (...l: string[]) => lines.push(...l);

  // A priced line for extras, shown only when used.
  const extra = (label: string, line: PriceLine, unit = 'ท่าน') => {
    if (n(line.count) > 0 && n(line.price) > 0) push(`• ${label} : ${money(n(line.price))} (จำนวน ${n(line.count)} ${unit})`);
  };
  const perTraveller = (label: string, line: PerTravellerLine) => {
    const count = resolvedCount(line, totals.travellers);
    if (n(line.price) > 0 && count > 0) push(`• ${label} : ${money(n(line.price))} (จำนวน ${count} ท่าน)`);
  };

  push('🔴สรุปรายละเอียดการจอง🔴');
  push(`• จองผ่าน : ${orDash(input.source?.name)}`);
  push(`• ${refNameLabel(input.source)} : ${orDash(input.refName)}`);
  push(`• เลข booking : ${orDash(input.refBookingNo)}`);
  push(`• ชื่อโปรแกรม : ${tour ? `${tour.name} (${tour.days} วัน ${tour.nights} คืน)` : '-'}`);
  push(`• รหัสทัวร์ : ${orDash(tour?.code)}`);
  push(`• โปรแกรมทัวร์ : ${orDash(tour?.pdfFile)}`);
  push(`• สายการบิน : ${orDash(tour?.airlineName)}`);
  push(`• Wholesale : ${orDash(tour?.wholesaleName)}`);
  push(`• วันเดินทาง : ${period ? `${thaiDate(period.startDate)} - ${thaiDate(period.endDate)}` : '-'}`);
  push(`• จำนวนผู้เดินทาง : ${totals.travellers > 0 ? `${totals.travellers} ท่าน` : '-'}`);

  const adultPrice = n(input.adult.price);
  const before = n(input.adultPriceBeforeDiscount);
  push(`• ราคาทัวร์ผู้ใหญ่/ท่าน : ${adultPrice > 0 ? money(adultPrice) : '-'}${before > adultPrice && adultPrice > 0 ? ` (จาก ${money(before)})` : ''}`);
  if (n(input.child.count) > 0) push(`• ราคาเด็กอายุต่ำกว่า 20 ปี/ท่าน : ${n(input.child.price) > 0 ? `${money(n(input.child.price))} (จำนวน ${n(input.child.count)} ท่าน)` : '-'}`);
  if (n(input.infant.count) > 0) push(`• ราคาเด็กทารก/ท่าน : ${n(input.infant.price) > 0 ? `${money(n(input.infant.price))} (จำนวน ${n(input.infant.count)} ท่าน)` : '-'}`);
  extra('ค่า visa/ท่าน', input.visa);
  extra('ค่า passport ต่างชาติ/ท่าน', input.foreignPassport);
  extra('ค่าห้องพักเดี่ยว/ห้อง', input.singleRoom, 'ห้อง');
  extra('ค่า join land/ท่าน', input.joinLand);
  perTraveller('ค่าภาษีน้ำมัน/ท่าน', input.fuelTax);
  perTraveller('ค่าภาษีสนามบิน/ท่าน', input.airportTax);
  for (const o of input.options) {
    if (n(o.qty) > 0 && n(o.price) > 0) push(`• ${o.name.trim() || 'Option'}/ท่าน : ${money(n(o.price))} (จำนวน ${n(o.qty)} ท่าน)`);
  }
  if (n(input.discountPerPerson) > 0) push(`• ส่วนลด/ท่าน : ${money(n(input.discountPerPerson))}`);
  if (input.paymentMethod === 'credit') push(`• ค่าธรรมเนียมบัตรเครดิต (${n(input.creditFeePercent)}%) : ${money(totals.fee)}`);
  push(`• รวมเป็นเงิน : ${totals.grandTotal > 0 ? money(totals.grandTotal) : '-'}`);

  push('', '🔴รายละเอียดการชำระเงิน🔴');
  if (input.isDeposit) {
    const perPerson = n(input.depositPerPerson) > 0 ? ` (ท่านละ ${money(n(input.depositPerPerson))})` : '';
    push(`• กรุณาชำระเงินมัดจำยอดเงิน ${totals.depositToPay > 0 ? money(totals.depositToPay) : '- บาท'}${perPerson} ${deadline(input.depositDueDate, input.depositDueTime)} ${SEAT_CUT}`);
    push(`• กรุณาชำระเงินส่วนเหลือยอดเงิน ${totals.deposit > 0 ? money(totals.remaining) : '- บาท'} ภายในวันที่ ${thaiDate(input.remainingDueDate)}`);
  } else {
    push(`• กรุณาชำระเงินเต็มจำนวน ${totals.grandTotal > 0 ? money(totals.grandTotal) : '- บาท'} ${deadline(input.fullDueDate, input.fullDueTime)} ${SEAT_CUT}`);
  }

  push(
    '',
    '✅ ราคาที่ท่านซื้อ ถือเป็นราคาที่ต่ำ ที่ท่านพึงพอใจและตัดสินใจดีแล้ว เมื่อมีที่หลุดเพิ่มเติมในกรณีต่าง ๆ อาจเป็นราคาที่ต่ำกว่าที่ท่านซื้อก็เป็นไปได้',
    '✅ ไม่สามารถคืนเงิน เปลี่ยนแปลง หรือยกเลิกได้ในทุกกรณี',
    '-------------------------------',
    '',
    '🔴ข้อมูลลูกค้า🔴',
    `• ผู้จอง : ${input.customerName.trim() ? `คุณ${input.customerName.trim()}` : '-'}`,
    `• เบอร์โทร : ${orDash(input.customerPhone)}`,
    `• ที่อยู่ : ${orDash(input.customerAddress)}`,
    `• ผู้ดูแล : ${salesLabel(input.sales)}`,
  );

  return lines.join('\n');
}
