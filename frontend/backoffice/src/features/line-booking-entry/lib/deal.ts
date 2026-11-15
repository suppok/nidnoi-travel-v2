import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { LineBookingStates } from '../../../constants/lineBookingStates.ts';
import type { LineBooking, LineBookingCharge, LineBookingDeal } from '../../../types/lineBooking.ts';
import type { TourProgramPeriod } from '../../../types/tourProgram.ts';
import {
  computeTotals,
  defaultIsDeposit,
  defaultRemainingDueDate,
  type PaymentMethod,
  type SummaryInput,
  type SummaryTotals,
} from '../../sales-tool/booking-summary/summary.ts';
import type { ParsedSummary, PricedLine } from './parseSummary.ts';

// The deal of a LINE booking as the entry form holds it: the same rows as the "สรุปยอดโอน" tool, so
// totals come from its computeTotals (summary.ts). The backend recomputes the total with quoteLine
// (backend/pkg/modules/linebooking/usecase) — keep both in sync. dealRequest maps the rows to the
// backend's room types: single rooms → adults in single rooms (supplement = the room price), the
// other adults → twin, children → "with bed".

export type Num = number | null;

export type DealOption = { name?: string; qty?: Num; price?: Num };

export type DealValues = {
  adultCount: Num;
  adultPrice: Num;
  childCount: Num; // under 20
  childPrice: Num;
  infantCount: Num;
  infantPrice: Num;
  visaCount: Num;
  visaPrice: Num;
  foreignCount: Num;
  foreignPrice: Num;
  singleRoomCount: Num; // rooms
  singleRoomPrice: Num;
  joinLandCount: Num;
  joinLandPrice: Num;
  fuelTaxCount: Num; // null = number of travellers
  fuelTaxPrice: Num;
  airportTaxCount: Num; // null = number of travellers
  airportTaxPrice: Num;
  options: DealOption[];
  discountPerPerson: Num; // off adult, child and infant prices
  paymentMethod: PaymentMethod;
  creditFeePercent: Num;
  isDeposit: boolean;
  depositPerPerson: Num;
  depositDueDate: Dayjs | null;
  depositDueTime: Dayjs | null;
  remainingDueDate: Dayjs | null;
  fullDueDate: Dayjs | null;
  fullDueTime: Dayjs | null;
};

export const DEFAULT_CREDIT_FEE_PERCENT = 2.6;

export const EMPTY_DEAL: DealValues = {
  adultCount: 1,
  adultPrice: null,
  childCount: null,
  childPrice: null,
  infantCount: null,
  infantPrice: null,
  visaCount: null,
  visaPrice: null,
  foreignCount: null,
  foreignPrice: null,
  singleRoomCount: null,
  singleRoomPrice: null,
  joinLandCount: null,
  joinLandPrice: null,
  fuelTaxCount: null,
  fuelTaxPrice: null,
  airportTaxCount: null,
  airportTaxPrice: null,
  options: [],
  discountPerPerson: null,
  paymentMethod: 'transfer',
  creditFeePercent: DEFAULT_CREDIT_FEE_PERCENT,
  isDeposit: false,
  depositPerPerson: null,
  depositDueDate: null,
  depositDueTime: null,
  remainingDueDate: null,
  fullDueDate: null,
  fullDueTime: null,
};

const n = (v: Num | undefined) => (typeof v === 'number' && Number.isFinite(v) ? v : 0);
const date = (d: Dayjs | null | undefined) => (d ? d.format('YYYY-MM-DD') : null);
const time = (t: Dayjs | null | undefined) => (t ? t.format('HH:mm') : '');
const day = (s: string | null | undefined) => (s ? dayjs(s) : null);
const clock = (s: string | null | undefined) => (s ? dayjs(`2000-01-01T${s}`) : null);

// Totals with the summary tool's math.
export function dealTotals(v: DealValues): SummaryTotals {
  const line = (count: Num, price: Num) => ({ count: n(count), price: n(price) });
  const input: SummaryInput = {
    source: null,
    refName: '',
    refBookingNo: '',
    tour: null,
    period: null,
    adult: line(v.adultCount, v.adultPrice),
    adultPriceBeforeDiscount: null,
    child: line(v.childCount, v.childPrice),
    infant: line(v.infantCount, v.infantPrice),
    visa: line(v.visaCount, v.visaPrice),
    foreignPassport: line(v.foreignCount, v.foreignPrice),
    singleRoom: line(v.singleRoomCount, v.singleRoomPrice),
    joinLand: line(v.joinLandCount, v.joinLandPrice),
    fuelTax: { count: v.fuelTaxCount ?? null, price: n(v.fuelTaxPrice) },
    airportTax: { count: v.airportTaxCount ?? null, price: n(v.airportTaxPrice) },
    options: (v.options ?? []).map((o) => ({ name: o?.name ?? '', qty: n(o?.qty), price: n(o?.price) })),
    discountPerPerson: n(v.discountPerPerson),
    paymentMethod: v.paymentMethod,
    creditFeePercent: n(v.creditFeePercent),
    isDeposit: v.isDeposit,
    depositPerPerson: n(v.depositPerPerson),
    depositDueDate: null,
    depositDueTime: null,
    remainingDueDate: null,
    fullDueDate: null,
    fullDueTime: null,
    customerName: '',
    customerPhone: '',
    customerAddress: '',
    sales: null,
  };
  return computeTotals(input);
}

// Picking a departure fills its prices and the payment defaults, like the summary tool. Everything
// stays editable.
export function periodPatch(p: TourProgramPeriod): Partial<DealValues> {
  const deposit = defaultIsDeposit(p.startDate) && n(p.deposit) > 0;
  return {
    adultPrice: p.price,
    childPrice: p.childWithBedPrice ?? p.price,
    singleRoomPrice: p.singleSupplement,
    isDeposit: deposit,
    ...(deposit ? { depositPerPerson: p.deposit } : {}),
    remainingDueDate: dayjs(defaultRemainingDueDate(p.startDate)),
  };
}

// Client-side checks before saving; the backend checks the same and more.
export function dealError(v: DealValues, totals: SummaryTotals): string | null {
  const adults = n(v.adultCount);
  if (adults < 1) return 'กรุณาระบุผู้ใหญ่อย่างน้อย 1 ท่าน';
  if (totals.travellers > 50) return 'ผู้เดินทางได้ไม่เกิน 50 ท่านต่อการจอง';
  if (v.adultPrice === null || v.adultPrice === undefined) return 'กรุณาระบุราคาผู้ใหญ่';
  if (n(v.childCount) > 0 && v.childPrice == null) return 'กรุณาระบุราคาเด็ก';
  if (n(v.infantCount) > 0 && v.infantPrice == null) return 'กรุณาระบุราคาทารก';
  if (n(v.singleRoomCount) > adults) return 'ห้องพักเดี่ยวมากกว่าจำนวนผู้ใหญ่';
  if (n(v.singleRoomCount) > 0 && v.singleRoomPrice == null) return 'กรุณาระบุราคาห้องพักเดี่ยว';
  if ((v.options ?? []).some((o) => n(o?.qty) > 0 && n(o?.price) > 0 && !o?.name?.trim())) return 'กรุณาระบุชื่อรายการเพิ่มเติม';
  if (totals.subtotal < 0) return 'ส่วนลดมากกว่าราคา';
  if (v.isDeposit && n(v.depositPerPerson) <= 0) return 'กรุณาระบุมัดจำต่อท่าน';
  if (v.isDeposit && totals.deposit > totals.subtotal) return 'มัดจำมากกว่ายอดรวม';
  return null;
}

// The charge lines sent to the backend: counts resolved, unused lines dropped.
function chargeLines(v: DealValues, travellers: number): LineBookingCharge[] {
  const lines: LineBookingCharge[] = [];
  const add = (kind: LineBookingCharge['kind'], count: number, price: Num) => {
    if (count > 0 && n(price) > 0) lines.push({ kind, name: '', count, price: n(price) });
  };
  add('visa', n(v.visaCount), v.visaPrice);
  add('foreignPassport', n(v.foreignCount), v.foreignPrice);
  add('joinLand', n(v.joinLandCount), v.joinLandPrice);
  add('fuelTax', v.fuelTaxCount == null ? travellers : n(v.fuelTaxCount), v.fuelTaxPrice);
  add('airportTax', v.airportTaxCount == null ? travellers : n(v.airportTaxCount), v.airportTaxPrice);
  for (const o of v.options ?? []) {
    if (n(o?.qty) > 0 && n(o?.price) > 0) lines.push({ kind: 'option', name: o?.name?.trim() || 'Option', count: n(o?.qty), price: n(o?.price) });
  }
  return lines;
}

// The deal part of POST /line-bookings and PUT /line-bookings/editable.
export function dealRequest(v: DealValues, adultPriceBeforeDiscount: number | null): LineBookingDeal {
  const adults = n(v.adultCount);
  const single = Math.min(n(v.singleRoomCount), adults);
  const children = n(v.childCount);
  const infants = n(v.infantCount);
  const travellers = adults + children + infants;
  return {
    adultTwin: adults - single,
    adultSingle: single,
    childWithBed: children,
    childNoBed: 0,
    infant: infants,
    adultPrice: n(v.adultPrice),
    singleSupplement: single > 0 ? n(v.singleRoomPrice) : null,
    childWithBedPrice: children > 0 ? n(v.childPrice) : null,
    childNoBedPrice: null,
    infantPrice: infants > 0 ? n(v.infantPrice) : null,
    adultPriceBeforeDiscount,
    discountPerPerson: n(v.discountPerPerson),
    charges: chargeLines(v, travellers),
    paymentMethod: v.paymentMethod,
    creditFeePercent: v.paymentMethod === 'credit' ? n(v.creditFeePercent) : 0,
    isDeposit: v.isDeposit,
    depositPerPerson: v.isDeposit ? n(v.depositPerPerson) : 0,
    depositDueDate: v.isDeposit ? date(v.depositDueDate) : null,
    depositDueTime: v.isDeposit ? time(v.depositDueTime) : '',
    remainingDueDate: v.isDeposit ? date(v.remainingDueDate) : null,
    fullDueDate: v.isDeposit ? null : date(v.fullDueDate),
    fullDueTime: v.isDeposit ? '' : time(v.fullDueTime),
  };
}

const priced = (line: PricedLine | null): [Num, Num] => (line ? [line.count || null, line.price] : [null, null]);

// The deal from a pasted summary: the same rows, so it maps one to one.
export function dealFromSummary(s: ParsedSummary): DealValues {
  const [childCount, childPrice] = priced(s.child);
  const [infantCount, infantPrice] = priced(s.infant);
  const [singleRoomCount, singleRoomPrice] = priced(s.singleRoom);
  const [visaCount, visaPrice] = priced(s.visa);
  const [foreignCount, foreignPrice] = priced(s.foreignPassport);
  const [joinLandCount, joinLandPrice] = priced(s.joinLand);
  const [fuelTaxCount, fuelTaxPrice] = priced(s.fuelTax);
  const [airportTaxCount, airportTaxPrice] = priced(s.airportTax);
  const adults = Math.max(0, (s.travellers ?? 0) - n(childCount) - n(infantCount));
  return {
    adultCount: adults || null,
    adultPrice: s.adultPrice,
    childCount,
    childPrice,
    infantCount,
    infantPrice,
    visaCount,
    visaPrice,
    foreignCount,
    foreignPrice,
    singleRoomCount,
    singleRoomPrice,
    joinLandCount,
    joinLandPrice,
    fuelTaxCount,
    fuelTaxPrice,
    airportTaxCount,
    airportTaxPrice,
    options: s.options.map((o) => ({ name: o.name, qty: o.count, price: o.price })),
    discountPerPerson: s.discountPerPerson,
    paymentMethod: s.creditFeePercent !== null ? 'credit' : 'transfer',
    creditFeePercent: s.creditFeePercent ?? DEFAULT_CREDIT_FEE_PERCENT,
    isDeposit: s.isDeposit ?? false,
    depositPerPerson: s.depositPerPerson,
    depositDueDate: day(s.depositDueDate),
    depositDueTime: clock(s.depositDueTime),
    remainingDueDate: day(s.remainingDueDate),
    fullDueDate: day(s.fullDueDate),
    fullDueTime: clock(s.fullDueTime),
  };
}

// The deal of a reopened booking: the backend's room types back to the summary rows (the reverse
// of dealRequest). Adults = twin + single, single room price = the supplement, children = with + no
// bed. The payment plan (full or deposit) is kept.
export function dealFromBooking(b: LineBooking): DealValues {
  const adults = b.adultTwinCount + b.adultSingleCount;
  const children = b.childWithBedCount + b.childNoBedCount;
  const adultPrice = b.adultTwinPrice;
  const charge = (kind: LineBookingCharge['kind']): [Num, Num] => {
    const c = b.charges.find((x) => x.kind === kind);
    return c ? [c.count, c.price] : [null, null];
  };
  const [visaCount, visaPrice] = charge('visa');
  const [foreignCount, foreignPrice] = charge('foreignPassport');
  const [joinLandCount, joinLandPrice] = charge('joinLand');
  const [fuelTaxCount, fuelTaxPrice] = charge('fuelTax');
  const [airportTaxCount, airportTaxPrice] = charge('airportTax');
  return {
    adultCount: adults || null,
    adultPrice,
    childCount: children || null,
    childPrice: b.childWithBedPrice ?? b.childNoBedPrice ?? (children > 0 ? adultPrice : null),
    infantCount: b.infantCount || null,
    infantPrice: b.infantPrice,
    visaCount,
    visaPrice,
    foreignCount,
    foreignPrice,
    singleRoomCount: b.adultSingleCount || null,
    singleRoomPrice: b.adultSinglePrice !== null ? b.adultSinglePrice - adultPrice : null,
    joinLandCount,
    joinLandPrice,
    fuelTaxCount,
    fuelTaxPrice,
    airportTaxCount,
    airportTaxPrice,
    options: b.charges.filter((c) => c.kind === 'option').map((c) => ({ name: c.name, qty: c.count, price: c.price })),
    discountPerPerson: b.discountPerPerson || null,
    paymentMethod: b.paymentMethod,
    creditFeePercent: b.paymentMethod === 'credit' ? b.creditFeePercent : DEFAULT_CREDIT_FEE_PERCENT,
    isDeposit: b.isDeposit,
    depositPerPerson: b.depositPerPerson || null,
    depositDueDate: day(b.depositDueDate),
    depositDueTime: clock(b.depositDueTime),
    remainingDueDate: day(b.remainingDueDate),
    fullDueDate: day(b.fullDueDate),
    fullDueTime: clock(b.fullDueTime),
  };
}

// A deposit booking not yet paid in full whose deposit documents were approved can take the balance.
export function canRecordBalance(b: LineBooking): boolean {
  return b.isDeposit && b.state === LineBookingStates.DepositPaid && b.documentsApprovedAt !== null;
}

// Why saving a booking whose documents aren't approved yet needs care: finance may issue them for
// the old amounts.
export function financeWarning(b: LineBooking): string {
  const finance = b.financeUser?.name ?? 'ฝ่ายการเงิน';
  return b.documentsAt
    ? `${finance} ส่งเอกสารการเงินแล้วแต่ยังรอเซลล์ตรวจ เอกสารอาจไม่ตรงกับข้อมูลที่แก้ไข`
    : `ใบจองนี้รอ ${finance} ออกเอกสารการเงิน การแก้ไขตอนนี้อาจทำให้เอกสารไม่ตรงกับยอดใหม่`;
}
