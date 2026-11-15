import { LineBookingEvents } from '../../constants/lineBookingEvents.ts';
import { LineBookingStates, type LineBookingState } from '../../constants/lineBookingStates.ts';
import { Permissions } from '../../constants/permissions.ts';
import { Roles } from '../../constants/roles.ts';
import { ErrorCodes } from '../../constants/errorCodes.ts';
import type { ListResponse } from '../../types/api.ts';
import type {
  LineBooking,
  LineBookingCreateRequest,
  LineBookingDeal,
  LineBookingDealUpdateRequest,
  LineBookingEvent,
  LineBookingEventChange,
} from '../../types/lineBooking.ts';
import { users } from './db.ts';
import { asRecord, authorize, type MockHandler, MockFail, ok, str, validation } from './http.ts';

// Mock-only LINE bookings: list, get and history, plus the booking entry page (create, slip upload,
// reopen by booking no or customer chat, edit / balance). Four bookings at different stages, each
// with its history, so the list tags, the drawer and the "ประวัติ" timeline can be tried. Staff
// actions (confirm, documents, …) are not mocked. New bookings live for this page load only.

const sales = { id: users[2].id, name: users[2].name }; // สมใจ ขายเก่ง
const finance = { id: users[3].id, name: users[3].name }; // มานี บัญชีดี
const PDF = (name: string) => `https://example.com/mock/${name}.pdf`;
const SLIP = (text: string) => `https://placehold.co/400x600?text=${encodeURIComponent(text)}`;

const ago = (hours: number) => new Date(Date.now() - hours * 3_600_000).toISOString();

function base(id: number, bookingNo: string, hoursAgo: number, state: LineBookingState): LineBooking {
  return {
    id: `1b000000-0000-4000-8000-00000000000${id}`,
    bookingNo,
    lineUserId: 'U-mock',
    lineDisplayName: 'Somjai',
    createdBy: sales,
    tourProgramId: '00000000-0000-4000-8000-000000000001',
    tourProgramPeriodId: '00000000-0000-4000-8000-000000000002',
    tourCode: 'NIDN000001',
    tourName: 'ทัวร์มาเลเซีย Munz Malaysia มาเลเซีย กัวลาลัมเปอร์ เก็นติ้ง สวนสนุก SKYWORLD',
    startDate: '2026-12-30',
    endDate: '2027-01-01',
    adultTwinCount: 2,
    adultTwinPrice: 9499,
    adultSingleCount: 0,
    adultSinglePrice: null,
    childWithBedCount: 0,
    childWithBedPrice: null,
    childNoBedCount: 0,
    childNoBedPrice: null,
    totalPrice: 18998,
    customerName: 'สมชาย ใจดี',
    customerPhone: '0812345678',
    slipImage: SLIP('deposit slip'),
    slipImages: [SLIP('deposit slip')],
    note: '',
    state,
    staffNote: '',
    status: true,
    createDate: ago(hoursAgo),
    updateDate: ago(hoursAgo),
    infantCount: 0,
    infantPrice: null,
    adultPriceBeforeDiscount: null,
    discountPerPerson: 0,
    charges: [],
    paymentMethod: 'transfer',
    creditFeePercent: 0,
    cardFee: 0,
    isDeposit: true,
    depositPerPerson: 3000,
    depositDueDate: '2026-11-15',
    depositDueTime: '18:00',
    remainingDueDate: '2026-11-30',
    fullDueDate: null,
    fullDueTime: '',
    customerAddress: '',
    bookingSource: null,
    refName: '',
    refBookingNo: '',
    salesUser: sales,
    financeUser: state === LineBookingStates.New ? null : finance,
    confirmedAt: state === LineBookingStates.New ? null : ago(hoursAgo - 1),
    aiFile: '',
    ivFile: '',
    reFile: '',
    documentsAt: null,
    documentsApprovedAt: null,
    customerSentAt: null,
    fixNote: '',
    fixRequestedAt: null,
    balanceSlipImage: '',
    balanceSlipImages: [],
    balanceSubmittedAt: null,
    balanceConfirmedAt: null,
    lineContact: { id: 'c1000000-0000-4000-8000-000000000001', lineUserId: 'U1', displayName: 'Somchai', pictureUrl: null, following: true, chatUrl: '' },
  };
}

const event = (
  id: string,
  action: string,
  hoursAgo: number,
  user: { id: string; name: string },
  extra: { target?: string; note?: string; changes?: LineBookingEventChange[]; mentions?: { id: string; name: string }[] } = {},
): LineBookingEvent => ({
  id,
  action,
  changes: extra.changes ?? [],
  target: extra.target ?? '',
  note: extra.note ?? '',
  mentions: extra.mentions ?? [],
  createDate: ago(hoursAgo),
  user,
});

const E = LineBookingEvents;

// 1. New: waits for the owner's confirm.
const newBooking = base(1, '26NEWMCK', 2, LineBookingStates.New);

// 2. Deposit paid, finance sent AI + IV back after a fix request; waits for the owner's review.
const reviewBooking: LineBooking = {
  ...base(2, '26REVMCK', 30, LineBookingStates.DepositPaid),
  aiFile: PDF('AI-26REVMCK'),
  ivFile: PDF('IV-26REVMCK'),
  documentsAt: ago(3),
};

// 3. Deposit documents sent to the customer, the deal edited, then the balance slip recorded.
const balanceBooking: LineBooking = {
  ...base(3, '26BALMCK', 72, LineBookingStates.DepositPaid),
  adultTwinCount: 3,
  totalPrice: 28497,
  aiFile: PDF('AI-26BALMCK'),
  ivFile: PDF('IV-26BALMCK'),
  documentsAt: ago(60),
  documentsApprovedAt: ago(59),
  customerSentAt: ago(59),
  balanceSlipImage: SLIP('balance slip 1'),
  balanceSlipImages: [SLIP('balance slip 1'), SLIP('balance slip 2')],
  balanceSubmittedAt: ago(1),
  updateDate: ago(1),
};

// 4. Paid in full after the balance; the RE was sent to the customer.
const paidBooking: LineBooking = {
  ...base(4, '26PAIDMK', 200, LineBookingStates.PaidInFull),
  aiFile: PDF('AI-26PAIDMK'),
  ivFile: PDF('IV-26PAIDMK'),
  reFile: PDF('RE-26PAIDMK'),
  documentsAt: ago(20),
  documentsApprovedAt: ago(19),
  customerSentAt: ago(19),
  balanceSlipImage: SLIP('balance slip 1'),
  balanceSlipImages: [SLIP('balance slip 1'), SLIP('balance slip 2')],
  balanceSubmittedAt: ago(48),
  balanceConfirmedAt: ago(47),
  updateDate: ago(19),
};

const bookings: LineBooking[] = [newBooking, reviewBooking, balanceBooking, paidBooking];

const history: Record<string, LineBookingEvent[]> = {
  [newBooking.id]: [event('n1', E.Created, 2, sales)],
  [reviewBooking.id]: [
    event('r1', E.Created, 30, sales),
    event('r2', E.Confirmed, 29, sales, { target: finance.name }),
    event('r3', E.DocumentsSent, 20, finance, { target: 'AI + IV' }),
    event('r4', E.DocumentsRejected, 10, sales, { target: 'AI + IV', note: 'ชื่อลูกค้าในใบรับเงินมัดจำสะกดผิด' }),
    event('r5', E.DocumentsSent, 3, finance, { target: 'AI + IV' }),
  ],
  [balanceBooking.id]: [
    event('b1', E.Created, 72, sales),
    event('b2', E.Confirmed, 71, sales, { target: finance.name }),
    event('b3', E.DocumentsSent, 60, finance, { target: 'AI + IV' }),
    event('b4', E.DocumentsApproved, 59, sales, { target: 'Somchai', note: 'AI + IV' }),
    event('b5', E.Edited, 24, sales, {
      changes: [
        { field: 'adultTwin', from: '2', to: '3' },
        { field: 'totalPrice', from: '18998', to: '28497' },
      ],
    }),
    event('b6', E.StaffNoteUpdated, 5, sales, { changes: [{ field: 'staffNote', from: '', to: 'ลูกค้าขอเพิ่ม 1 ท่าน' }] }),
    event('b6c', E.Comment, 4, finance, { note: `@${sales.name} ลูกค้าโอนส่วนที่เหลือมาแล้ว รบกวนแนบสลิปด้วยค่ะ`, mentions: [sales] }),
    event('b7', E.BalanceSubmitted, 1, sales),
  ],
  [paidBooking.id]: [
    event('p1', E.Created, 200, sales),
    event('p2', E.Confirmed, 199, sales, { target: finance.name }),
    event('p3', E.DocumentsSent, 150, finance, { target: 'AI + IV' }),
    event('p4', E.DocumentsApproved, 149, sales, { target: 'Somchai', note: 'AI + IV' }),
    event('p5', E.BalanceSubmitted, 48, sales),
    event('p6', E.BalanceConfirmed, 47, sales, { target: finance.name }),
    event('p7', E.DocumentsSent, 20, finance, { target: 'RE' }),
    event('p8', E.DocumentsApproved, 19, sales, { target: 'Somchai', note: 'RE' }),
  ],
};

const PAID: number[] = [LineBookingStates.DepositPaid, LineBookingStates.PaidInFull];

// Owner scope like the backend: admins see everything; sales their own, finance the ones sent to them.
function visible(b: LineBooking, user: (typeof users)[number]): boolean {
  if (user.roles.some((r) => r === Roles.Admin || r === Roles.SuperAdmin)) return true;
  return (user.roles.includes(Roles.Sales) && b.salesUser?.id === user.id) || (user.roles.includes(Roles.Finance) && b.financeUser?.id === user.id);
}

function find(req: Parameters<MockHandler>[0], route: string) {
  const user = authorize(req, route, Permissions.LineBookingGet);
  const b = bookings.find((x) => x.id === req.params.id);
  if (!b || !visible(b, user)) throw new MockFail(404, ErrorCodes.NotFound, 'Not found');
  return b;
}

// Like the backend's quoteLine: rows - discount + charges + card fee (on the deposit when paying one).
function mockTotal(d: LineBookingDeal): number {
  const pax = d.adultTwin + d.adultSingle + d.childWithBed + d.childNoBed + d.infant;
  const adult = d.adultPrice ?? 0;
  const rows =
    d.adultTwin * adult +
    d.adultSingle * (adult + (d.singleSupplement ?? 0)) +
    d.childWithBed * (d.childWithBedPrice ?? 0) +
    d.childNoBed * (d.childNoBedPrice ?? 0) +
    d.infant * (d.infantPrice ?? 0) -
    d.discountPerPerson * pax +
    d.charges.reduce((sum, c) => sum + c.count * c.price, 0);
  const feeBase = d.isDeposit ? d.depositPerPerson * pax : rows;
  const fee = d.paymentMethod === 'credit' ? Math.ceil((feeBase * Math.round(d.creditFeePercent * 100)) / 10000) : 0;
  return rows + fee;
}

function dealFields(d: LineBookingDeal, total: number): Partial<LineBooking> {
  return {
    adultTwinCount: d.adultTwin,
    adultTwinPrice: d.adultPrice ?? 0,
    adultSingleCount: d.adultSingle,
    adultSinglePrice: d.adultSingle > 0 ? (d.adultPrice ?? 0) + (d.singleSupplement ?? 0) : null,
    childWithBedCount: d.childWithBed,
    childWithBedPrice: d.childWithBedPrice,
    childNoBedCount: d.childNoBed,
    childNoBedPrice: d.childNoBedPrice,
    infantCount: d.infant,
    infantPrice: d.infantPrice,
    adultPriceBeforeDiscount: d.adultPriceBeforeDiscount,
    discountPerPerson: d.discountPerPerson,
    charges: d.charges,
    paymentMethod: d.paymentMethod,
    creditFeePercent: d.creditFeePercent,
    isDeposit: d.isDeposit,
    depositPerPerson: d.depositPerPerson,
    depositDueDate: d.depositDueDate,
    depositDueTime: d.depositDueTime,
    remainingDueDate: d.remainingDueDate,
    fullDueDate: d.fullDueDate,
    fullDueTime: d.fullDueTime,
    totalPrice: total,
  };
}

// Reopen rule like the backend: an active, paid booking; any user with line-booking:create.
function findEditable(req: Parameters<MockHandler>[0], route: string, bookingNo: string): LineBooking {
  authorize(req, route, Permissions.LineBookingCreate);
  const no = bookingNo.trim().toUpperCase();
  const b = bookings.find((x) => x.bookingNo === no && x.status);
  if (!b) throw new MockFail(404, ErrorCodes.NotFound, 'Not found');
  if (!PAID.includes(b.state)) throw new MockFail(409, ErrorCodes.LineBookingNotEditable, 'Only a paid booking can be reopened');
  return b;
}

const SOMCHAI = 'c1000000-0000-4000-8000-000000000001';
let somchaiChatUrl = '';
const contactOf = (id: string | null, chatUrl = '') => {
  if (id !== SOMCHAI) return null;
  if (chatUrl) somchaiChatUrl = chatUrl;
  return { id, lineUserId: 'U1', displayName: 'Somchai', pictureUrl: null, following: true, chatUrl: somchaiChatUrl };
};

export const lineBookingRoutes: Record<string, MockHandler> = {
  'POST /line-bookings': (req) => {
    const user = authorize(req, 'POST /line-bookings', Permissions.LineBookingCreate);
    const body = req.body as LineBookingCreateRequest;
    if (!body.slipImages?.length) throw validation({ slipImages: 'at least one slip is required' });
    const owner = users.find((u) => u.id === body.salesUserId);
    if (!owner || !owner.roles.includes(Roles.Sales)) throw validation({ salesUserId: 'must be an active sales user' });
    const total = mockTotal(body);
    const no = `26${Math.random().toString(36).slice(2, 8).toUpperCase()}`;
    const b: LineBooking = {
      ...base(9, no, 0, LineBookingStates.New),
      ...dealFields(body, total),
      id: crypto.randomUUID(),
      lineDisplayName: user.name,
      createdBy: { id: user.id, name: user.name },
      tourCode: body.tourCode,
      tourName: `โปรแกรม ${body.tourCode} (mock)`,
      customerName: body.customerName,
      customerPhone: body.customerPhone,
      customerAddress: body.customerAddress,
      refName: body.refName,
      refBookingNo: body.refBookingNo,
      note: body.note,
      slipImage: body.slipImages[0],
      slipImages: body.slipImages,
      salesUser: { id: owner.id, name: owner.name },
      lineContact: contactOf(body.lineContactId, body.chatUrl),
    };
    bookings.unshift(b);
    history[b.id] = [event(crypto.randomUUID(), E.Created, 0, { id: user.id, name: user.name })];
    console.info('[mock] POST /line-bookings', body);
    return ok({ bookingNo: no, totalPrice: total });
  },
  'POST /media/upload/line-booking-slip': (req) => {
    authorize(req, 'POST /media/upload/line-booking-slip', Permissions.LineBookingCreate);
    const file = req.body instanceof FormData ? req.body.get('file') : null;
    if (!(file instanceof File) || file.size === 0) throw validation({ file: 'file is required' });
    return ok({ url: URL.createObjectURL(file), key: `bookings/slips/${crypto.randomUUID()}`, size: file.size, mime: file.type, fileName: '', width: 0, height: 0 });
  },
  'POST /line-bookings/editable/list': (req) => {
    authorize(req, 'POST /line-bookings/editable/list', Permissions.LineBookingCreate);
    const id = str(asRecord(req.body).lineContactId);
    return ok({ items: bookings.filter((b) => b.status && PAID.includes(b.state) && b.lineContact?.id === id) });
  },
  'GET /line-bookings/editable/:bookingNo/events': (req) => {
    const b = findEditable(req, 'GET /line-bookings/editable/:bookingNo/events', req.params.bookingNo ?? '');
    return ok({ items: history[b.id] ?? [] });
  },
  'GET /line-bookings/editable/:bookingNo': (req) => ok(findEditable(req, 'GET /line-bookings/editable/:bookingNo', req.params.bookingNo ?? '')),
  'PUT /line-bookings/editable': (req) => {
    const body = req.body as LineBookingDealUpdateRequest;
    const b = findEditable(req, 'PUT /line-bookings/editable', body.bookingNo ?? '');
    const balance = (body.balanceSlipImages ?? []).length > 0;
    if (balance && !(b.isDeposit && b.state === LineBookingStates.DepositPaid && b.documentsApprovedAt)) {
      throw validation({ balanceSlipImages: 'this booking does not take a balance' });
    }
    const deal = { ...body, isDeposit: b.isDeposit };
    const total = mockTotal(deal);
    Object.assign(b, dealFields(deal, total), {
      customerName: body.customerName,
      customerPhone: body.customerPhone,
      customerAddress: body.customerAddress,
      note: body.note,
      lineContact: contactOf(body.lineContactId, body.chatUrl) ?? b.lineContact,
      updateDate: new Date().toISOString(),
      ...(balance ? { balanceSlipImages: body.balanceSlipImages, balanceSlipImage: body.balanceSlipImages[0], balanceSubmittedAt: new Date().toISOString() } : {}),
    });
    console.info('[mock] PUT /line-bookings/editable', body);
    return ok({ bookingNo: b.bookingNo, totalPrice: total, balancePending: balance });
  },
  'POST /line-bookings/list': (req) => {
    const user = authorize(req, 'POST /line-bookings/list', Permissions.LineBookingList);
    const body = asRecord(req.body);
    const keyword = str(body.keyword).trim().toLowerCase();
    const flag = (name: string) => (typeof body[name] === 'boolean' ? (body[name] as boolean) : undefined);
    const page = Number(body.page) || 1;
    const pageSize = Number(body.pageSize) || 20;
    const items = bookings.filter((b) => {
      if (!visible(b, user)) return false;
      if (keyword && ![b.bookingNo, b.customerName, b.customerPhone, b.tourCode].some((v) => v.toLowerCase().includes(keyword))) return false;
      if (body.state && b.state !== body.state) return false;
      if (body.paid && !PAID.includes(b.state)) return false;
      if (flag('status') !== undefined && b.status !== flag('status')) return false;
      if (body.ownedByMe && b.salesUser?.id !== user.id) return false;
      if (body.assignedToMe && b.financeUser?.id !== user.id) return false;
      if (flag('documentsDone') !== undefined && (b.documentsAt !== null) !== flag('documentsDone')) return false;
      if (flag('documentsApproved') !== undefined && (b.documentsApprovedAt !== null) !== flag('documentsApproved')) return false;
      const pending = b.state === LineBookingStates.DepositPaid && !!b.balanceSubmittedAt && !b.balanceConfirmedAt;
      if (flag('balancePending') !== undefined && pending !== flag('balancePending')) return false;
      return true;
    });
    const res: ListResponse<LineBooking> = { items: items.slice((page - 1) * pageSize, page * pageSize), total: items.length, page, pageSize };
    return ok(res);
  },
  'GET /line-bookings/:id/events': (req) => ok({ items: history[find(req, 'GET /line-bookings/:id/events').id] ?? [] }),
  'GET /line-bookings/:id/finance-users': (req) => {
    find(req, 'GET /line-bookings/:id/finance-users');
    return ok({ items: [finance], defaultId: finance.id });
  },
  'GET /line-bookings/:id': (req) => ok(find(req, 'GET /line-bookings/:id')),
  'GET /line-bookings/:id/mentionable': (req) => {
    find(req, 'GET /line-bookings/:id/mentionable');
    return ok({ items: [finance, sales, { id: users[1].id, name: users[1].name }] });
  },
  // Comments are kept in memory for this page load; notifications are not mocked (always none).
  'POST /line-bookings/comments': (req) => {
    const user = authorize(req, 'POST /line-bookings/comments', Permissions.LineBookingGet);
    const body = asRecord(req.body);
    const b = bookings.find((x) => x.id === str(body.id));
    if (!b || !visible(b, user)) throw new MockFail(404, ErrorCodes.NotFound, 'Not found');
    const ids = Array.isArray(body.mentionUserIds) ? (body.mentionUserIds as string[]) : [];
    const mentions = users.filter((u) => ids.includes(u.id)).map((u) => ({ id: u.id, name: u.name }));
    const e: LineBookingEvent = { ...event(crypto.randomUUID(), E.Comment, 0, { id: user.id, name: user.name }, { note: str(body.text), mentions }) };
    (history[b.id] ??= []).push(e);
    return ok(e);
  },
  'POST /line-booking-notifications/list': (req) => {
    authorize(req, 'POST /line-booking-notifications/list', Permissions.LineBookingList);
    return ok({ items: [], total: 0, page: 1, pageSize: 20 });
  },
  'PUT /line-booking-notifications/read': (req) => {
    authorize(req, 'PUT /line-booking-notifications/read', Permissions.LineBookingList);
    return ok(null);
  },
  'POST /line-contacts/list': (req) => {
    authorize(req, 'POST /line-contacts/list', Permissions.LineContactList);
    const keyword = str(asRecord(req.body).keyword).trim().toLowerCase();
    const contacts = [{ id: 'c1000000-0000-4000-8000-000000000001', lineUserId: 'U1', displayName: 'Somchai', pictureUrl: null, following: true, lastMessageAt: ago(1), chatUrl: somchaiChatUrl }];
    const items = contacts.filter((c) => !keyword || c.displayName.toLowerCase() === keyword);
    return ok({ items, total: items.length, page: 1, pageSize: 20 });
  },
};
