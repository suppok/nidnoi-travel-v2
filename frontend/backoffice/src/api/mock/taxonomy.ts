import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { ListResponse } from '../../types/api.ts';
import type { City, Country } from '../../types/taxonomy.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';
import { SEED_CITIES, SEED_COUNTRIES } from './taxonomySeed.ts';

// Mock-only persistence for countries and cities. Cities store countryId; the country summary is
// joined on read, like the backend's Relation("Country").
type StoredCity = Omit<City, 'country'>;
type Store = { countries: Country[]; cities: StoredCity[] };

const DB_KEY = 'nidnoi.mock.taxonomy.v1';
const SLUG_RE = /^[a-z0-9]+(-[a-z0-9]+)*$/;

function seed(): Store {
  const base = { description: '', heroImage: null, sortOrder: 0, status: true, createDate: '2026-10-30T09:00:00.000Z', updateDate: '2026-10-30T09:00:00.000Z' };
  const countries: Country[] = SEED_COUNTRIES.map(([refId, slug, nameTh, nameEn]) => ({ ...base, id: crypto.randomUUID(), refId, slug, nameTh, nameEn }));
  const idBySlug = new Map(countries.map((c) => [c.slug, c.id]));
  const cities: StoredCity[] = SEED_CITIES.map(([refId, countrySlug, slug, nameTh, nameEn]) => ({
    ...base,
    id: crypto.randomUUID(),
    refId,
    countryId: idBySlug.get(countrySlug) ?? '',
    slug,
    nameTh,
    nameEn,
  }));
  return { countries, cities };
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

function toCity(city: StoredCity): City {
  const c = db.countries.find((x) => x.id === city.countryId);
  return { ...city, country: { id: city.countryId, slug: c?.slug ?? '', nameTh: c?.nameTh ?? '', nameEn: c?.nameEn ?? '', status: c?.status ?? false } };
}

// --- Shared validation --------------------------------------------------------------------------

type Row = { id: string; refId: number | null; slug: string };

function parseFields(body: Record<string, unknown>) {
  const fields: Record<string, string> = {};
  const nameTh = str(body.nameTh);
  const nameEn = str(body.nameEn);
  const description = str(body.description);
  const heroImage = str(body.heroImage) || null;
  const refId = body.refId === null || body.refId === undefined ? null : Number(body.refId);
  const sortOrder = body.sortOrder === undefined ? 0 : Number(body.sortOrder);

  if (!nameTh) fields.nameTh = 'กรุณากรอกชื่อภาษาไทย';
  else if (nameTh.length > 100) fields.nameTh = 'ยาวได้ไม่เกิน 100 ตัวอักษร';
  if (!nameEn) fields.nameEn = 'กรุณากรอกชื่อภาษาอังกฤษ';
  else if (nameEn.length > 100) fields.nameEn = 'ยาวได้ไม่เกิน 100 ตัวอักษร';
  if (description.length > 5000) fields.description = 'ยาวได้ไม่เกิน 5000 ตัวอักษร';
  if (refId !== null && (!Number.isInteger(refId) || refId < 1)) fields.refId = 'ต้องเป็นจำนวนเต็มบวก';
  if (!Number.isInteger(sortOrder) || sortOrder < 0 || sortOrder > 100000) fields.sortOrder = 'ต้องเป็นจำนวนเต็ม 0–100000';
  return { fields, values: { nameTh, nameEn, description, heroImage, refId, sortOrder } };
}

function checkSlug(slug: string, fields: Record<string, string>) {
  if (!slug) fields.slug = 'กรุณากรอก slug';
  else if (slug.length > 60 || !SLUG_RE.test(slug)) fields.slug = 'ใช้ได้เฉพาะ a-z, 0-9 และขีด (-) ไม่เกิน 60 ตัวอักษร';
}

function ensureUnique(rows: Row[], selfId: string | null, slug: string | null, refId: number | null) {
  if (slug !== null && rows.some((r) => r.slug === slug && r.id !== selfId)) throw new MockFail(409, ErrorCodes.SlugTaken, 'Slug is already in use');
  if (refId !== null && rows.some((r) => r.refId === refId && r.id !== selfId)) throw new MockFail(409, ErrorCodes.RefIdTaken, 'Ref ID is already in use');
}

function listParams(body: Record<string, unknown>) {
  const pageSize = Number(body.pageSize) || 20;
  if (pageSize > 500) throw validation({ pageSize: 'must be no greater than 500' });
  return { keyword: str(body.keyword).toLowerCase(), page: Math.max(1, Number(body.page) || 1), pageSize: Math.max(1, pageSize) };
}

function matches(row: { nameTh: string; nameEn: string; slug: string }, keyword: string) {
  return !keyword || row.nameTh.includes(keyword) || row.nameEn.toLowerCase().includes(keyword) || row.slug.includes(keyword);
}

const byOrder = (a: { sortOrder: number; nameTh: string }, b: { sortOrder: number; nameTh: string }) =>
  a.sortOrder - b.sortOrder || a.nameTh.localeCompare(b.nameTh, 'th');

function page<T>(rows: T[], p: number, pageSize: number): ListResponse<T> {
  return { items: rows.slice((p - 1) * pageSize, p * pageSize), total: rows.length, page: p, pageSize };
}

function activeCountry(id: unknown): Country {
  const country = db.countries.find((c) => c.id === id);
  if (!country || !country.status) throw validation({ countryId: 'country not found' });
  return country;
}

function findCountry(id: unknown): Country {
  const country = db.countries.find((c) => c.id === id);
  if (!country) throw new MockFail(404, ErrorCodes.NotFound, 'Country not found');
  return country;
}

function findCity(id: unknown): StoredCity {
  const city = db.cities.find((c) => c.id === id);
  if (!city) throw new MockFail(404, ErrorCodes.NotFound, 'City not found');
  return city;
}

function statusOf(body: Record<string, unknown>): boolean {
  if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
  return body.status;
}

// --- Countries ----------------------------------------------------------------------------------

export const taxonomyRoutes: Record<string, MockHandler> = {
  'POST /countries/list': (req) => {
    authorize(req, 'POST /countries/list', Permissions.CountryList);
    const body = asRecord(req.body);
    const { keyword, page: p, pageSize } = listParams(body);
    const rows = db.countries
      .filter((c) => matches(c, keyword))
      .filter((c) => typeof body.status !== 'boolean' || c.status === body.status)
      .sort(byOrder);
    return ok(page(rows, p, pageSize));
  },

  'GET /countries/:id': (req) => {
    authorize(req, 'GET /countries/:id', Permissions.CountryGet);
    return ok(findCountry(req.params.id));
  },

  'POST /countries': (req) => {
    authorize(req, 'POST /countries', Permissions.CountryCreate);
    const body = asRecord(req.body);
    const slug = str(body.slug);
    const { fields, values } = parseFields(body);
    checkSlug(slug, fields);
    if (Object.keys(fields).length) throw validation(fields);
    ensureUnique(db.countries, null, slug, values.refId);

    const now = new Date().toISOString();
    const country: Country = { id: crypto.randomUUID(), slug, ...values, status: true, createDate: now, updateDate: now };
    db.countries.push(country);
    save();
    return ok(country);
  },

  'PUT /countries': (req) => {
    authorize(req, 'PUT /countries', Permissions.CountryUpdate);
    const body = asRecord(req.body);
    const country = findCountry(body.id);
    const { fields, values } = parseFields(body);
    if (Object.keys(fields).length) throw validation(fields);
    ensureUnique(db.countries, country.id, null, values.refId);

    Object.assign(country, values, { updateDate: new Date().toISOString() }); // slug is never written
    save();
    return ok(country);
  },

  'PUT /countries/status': (req) => {
    authorize(req, 'PUT /countries/status', Permissions.CountryStatus);
    const body = asRecord(req.body);
    const country = findCountry(body.id);
    country.status = statusOf(body);
    country.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },

  // --- Cities -----------------------------------------------------------------------------------

  'POST /cities/list': (req) => {
    authorize(req, 'POST /cities/list', Permissions.CityList);
    const body = asRecord(req.body);
    const { keyword, page: p, pageSize } = listParams(body);
    const countryId = str(body.countryId);
    const rows = db.cities
      .filter((c) => matches(c, keyword))
      .filter((c) => !countryId || c.countryId === countryId)
      .filter((c) => typeof body.status !== 'boolean' || c.status === body.status)
      .sort(byOrder);
    const result = page(rows, p, pageSize);
    return ok({ ...result, items: result.items.map(toCity) });
  },

  'GET /cities/:id': (req) => {
    authorize(req, 'GET /cities/:id', Permissions.CityGet);
    return ok(toCity(findCity(req.params.id)));
  },

  'POST /cities': (req) => {
    authorize(req, 'POST /cities', Permissions.CityCreate);
    const body = asRecord(req.body);
    const slug = str(body.slug);
    const { fields, values } = parseFields(body);
    checkSlug(slug, fields);
    if (!str(body.countryId)) fields.countryId = 'กรุณาเลือกประเทศ';
    if (Object.keys(fields).length) throw validation(fields);
    const country = activeCountry(body.countryId);
    ensureUnique(db.cities, null, slug, values.refId);

    const now = new Date().toISOString();
    const city: StoredCity = { id: crypto.randomUUID(), countryId: country.id, slug, ...values, status: true, createDate: now, updateDate: now };
    db.cities.push(city);
    save();
    return ok(toCity(city));
  },

  'PUT /cities': (req) => {
    authorize(req, 'PUT /cities', Permissions.CityUpdate);
    const body = asRecord(req.body);
    const city = findCity(body.id);
    const { fields, values } = parseFields(body);
    if (!str(body.countryId)) fields.countryId = 'กรุณาเลือกประเทศ';
    if (Object.keys(fields).length) throw validation(fields);
    if (body.countryId !== city.countryId) activeCountry(body.countryId);
    ensureUnique(db.cities, city.id, null, values.refId);

    Object.assign(city, values, { countryId: body.countryId as string, updateDate: new Date().toISOString() });
    save();
    return ok(toCity(city));
  },

  'PUT /cities/status': (req) => {
    authorize(req, 'PUT /cities/status', Permissions.CityStatus);
    const body = asRecord(req.body);
    const city = findCity(body.id);
    const status = statusOf(body);
    if (status && !city.status) activeCountry(city.countryId); // restoring needs an active country
    city.status = status;
    city.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};

// Lookups for other mocks (tour programs join countries and cities on read, like the backend).
export function mockCountry(id: string): Country | undefined {
  return db.countries.find((c) => c.id === id);
}

export function mockCity(id: string): StoredCity | undefined {
  return db.cities.find((c) => c.id === id);
}
