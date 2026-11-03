import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import { TOUR_PROGRAM_STATE_INFO, type TourProgramState, TourProgramStates } from '../../constants/tourProgramStates.ts';
import type { ListResponse } from '../../types/api.ts';
import type { TourProgram, TourProgramCity, TourProgramCountry, TourProgramListItem, TourProgramPeriod } from '../../types/tourProgram.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';
import { mockAirline } from './airlines.ts';
import { mockCity, mockCountry } from './taxonomy.ts';
import { mockWholesale } from './wholesales.ts';

// Mock-only persistence for tour programs and periods. Like the backend, tours store ids and join
// the wholesale, airline, countries and cities on read.
type StoredTour = Omit<TourProgram, 'wholesale' | 'airline' | 'countries' | 'cities' | 'images'> & {
  wholesaleId: string;
  airlineId?: string | null; // missing on tours stored before airlines were linked
  countryIds: string[];
  cityIds: string[];
  images: string[];
};
type Store = { nextCode: number; tours: StoredTour[]; periods: TourProgramPeriod[] };

const DB_KEY = 'nidnoi.mock.tour-programs.v1';
const CODE_PREFIX = 'NIDN'; // + 6 digits, like the backend (NIDN000001)
const KNOWN_STATES = new Set<number>(TOUR_PROGRAM_STATE_INFO.map((s) => s.value));
const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;

function load(): Store {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Store;
  } catch {
    // fall through to an empty store
  }
  return { nextCode: 1, tours: [], periods: [] };
}

const db: Store = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(db));
  } catch {
    // storage unavailable: keep in memory only
  }
}

function countrySummary(id: string): TourProgramCountry {
  const c = mockCountry(id);
  return { id, slug: c?.slug ?? '', nameTh: c?.nameTh ?? '', nameEn: c?.nameEn ?? '', status: c?.status ?? false };
}

function citySummary(id: string): TourProgramCity {
  const c = mockCity(id);
  return { id, countryId: c?.countryId ?? '', slug: c?.slug ?? '', nameTh: c?.nameTh ?? '', nameEn: c?.nameEn ?? '', status: c?.status ?? false };
}

function airlineSummary(id: string | null | undefined): TourProgram['airline'] {
  if (!id) return null;
  const a = mockAirline(id);
  return { id, code: a?.code ?? '', name: a?.name ?? '', logo: a?.logo ?? null, status: a?.status ?? false };
}

function toTour({ wholesaleId, airlineId, countryIds, cityIds, images, ...t }: StoredTour): TourProgram {
  const w = mockWholesale(wholesaleId);
  return {
    ...t,
    wholesale: { id: wholesaleId, code: w?.code ?? '', name: w?.name ?? '', status: w?.status ?? false },
    airline: airlineSummary(airlineId),
    countries: countryIds.map(countrySummary),
    cities: cityIds.map(citySummary),
    images: images.map((url, i) => ({ id: `${t.id}-${i}`, url })),
  };
}

function toListItem(stored: StoredTour): TourProgramListItem {
  const { description: _d, highlight: _h, airline: _a, pdfFile: _p, docxFile: _x, internalNote: _n, cities: _c, images: _i, ...item } = toTour(stored);
  return item;
}

function findTour(id: unknown): StoredTour {
  const t = db.tours.find((x) => x.id === id);
  if (!t) throw new MockFail(404, ErrorCodes.NotFound, 'Tour program not found');
  return t;
}

function findPeriod(id: unknown): TourProgramPeriod {
  const p = db.periods.find((x) => x.id === id);
  if (!p) throw new MockFail(404, ErrorCodes.NotFound, 'Period not found');
  return p;
}

const optionalNumber = (v: unknown) => (v === null || v === undefined || v === '' ? null : Number(v));
const ids = (v: unknown) => (Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : []);

// --- Tours ---------------------------------------------------------------------------------------

function parseTour(body: Record<string, unknown>, previous: StoredTour | null) {
  const fields: Record<string, string> = {};
  const name = str(body.name);
  const refId = optionalNumber(body.refId);
  const wholesaleId = str(body.wholesaleId);
  const airlineId = str(body.airlineId) || null;
  const state = (body.state === undefined || body.state === 0 ? TourProgramStates.Draft : Number(body.state)) as TourProgramState;
  const days = Number(body.days);
  const nights = Number(body.nights);
  const countryIds = ids(body.countryIds);
  const cityIds = ids(body.cityIds);
  const images = ids(body.images);

  if (!name) fields.name = 'กรุณากรอกชื่อโปรแกรม';
  else if (name.length > 200) fields.name = 'ยาวได้ไม่เกิน 200 ตัวอักษร';
  if (refId !== null && (!Number.isInteger(refId) || refId < 1)) fields.refId = 'ต้องเป็นจำนวนเต็มบวก';
  if (!KNOWN_STATES.has(state)) fields.state = 'สถานะไม่ถูกต้อง';
  if (!Number.isInteger(days) || days < 1 || days > 60) fields.days = 'ต้องเป็น 1–60 วัน';
  if (!Number.isInteger(nights) || nights < 0 || nights > days) fields.nights = 'จำนวนคืนต้องไม่เกินจำนวนวัน';
  if (str(body.wholesaleTourCode).length > 50) fields.wholesaleTourCode = 'ยาวได้ไม่เกิน 50 ตัวอักษร';
  if (str(body.highlight).length > 500) fields.highlight = 'ยาวได้ไม่เกิน 500 ตัวอักษร';
  if (images.length > 30) fields.images = 'มีรูปได้ไม่เกิน 30 รูป';

  const ws = mockWholesale(wholesaleId);
  if (!ws || (!ws.status && previous?.wholesaleId !== wholesaleId)) fields.wholesaleId = 'wholesale not found or deleted';

  if (airlineId) {
    const al = mockAirline(airlineId);
    if (!al || (!al.status && previous?.airlineId !== airlineId)) fields.airlineId = 'airline not found or deleted';
  }

  if (countryIds.length === 0) fields.countryIds = 'select at least one country';
  else if (countryIds.some((id) => !mockCountry(id) || (!mockCountry(id)?.status && !previous?.countryIds.includes(id)))) fields.countryIds = 'a country was not found or is deleted';

  for (const id of cityIds) {
    const city = mockCity(id);
    if (!city || (!city.status && !previous?.cityIds.includes(id))) fields.cityIds = 'a city was not found or is deleted';
    else if (!countryIds.includes(city.countryId)) fields.cityIds = 'every city must belong to one of the selected countries';
  }
  if (Object.keys(fields).length) throw validation(fields);

  if (refId !== null && db.tours.some((t) => t.refId === refId && t.id !== previous?.id)) throw new MockFail(409, ErrorCodes.RefIdTaken, 'Ref ID is already in use');

  return {
    refId,
    name,
    wholesaleId,
    wholesaleTourCode: str(body.wholesaleTourCode),
    state,
    days,
    nights,
    airlineId,
    highlight: str(body.highlight),
    description: str(body.description),
    bannerImage: str(body.bannerImage) || null,
    cardImage: str(body.cardImage) || null,
    pdfFile: str(body.pdfFile) || null,
    docxFile: str(body.docxFile) || null,
    internalNote: str(body.internalNote),
    countryIds,
    cityIds,
    images,
  };
}

const listTours: MockHandler = (req) => {
  authorize(req, 'POST /tour-programs/list', Permissions.TourProgramList);
  const body = asRecord(req.body);
  const keyword = str(body.keyword).toLowerCase();
  const page = Math.max(1, Number(body.page) || 1);
  const pageSize = Number(body.pageSize) || 20;
  if (pageSize > 100) throw validation({ pageSize: 'must be no greater than 100' });

  const rows = db.tours
    .filter((t) => !keyword || [t.name, t.code, t.wholesaleTourCode].some((v) => v.toLowerCase().includes(keyword)))
    .filter((t) => !body.countryId || t.countryIds.includes(String(body.countryId)))
    .filter((t) => !body.cityId || t.cityIds.includes(String(body.cityId)))
    .filter((t) => !body.wholesaleId || t.wholesaleId === body.wholesaleId)
    .filter((t) => !body.state || t.state === body.state)
    .filter((t) => typeof body.status !== 'boolean' || t.status === body.status)
    .sort((a, b) => b.updateDate.localeCompare(a.updateDate));

  return ok({
    items: rows.slice((page - 1) * pageSize, page * pageSize).map(toListItem),
    total: rows.length,
    page,
    pageSize,
  } satisfies ListResponse<TourProgramListItem>);
};

// --- Periods -------------------------------------------------------------------------------------

function parsePeriod(body: Record<string, unknown>) {
  const fields: Record<string, string> = {};
  const startDate = str(body.startDate);
  const endDate = str(body.endDate);
  const price = Number(body.price);
  const seatsTotal = Number(body.seatsTotal);
  const seatsAvailable = Number(body.seatsAvailable);
  const priceBeforeDiscount = optionalNumber(body.priceBeforeDiscount);

  if (!DATE_RE.test(startDate)) fields.startDate = 'must be a date in YYYY-MM-DD format';
  if (!DATE_RE.test(endDate)) fields.endDate = 'must be a date in YYYY-MM-DD format';
  else if (endDate < startDate) fields.endDate = 'must be on or after the start date';
  if (!Number.isInteger(price) || price < 0) fields.price = 'must be no less than 0';
  if (priceBeforeDiscount !== null && priceBeforeDiscount <= price) fields.priceBeforeDiscount = 'must be higher than the price';
  if (!Number.isInteger(seatsTotal) || seatsTotal < 0 || seatsTotal > 999) fields.seatsTotal = 'must be between 0 and 999';
  if (!Number.isInteger(seatsAvailable) || seatsAvailable < 0 || seatsAvailable > seatsTotal) fields.seatsAvailable = 'must not be more than total seats';
  if (Object.keys(fields).length) throw validation(fields);

  return {
    startDate,
    endDate,
    price,
    priceBeforeDiscount,
    childWithBedPrice: optionalNumber(body.childWithBedPrice),
    childNoBedPrice: optionalNumber(body.childNoBedPrice),
    singleSupplement: optionalNumber(body.singleSupplement),
    deposit: optionalNumber(body.deposit),
    seatsTotal,
    seatsAvailable,
    note: str(body.note),
  };
}

function statusOf(body: Record<string, unknown>): boolean {
  if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
  return body.status;
}

export const tourProgramRoutes: Record<string, MockHandler> = {
  'POST /tour-programs/list': listTours,

  'GET /tour-programs/:id': (req) => {
    authorize(req, 'GET /tour-programs/:id', Permissions.TourProgramGet);
    return ok(toTour(findTour(req.params.id)));
  },

  'POST /tour-programs': (req) => {
    authorize(req, 'POST /tour-programs', Permissions.TourProgramCreate);
    const now = new Date().toISOString();
    const tour: StoredTour = {
      id: crypto.randomUUID(),
      code: `${CODE_PREFIX}${String(db.nextCode++).padStart(6, '0')}`,
      ...parseTour(asRecord(req.body), null),
      status: true,
      createDate: now,
      updateDate: now,
    };
    db.tours.push(tour);
    save();
    return ok(toTour(tour));
  },

  'PUT /tour-programs': (req) => {
    authorize(req, 'PUT /tour-programs', Permissions.TourProgramUpdate);
    const body = asRecord(req.body);
    const tour = findTour(body.id);
    Object.assign(tour, parseTour(body, tour), { updateDate: new Date().toISOString() });
    save();
    return ok(toTour(tour));
  },

  'PUT /tour-programs/status': (req) => {
    authorize(req, 'PUT /tour-programs/status', Permissions.TourProgramStatus);
    const body = asRecord(req.body);
    const tour = findTour(body.id);
    tour.status = statusOf(body);
    tour.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },

  'POST /tour-program-periods/list': (req) => {
    authorize(req, 'POST /tour-program-periods/list', Permissions.PeriodList);
    const body = asRecord(req.body);
    if (!str(body.tourProgramId)) throw validation({ tourProgramId: 'cannot be blank' });
    const items = db.periods
      .filter((p) => p.tourProgramId === body.tourProgramId)
      .filter((p) => typeof body.status !== 'boolean' || p.status === body.status)
      .sort((a, b) => a.startDate.localeCompare(b.startDate) || a.endDate.localeCompare(b.endDate));
    return ok({ items });
  },

  'POST /tour-program-periods': (req) => {
    authorize(req, 'POST /tour-program-periods', Permissions.PeriodCreate);
    const body = asRecord(req.body);
    const tour = db.tours.find((t) => t.id === body.tourProgramId);
    if (!tour || !tour.status) throw validation({ tourProgramId: 'tour program not found or deleted' });
    const now = new Date().toISOString();
    const period: TourProgramPeriod = { id: crypto.randomUUID(), tourProgramId: tour.id, ...parsePeriod(body), status: true, createDate: now, updateDate: now };
    db.periods.push(period);
    save();
    return ok(period);
  },

  'PUT /tour-program-periods': (req) => {
    authorize(req, 'PUT /tour-program-periods', Permissions.PeriodUpdate);
    const body = asRecord(req.body);
    const period = findPeriod(body.id);
    Object.assign(period, parsePeriod(body), { updateDate: new Date().toISOString() });
    save();
    return ok(period);
  },

  'PUT /tour-program-periods/status': (req) => {
    authorize(req, 'PUT /tour-program-periods/status', Permissions.PeriodStatus);
    const body = asRecord(req.body);
    const period = findPeriod(body.id);
    period.status = statusOf(body);
    period.updateDate = new Date().toISOString();
    save();
    return ok(null);
  },
};
