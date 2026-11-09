import { BOOKING_STATE_INFO, type BookingState, BookingStates } from '../../constants/bookingStates.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { ListResponse } from '../../types/api.ts';
import type { Booking } from '../../types/booking.ts';
import { users } from './db.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';
import { ErrorCodes } from '../../constants/errorCodes.ts';

// Mock-only booking requests. The real ones come from the customer site (POST /bookings/public).
// In the browser console, `__mockNewBooking()` adds a new one to try the new-booking alert.
const DB_KEY = 'nidnoi.mock.bookings.v2';
const KNOWN_STATES = new Set<number>(BOOKING_STATE_INFO.map((s) => s.value));
const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

type Store = { bookings: Booking[] };

// Like the backend: 2-digit year + 6 random characters without 0/O/1/I/L.
const NO_ALPHABET = '23456789ABCDEFGHJKMNPQRSTUVWXYZ';
const randomNo = () => String(new Date().getFullYear()).slice(2) + Array.from({ length: 6 }, () => NO_ALPHABET[Math.floor(Math.random() * NO_ALPHABET.length)]).join('');

const SAMPLE_CUSTOMERS = [
  { firstName: 'สมชาย', lastName: 'ใจดี', email: 'somchai@example.com', phone: '081-234-5678', note: 'ขอห้องติดกัน 2 ห้อง' },
  { firstName: 'วิภา', lastName: 'รักเที่ยว', email: 'wipa@example.com', phone: '089-111-2222', note: '' },
  { firstName: 'John', lastName: 'Smith', email: 'john@example.com', phone: '+66 92 345 6789', note: 'แพ้อาหารทะเล' },
  { firstName: 'นภา', lastName: '', email: 'napa@example.com', phone: '0623334444', note: 'สะดวกให้ติดต่อหลัง 18:00' },
];

function makeBooking(minutesAgo: number, state: BookingState, i: number): Booking {
  const c = SAMPLE_CUSTOMERS[i % SAMPLE_CUSTOMERS.length];
  const at = new Date(Date.now() - minutesAgo * 60_000).toISOString();
  const twin = 2 + (i % 2);
  const single = i % 3 === 0 ? 1 : 0;
  const child = i % 2;
  const assignee = state === BookingStates.New ? null : { id: users[2].id, name: users[2].name };
  return {
    id: crypto.randomUUID(),
    bookingNo: randomNo(),
    tourProgramId: '00000000-0000-4000-8000-000000000001',
    tourProgramPeriodId: '00000000-0000-4000-8000-000000000002',
    tourCode: 'NIDN264159',
    tourName: 'ทัวร์จีน ชิงเต่า เว่ยไห่ เยียนไถ QINGDAO ทัวร์ เรียล เรียล...ชิงเต่า Autumn ฟินจัดเต็ม',
    startDate: '2026-11-12',
    endDate: '2026-11-18',
    adultTwinCount: twin,
    adultTwinPrice: 16919,
    adultSingleCount: single,
    adultSinglePrice: 21819,
    childWithBedCount: child,
    childWithBedPrice: 15919,
    childNoBedCount: 0,
    childNoBedPrice: null,
    totalPrice: twin * 16919 + single * 21819 + child * 15919,
    firstName: c.firstName,
    lastName: c.lastName,
    email: c.email,
    phone: c.phone,
    customerNote: c.note,
    state,
    assignee,
    staffNote: state === BookingStates.New ? '' : 'โทรแจ้งราคาแล้ว ลูกค้าขอคิดก่อน 1 วัน',
    status: true,
    createDate: at,
    updateDate: at,
  };
}

function seed(): Store {
  return {
    bookings: [
      makeBooking(12, BookingStates.New, 0),
      makeBooking(60 * 3, BookingStates.Contacted, 1),
      makeBooking(60 * 26, BookingStates.Confirmed, 2),
      makeBooking(60 * 50, BookingStates.Cancelled, 3),
    ],
  };
}

function load(): Store {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Store;
  } catch {
    // fall through to seed
  }
  return seed();
}

const db: Store = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(db));
  } catch {
    // storage unavailable: keep in memory only
  }
}

(window as unknown as { __mockNewBooking: () => string }).__mockNewBooking = () => {
  const b = makeBooking(0, BookingStates.New, db.bookings.length);
  db.bookings.unshift(b);
  save();
  return b.bookingNo;
};

function find(id: string): Booking {
  const b = UUID_RE.test(id) ? db.bookings.find((x) => x.id === id) : undefined;
  if (!b) throw new MockFail(404, ErrorCodes.NotFound, 'Not found');
  return b;
}

export const bookingRoutes: Record<string, MockHandler> = {
  'POST /bookings/list': (req) => {
    authorize(req, 'POST /bookings/list', Permissions.BookingList);
    const body = asRecord(req.body);
    const keyword = str(body.keyword).toLowerCase();
    const state = Number(body.state) || 0;
    if (state && !KNOWN_STATES.has(state)) throw validation({ state: 'must be a valid value' });
    const status = typeof body.status === 'boolean' ? body.status : undefined;
    const page = Math.max(1, Number(body.page) || 1);
    const pageSize = Number(body.pageSize) || 20;
    if (pageSize > 100) throw validation({ pageSize: 'must be no greater than 100' });

    const rows = db.bookings
      .filter((b) => !state || b.state === state)
      .filter((b) => status === undefined || b.status === status)
      .filter(
        (b) =>
          !keyword ||
          [b.bookingNo, `${b.firstName} ${b.lastName}`, b.email, b.phone, b.tourCode].some((v) => v.toLowerCase().includes(keyword)),
      )
      .sort((a, b) => b.createDate.localeCompare(a.createDate));
    return ok({ items: rows.slice((page - 1) * pageSize, page * pageSize), total: rows.length, page, pageSize } satisfies ListResponse<Booking>);
  },

  'GET /bookings/:id': (req) => {
    authorize(req, 'GET /bookings/:id', Permissions.BookingGet);
    return ok(find(req.params.id));
  },

  'PUT /bookings': (req) => {
    const user = authorize(req, 'PUT /bookings', Permissions.BookingUpdate);
    const body = asRecord(req.body);
    const target = find(str(body.id));
    const state = Number(body.state);
    if (!KNOWN_STATES.has(state)) throw validation({ state: 'must be a valid value' });
    const staffNote = str(body.staffNote);
    if (staffNote.length > 2000) throw validation({ staffNote: 'the length must be between 0 and 2000' });

    if (!target.assignee && state !== BookingStates.New) target.assignee = { id: user.id, name: user.name };
    target.state = state as BookingState;
    target.staffNote = staffNote;
    target.updateDate = new Date().toISOString();
    save();
    return ok(target);
  },

  'PUT /bookings/status': (req) => {
    authorize(req, 'PUT /bookings/status', Permissions.BookingStatus);
    const body = asRecord(req.body);
    if (typeof body.status !== 'boolean') throw validation({ status: 'is required' });
    const target = find(str(body.id));
    target.status = body.status;
    target.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};
