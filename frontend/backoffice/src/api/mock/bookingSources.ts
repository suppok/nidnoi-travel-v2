import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { ListResponse } from '../../types/api.ts';
import type { BookingSource } from '../../types/bookingSource.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';

// Mock-only persistence. The backend table starts empty; these samples just make the tools usable.
const DB_KEY = 'nidnoi.mock.booking-sources.v1';

function seed(): BookingSource[] {
  const base = { status: true, createDate: '2026-11-11T09:00:00.000Z', updateDate: '2026-11-11T09:00:00.000Z' };
  return [
    { ...base, id: 'b5000000-0000-4000-8000-000000000001', name: 'LINE OA', shortName: 'LINE', sortOrder: 10 },
    { ...base, id: 'b5000000-0000-4000-8000-000000000002', name: 'Facebook', shortName: 'Facebook', sortOrder: 20 },
    { ...base, id: 'b5000000-0000-4000-8000-000000000003', name: 'เอเจนซี่', shortName: 'เอเจนซี่', sortOrder: 30 },
    { ...base, id: 'b5000000-0000-4000-8000-000000000004', name: 'Walk-in', shortName: '', sortOrder: 40 },
  ];
}

function load(): BookingSource[] {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as BookingSource[];
  } catch {
    // fall through to seed
  }
  return seed();
}

const sources: BookingSource[] = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(sources));
  } catch {
    // storage unavailable: keep in memory only
  }
}

function find(id: unknown): BookingSource {
  const s = sources.find((x) => x.id === id);
  if (!s) throw new MockFail(404, ErrorCodes.NotFound, 'Booking source not found');
  return s;
}

function parse(body: Record<string, unknown>, selfId: string | null) {
  const fields: Record<string, string> = {};
  const name = str(body.name);
  const shortName = str(body.shortName);
  const sortOrder = Number(body.sortOrder ?? 0);

  if (!name) fields.name = 'กรุณากรอกชื่อ';
  else if (name.length > 100) fields.name = 'ยาวได้ไม่เกิน 100 ตัวอักษร';
  if (shortName.length > 50) fields.shortName = 'ยาวได้ไม่เกิน 50 ตัวอักษร';
  if (!Number.isInteger(sortOrder) || sortOrder < 0 || sortOrder > 9999) fields.sortOrder = 'ต้องเป็นจำนวนเต็ม 0–9999';
  if (Object.keys(fields).length) throw validation(fields);

  if (sources.some((s) => s.name.toLowerCase() === name.toLowerCase() && s.id !== selfId)) {
    throw new MockFail(409, ErrorCodes.NameTaken, 'Name is already in use');
  }
  return { name, shortName, sortOrder };
}

const list: MockHandler = (req) => {
  authorize(req, 'POST /booking-sources/list', Permissions.BookingSourceList);
  const body = asRecord(req.body);
  const keyword = str(body.keyword).toLowerCase();
  const page = Math.max(1, Number(body.page) || 1);
  const pageSize = Number(body.pageSize) || 20;
  if (pageSize > 500) throw validation({ pageSize: 'must be no greater than 500' });

  const rows = sources
    .filter((s) => !keyword || s.name.toLowerCase().includes(keyword) || s.shortName.toLowerCase().includes(keyword))
    .filter((s) => typeof body.status !== 'boolean' || s.status === body.status)
    .sort((a, b) => a.sortOrder - b.sortOrder || a.name.localeCompare(b.name, 'th'));

  return ok({ items: rows.slice((page - 1) * pageSize, page * pageSize), total: rows.length, page, pageSize } satisfies ListResponse<BookingSource>);
};

export const bookingSourceRoutes: Record<string, MockHandler> = {
  'POST /booking-sources/list': list,

  'GET /booking-sources/:id': (req) => {
    authorize(req, 'GET /booking-sources/:id', Permissions.BookingSourceGet);
    return ok(find(req.params.id));
  },

  'POST /booking-sources': (req) => {
    authorize(req, 'POST /booking-sources', Permissions.BookingSourceCreate);
    const now = new Date().toISOString();
    const s: BookingSource = { id: crypto.randomUUID(), ...parse(asRecord(req.body), null), status: true, createDate: now, updateDate: now };
    sources.push(s);
    save();
    return ok(s);
  },

  'PUT /booking-sources': (req) => {
    authorize(req, 'PUT /booking-sources', Permissions.BookingSourceUpdate);
    const body = asRecord(req.body);
    const s = find(body.id);
    Object.assign(s, parse(body, s.id), { updateDate: new Date().toISOString() });
    save();
    return ok(s);
  },

  'PUT /booking-sources/status': (req) => {
    authorize(req, 'PUT /booking-sources/status', Permissions.BookingSourceStatus);
    const body = asRecord(req.body);
    if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
    const s = find(body.id);
    s.status = body.status;
    s.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};
