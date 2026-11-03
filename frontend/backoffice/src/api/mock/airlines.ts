import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { Airline } from '../../types/airline.ts';
import type { ListResponse } from '../../types/api.ts';
import { SEED_AIRLINES } from './airlineSeed.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';

// Mock-only persistence for airlines, seeded like the backend migration *_seed_airline.
const DB_KEY = 'nidnoi.mock.airlines.v2';
const CODE_RE = /^[A-Z0-9]{2,3}$/;

function seed(): Airline[] {
  const base = { logo: null, status: true, createDate: '2026-11-03T09:00:00.000Z', updateDate: '2026-11-03T09:00:00.000Z' };
  return SEED_AIRLINES.map(([id, refId, code, name]) => ({ ...base, id, refId, code, name }));
}

function load(): Airline[] {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Airline[];
  } catch {
    // fall through to seed
  }
  return seed();
}

const airlines: Airline[] = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(airlines));
  } catch {
    // storage unavailable: keep in memory only
  }
}

function find(id: unknown): Airline {
  const a = airlines.find((x) => x.id === id);
  if (!a) throw new MockFail(404, ErrorCodes.NotFound, 'Airline not found');
  return a;
}

function parse(body: Record<string, unknown>, selfId: string | null) {
  const fields: Record<string, string> = {};
  const code = str(body.code).toUpperCase();
  const name = str(body.name);
  const refId = body.refId === null || body.refId === undefined ? null : Number(body.refId);
  const logo = str(body.logo) || null;

  if (!code) fields.code = 'กรุณากรอกรหัส';
  else if (!CODE_RE.test(code)) fields.code = 'ใช้รหัส IATA 2 ตัว หรือ ICAO 3 ตัว (A-Z, 0-9)';
  if (!name) fields.name = 'กรุณากรอกชื่อ';
  else if (name.length > 200) fields.name = 'ยาวได้ไม่เกิน 200 ตัวอักษร';
  if (refId !== null && (!Number.isInteger(refId) || refId < 1)) fields.refId = 'ต้องเป็นจำนวนเต็มบวก';
  if (Object.keys(fields).length) throw validation(fields);

  if (airlines.some((a) => a.code === code && a.id !== selfId)) throw new MockFail(409, ErrorCodes.CodeTaken, 'Code is already in use');
  if (refId !== null && airlines.some((a) => a.refId === refId && a.id !== selfId)) throw new MockFail(409, ErrorCodes.RefIdTaken, 'Ref ID is already in use');
  return { code, name, refId, logo };
}

const list: MockHandler = (req) => {
  authorize(req, 'POST /airlines/list', Permissions.AirlineList);
  const body = asRecord(req.body);
  const keyword = str(body.keyword).toLowerCase();
  const page = Math.max(1, Number(body.page) || 1);
  const pageSize = Number(body.pageSize) || 20;
  if (pageSize > 500) throw validation({ pageSize: 'must be no greater than 500' });

  const rows = airlines
    .filter((a) => !keyword || a.name.toLowerCase().includes(keyword) || a.code.toLowerCase().includes(keyword))
    .filter((a) => typeof body.status !== 'boolean' || a.status === body.status)
    .sort((a, b) => a.name.localeCompare(b.name, 'th'));

  return ok({ items: rows.slice((page - 1) * pageSize, page * pageSize), total: rows.length, page, pageSize } satisfies ListResponse<Airline>);
};

export const airlineRoutes: Record<string, MockHandler> = {
  'POST /airlines/list': list,

  'GET /airlines/:id': (req) => {
    authorize(req, 'GET /airlines/:id', Permissions.AirlineGet);
    return ok(find(req.params.id));
  },

  'POST /airlines': (req) => {
    authorize(req, 'POST /airlines', Permissions.AirlineCreate);
    const now = new Date().toISOString();
    const a: Airline = { id: crypto.randomUUID(), ...parse(asRecord(req.body), null), status: true, createDate: now, updateDate: now };
    airlines.push(a);
    save();
    return ok(a);
  },

  'PUT /airlines': (req) => {
    authorize(req, 'PUT /airlines', Permissions.AirlineUpdate);
    const body = asRecord(req.body);
    const a = find(body.id);
    Object.assign(a, parse(body, a.id), { updateDate: new Date().toISOString() });
    save();
    return ok(a);
  },

  'PUT /airlines/status': (req) => {
    authorize(req, 'PUT /airlines/status', Permissions.AirlineStatus);
    const body = asRecord(req.body);
    if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
    const a = find(body.id);
    a.status = body.status;
    a.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};

// Lookup for other mocks (tour programs join their airline on read).
export function mockAirline(id: string): Airline | undefined {
  return airlines.find((a) => a.id === id);
}
