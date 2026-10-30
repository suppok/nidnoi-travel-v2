import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import { WHOLESALE_STATE_INFO, type WholesaleState, WholesaleStates } from '../../constants/wholesaleStates.ts';
import type { ListResponse } from '../../types/api.ts';
import type { Wholesale, WholesaleContact } from '../../types/wholesale.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';
import { SEED_WHOLESALES } from './wholesaleSeed.ts';

// Mock-only persistence for wholesales (contacts stored inline; the backend uses a child table).
const DB_KEY = 'nidnoi.mock.wholesales.v2';
const CODE_RE = /^[A-Z0-9]+(-[A-Z0-9]+)*$/;
const MAX_CONTACTS = 20;
const KNOWN_STATES = new Set<number>(WHOLESALE_STATE_INFO.map((s) => s.value));

function seed(): Wholesale[] {
  const base = { status: true, createDate: '2026-10-30T09:00:00.000Z', updateDate: '2026-10-30T09:00:00.000Z' };
  return SEED_WHOLESALES.map(([id, refId, code, name, state, contacts]) => ({
    ...base,
    id,
    refId,
    code,
    name,
    state: state as WholesaleState,
    contacts: contacts.map(([contactName, phone]): WholesaleContact => ({ id: crypto.randomUUID(), name: contactName, phone })),
  }));
}

function load(): Wholesale[] {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Wholesale[];
  } catch {
    // fall through to seed
  }
  return seed();
}

const wholesales: Wholesale[] = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(wholesales));
  } catch {
    // storage unavailable: keep in memory only
  }
}

function find(id: unknown): Wholesale {
  const w = wholesales.find((x) => x.id === id);
  if (!w) throw new MockFail(404, ErrorCodes.NotFound, 'Wholesale not found');
  return w;
}

function parse(body: Record<string, unknown>, selfId: string | null) {
  const fields: Record<string, string> = {};
  const code = str(body.code).toUpperCase();
  const name = str(body.name);
  const refId = body.refId === null || body.refId === undefined ? null : Number(body.refId);
  const state = (body.state === undefined || body.state === 0 ? WholesaleStates.OnRequest : Number(body.state)) as WholesaleState;
  const rawContacts = Array.isArray(body.contacts) ? body.contacts : [];
  const contacts = rawContacts.map((c) => ({ id: crypto.randomUUID(), name: str(asRecord(c).name), phone: str(asRecord(c).phone) }));

  if (!code) fields.code = 'กรุณากรอกรหัส';
  else if (code.length > 20 || !CODE_RE.test(code)) fields.code = 'ใช้ได้เฉพาะ A-Z, 0-9 และขีด (-) ไม่เกิน 20 ตัวอักษร';
  if (!name) fields.name = 'กรุณากรอกชื่อ';
  else if (name.length > 200) fields.name = 'ยาวได้ไม่เกิน 200 ตัวอักษร';
  if (refId !== null && (!Number.isInteger(refId) || refId < 1)) fields.refId = 'ต้องเป็นจำนวนเต็มบวก';
  if (!KNOWN_STATES.has(state)) fields.state = 'สถานะไม่ถูกต้อง';
  if (contacts.length > MAX_CONTACTS) fields.contacts = `มีผู้ติดต่อได้ไม่เกิน ${MAX_CONTACTS} คน`;
  else if (contacts.some((c) => !c.name || !c.phone)) fields.contacts = 'ผู้ติดต่อทุกคนต้องมีชื่อและเบอร์โทร';
  if (Object.keys(fields).length) throw validation(fields);

  if (wholesales.some((w) => w.code === code && w.id !== selfId)) throw new MockFail(409, ErrorCodes.CodeTaken, 'Code is already in use');
  if (refId !== null && wholesales.some((w) => w.refId === refId && w.id !== selfId)) throw new MockFail(409, ErrorCodes.RefIdTaken, 'Ref ID is already in use');
  return { code, name, refId, state, contacts };
}

const list: MockHandler = (req) => {
  authorize(req, 'POST /wholesales/list', Permissions.WholesaleList);
  const body = asRecord(req.body);
  const keyword = str(body.keyword).toLowerCase();
  const page = Math.max(1, Number(body.page) || 1);
  const pageSize = Number(body.pageSize) || 20;
  if (pageSize > 500) throw validation({ pageSize: 'must be no greater than 500' });
  if (body.state !== undefined && body.state !== 0 && !KNOWN_STATES.has(Number(body.state))) throw validation({ state: 'must be a valid value' });

  const rows = wholesales
    .filter(
      (w) =>
        !keyword ||
        w.name.toLowerCase().includes(keyword) ||
        w.code.toLowerCase().includes(keyword) ||
        w.contacts.some((c) => c.name.toLowerCase().includes(keyword) || c.phone.includes(keyword)),
    )
    .filter((w) => !body.state || w.state === body.state)
    .filter((w) => typeof body.status !== 'boolean' || w.status === body.status)
    .sort((a, b) => a.name.localeCompare(b.name, 'th'));

  return ok({ items: rows.slice((page - 1) * pageSize, page * pageSize), total: rows.length, page, pageSize } satisfies ListResponse<Wholesale>);
};

export const wholesaleRoutes: Record<string, MockHandler> = {
  'POST /wholesales/list': list,

  'GET /wholesales/:id': (req) => {
    authorize(req, 'GET /wholesales/:id', Permissions.WholesaleGet);
    return ok(find(req.params.id));
  },

  'POST /wholesales': (req) => {
    authorize(req, 'POST /wholesales', Permissions.WholesaleCreate);
    const now = new Date().toISOString();
    const w: Wholesale = { id: crypto.randomUUID(), ...parse(asRecord(req.body), null), status: true, createDate: now, updateDate: now };
    wholesales.push(w);
    save();
    return ok(w);
  },

  'PUT /wholesales': (req) => {
    authorize(req, 'PUT /wholesales', Permissions.WholesaleUpdate);
    const body = asRecord(req.body);
    const w = find(body.id);
    Object.assign(w, parse(body, w.id), { updateDate: new Date().toISOString() });
    save();
    return ok(w);
  },

  'PUT /wholesales/status': (req) => {
    authorize(req, 'PUT /wholesales/status', Permissions.WholesaleStatus);
    const body = asRecord(req.body);
    if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
    const w = find(body.id);
    w.status = body.status;
    w.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};
