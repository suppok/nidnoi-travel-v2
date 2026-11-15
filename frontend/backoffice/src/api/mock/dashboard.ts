import { Permissions } from '../../constants/permissions.ts';
import { Roles } from '../../constants/roles.ts';
import {
  DASHBOARD_DAYS,
  type DashboardBucket,
  type DashboardLineBookings,
  type DashboardLineGroup,
  type DashboardSummary,
  type DashboardUnit,
} from '../../types/dashboard.ts';
import { asRecord, authenticate, type MockHandler, ok, validation } from './http.ts';
import { permissionsFor } from './policy.ts';

// Mock-only figures: synthetic but stable (seeded by date), so the dashboard can be designed and
// demoed without data. The real numbers come from POST /dashboard/summary aggregates.

// mulberry32: a tiny seeded PRNG.
function rng(seed: number) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function seedOf(text: string): number {
  let h = 2166136261;
  for (const ch of text) h = Math.imul(h ^ ch.charCodeAt(0), 16777619);
  return h >>> 0;
}

const iso = (d: Date) => d.toISOString().slice(0, 10);
const addDays = (d: Date, n: number) => new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate() + n));

function unitFor(days: number): DashboardUnit {
  return days > 120 ? 'month' : days > 31 ? 'week' : 'day';
}

function bucketStart(d: Date, unit: DashboardUnit): Date {
  if (unit === 'week') return addDays(d, -((d.getUTCDay() + 6) % 7));
  if (unit === 'month') return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), 1));
  return d;
}

function nextBucket(d: Date, unit: DashboardUnit): Date {
  if (unit === 'week') return addDays(d, 7);
  if (unit === 'month') return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + 1, 1));
  return addDays(d, 1);
}

// Today in Thailand as a UTC-midnight date, like the backend's TodayInThailand.
function todayInThailand(): Date {
  const t = new Date(Date.now() + 7 * 3600_000);
  return new Date(Date.UTC(t.getUTCFullYear(), t.getUTCMonth(), t.getUTCDate()));
}

function bucketDays(d: Date, unit: DashboardUnit): number {
  return Math.round((nextBucket(d, unit).getTime() - d.getTime()) / 86_400_000);
}

// Daily rates scaled by the bucket length; a gentle weekly wave plus noise.
function series(from: Date, to: Date, unit: DashboardUnit, perDay: number, scale: number, salt: string): DashboardBucket[] {
  const out: DashboardBucket[] = [];
  for (let d = from; d <= to; d = nextBucket(d, unit)) {
    const r = rng(seedOf(salt + iso(d)));
    const len = bucketDays(d, unit);
    const wave = 1 + 0.35 * Math.sin(d.getTime() / (86_400_000 * 3.5));
    const count = Math.max(0, Math.round(perDay * len * wave * (0.55 + r() * 0.9)));
    const paidShare = 0.55 + r() * 0.25;
    const revenue = Math.round((count * paidShare * (38_000 + r() * 30_000) * scale) / 100) * 100;
    out.push({ date: iso(d), count, revenue: salt.startsWith('web') ? 0 : revenue });
  }
  return out;
}

function sum(buckets: DashboardBucket[], key: 'count' | 'revenue'): number {
  return buckets.reduce((n, b) => n + b[key], 0);
}

function split(total: number, weights: number[]): number[] {
  const w = weights.reduce((a, b) => a + b, 0);
  const parts = weights.map((x) => Math.floor((total * x) / w));
  parts[0] += total - parts.reduce((a, b) => a + b, 0);
  return parts;
}

function groups(names: [string, string][], count: number, revenue: number, salt: string): DashboardLineGroup[] {
  const r = rng(seedOf(salt));
  const weights = names.map((_, i) => (names.length - i) * (0.6 + r() * 0.8));
  const counts = split(count, weights);
  const revenues = split(revenue, weights.map((x, i) => x * (0.8 + ((i * 7) % 5) * 0.1)));
  return names
    .map(([key, label], i) => {
      const paidCount = Math.round(counts[i] * 0.72);
      return { key, label, count: counts[i], paidCount, revenue: Math.round(revenues[i] / 100) * 100, travellers: Math.round(paidCount * 2.6) };
    })
    .filter((g) => g.count > 0)
    .sort((a, b) => b.revenue - a.revenue || b.count - a.count);
}

const SOURCES: [string, string][] = [
  ['src-line', 'LINE OA'],
  ['src-fb', 'Facebook'],
  ['src-web', 'เว็บไซต์'],
  ['src-tel', 'โทรศัพท์'],
  ['', ''],
];

const TOURS: [string, string][] = [
  ['NIDN000001', 'ญี่ปุ่น โตเกียว ฟูจิ 5 วัน 3 คืน'],
  ['NIDN000002', 'ฮอกไกโด ซัปโปโร โอตารุ 6 วัน 4 คืน'],
  ['NIDN000003', 'เกาหลี โซล นามิ เอเวอร์แลนด์ 5 วัน 3 คืน'],
  ['NIDN000004', 'ไต้หวัน ไทเป อาลีซาน 5 วัน 4 คืน'],
  ['NIDN000005', 'โอซาก้า เกียวโต นารา 5 วัน 3 คืน'],
  ['NIDN000006', 'เวียดนามเหนือ ซาปา ฮานอย 4 วัน 3 คืน'],
];

const SALES: [string, string][] = [
  ['0b9c6a52-0000-4000-8000-000000000002', 'สมใจ ขายเก่ง'],
  ['0b9c6a52-0000-4000-8000-000000000003', 'มานี บัญชีดี'],
  ['0b9c6a52-0000-4000-8000-000000000004', 'พนักงานใหม่'],
];

const CUSTOMERS = ['คุณสมชาย ใจดี', 'คุณวิภา รักเที่ยว', 'คุณอนันต์ สุขสันต์', 'คุณพรทิพย์ แสงทอง', 'คุณกิตติ มีสุข', 'คุณนภา ฟ้าใส'];

function lineSection(from: Date, to: Date, prevFrom: Date, unit: DashboardUnit, scopeAll: boolean, today: Date): DashboardLineBookings {
  const scale = scopeAll ? 1 : 0.4;
  const perDay = scopeAll ? 3.2 : 1.3;
  const buckets = series(from, to, unit, perDay, 1, `line${scale}`);
  const prev = series(bucketStart(prevFrom, unit), addDays(from, -1), unit, perDay * 0.9, 1, `lineprev${scale}`);
  const totals = (b: DashboardBucket[]) => {
    const count = sum(b, 'count');
    const revenue = sum(b, 'revenue');
    const paidCount = Math.round(count * 0.68);
    return { count, paidCount, cancelled: Math.round(count * 0.07), revenue, travellers: Math.round(paidCount * 2.7), averageValue: paidCount ? Math.round(revenue / paidCount) : 0 };
  };
  const t = totals(buckets);
  const newCount = Math.max(0, t.count - t.paidCount - t.cancelled);
  const deposit = Math.round(t.paidCount * 0.4);
  const r = rng(seedOf(`dep${iso(today)}${scale}`));
  return {
    scopeAll,
    totals: t,
    previous: totals(prev),
    buckets,
    byState: [
      { state: 10, count: newCount },
      { state: 20, count: deposit },
      { state: 25, count: t.paidCount - deposit },
      { state: 30, count: t.cancelled },
    ].filter((s) => s.count > 0),
    bySource: groups(SOURCES, t.count, t.revenue, `src${scale}`),
    byTour: groups(TOURS, t.count, t.revenue, `tour${scale}`),
    bySales: scopeAll ? groups(SALES, t.count, t.revenue, 'sales') : [],
    pipeline: { waitingConfirm: Math.min(newCount, 4), waitingDocuments: 3, waitingReview: 2, balancePending: 1 },
    departuresDays: 14,
    departures: Array.from({ length: scopeAll ? 7 : 3 }, (_, i) => {
      const [tourCode, tourName] = TOURS[i % TOURS.length];
      const start = addDays(today, 1 + i * 2);
      const travellers = 1 + Math.floor(r() * 5);
      const [salesId, salesName] = SALES[i % SALES.length];
      return {
        id: `mock-departure-${i}`,
        bookingNo: `26DEP${String(i).padStart(3, '0')}`,
        tourCode,
        tourName,
        startDate: iso(start),
        endDate: iso(addDays(start, 4)),
        customerName: CUSTOMERS[i % CUSTOMERS.length],
        travellers,
        state: i % 3 === 0 ? 20 : 25,
        totalPrice: travellers * (32_900 + i * 2_000),
        salesUser: { id: salesId, name: salesName },
      };
    }),
  };
}

export const dashboardRoutes: Record<string, MockHandler> = {
  'POST /dashboard/summary': (req) => {
    const user = authenticate(req, 'POST /dashboard/summary');
    const body = asRecord(req.body);
    const days = body.days === undefined || body.days === 0 ? 30 : Number(body.days);
    if (!(DASHBOARD_DAYS as readonly number[]).includes(days)) throw validation({ days: 'must be a valid value' });

    const perms = permissionsFor(user.roles);
    const scopeAll = user.roles.some((r) => r === Roles.Admin || r === Roles.SuperAdmin);
    const unit = unitFor(days);
    const today = todayInThailand();
    const from = bucketStart(addDays(today, -(days - 1)), unit);
    const length = Math.round((today.getTime() - from.getTime()) / 86_400_000) + 1;
    const prevFrom = addDays(from, -length);

    const webBuckets = series(from, today, unit, 1.4, 1, 'web');
    const webPrev = series(bucketStart(prevFrom, unit), addDays(from, -1), unit, 1.2, 1, 'webprev');
    const webTotal = sum(webBuckets, 'count');
    const [webNew, webContacted, webConfirmed, webCancelled] = split(webTotal, [2, 4, 3, 1]);

    const r = rng(seedOf(`tours${iso(today)}`));
    const summary: DashboardSummary = {
      range: { from: iso(from), to: iso(today), days, unit, previousFrom: iso(prevFrom), previousTo: iso(addDays(from, -1)) },
      generatedAt: new Date().toISOString(),
      tours: perms.includes(Permissions.TourProgramList)
        ? {
            byState: [
              { state: 10, count: 6 },
              { state: 20, count: 42 },
              { state: 30, count: 9 },
              { state: 40, count: 3 },
            ],
            byCountry: [
              ['japan', 'ญี่ปุ่น', 21],
              ['korea', 'เกาหลี', 11],
              ['taiwan', 'ไต้หวัน', 7],
              ['vietnam', 'เวียดนาม', 5],
              ['china', 'จีน', 4],
              ['europe', 'ยุโรป', 3],
            ].map(([countryId, nameTh, count]) => ({ countryId: String(countryId), nameTh: String(nameTh), count: Number(count) })),
            seats: { days: 30, lowSeatsThreshold: 5, periods: 64, seatsTotal: 1_920, seatsAvailable: 742, soldOut: 9, lowSeats: 12 },
            lowSeatPeriods: TOURS.map(([tourCode, tourName], i) => {
              const start = addDays(today, 3 + Math.floor(r() * 25));
              return {
                periodId: `mock-period-${i}`,
                tourProgramId: `mock-tour-${i}`,
                tourCode,
                tourName,
                startDate: iso(start),
                endDate: iso(addDays(start, 4)),
                price: 25_900 + i * 3_000,
                seatsTotal: 30,
                seatsAvailable: 1 + i,
              };
            }),
          }
        : null,
      bookings: perms.includes(Permissions.BookingList)
        ? {
            total: webTotal,
            previousTotal: sum(webPrev, 'count'),
            newCount: Math.max(webNew, 2),
            byState: [
              { state: 10, count: webNew },
              { state: 20, count: webContacted },
              { state: 30, count: webConfirmed },
              { state: 40, count: webCancelled },
            ].filter((s) => s.count > 0),
            buckets: webBuckets,
          }
        : null,
      lineBookings: perms.includes(Permissions.LineBookingList) ? lineSection(from, today, prevFrom, unit, scopeAll, today) : null,
    };
    return ok(summary);
  },
};
