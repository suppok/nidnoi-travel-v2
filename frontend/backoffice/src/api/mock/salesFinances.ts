import { Permissions } from '../../constants/permissions.ts';
import { Roles } from '../../constants/roles.ts';
import type { SalesFinanceSettings } from '../../types/salesFinance.ts';
import { users } from './db.ts';
import { asRecord, authorize, type MockHandler, ok, str, validation } from './http.ts';

// Mock-only persistence: sales user id → their finance user ids and default.
const DB_KEY = 'nidnoi.mock.sales-finances.v1';
type Links = Record<string, { ids: string[]; defaultId: string }>;

function load(): Links {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Links;
  } catch {
    // fall through to empty
  }
  return {};
}

const links: Links = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(links));
  } catch {
    // storage unavailable: keep in memory
  }
}

const isFinance = (id: string) => users.some((u) => u.id === id && u.status && u.roles.includes(Roles.Finance));

function settings(me: string): SalesFinanceSettings {
  const options = users
    .filter((u) => u.status && u.roles.includes(Roles.Finance) && u.id !== me)
    .sort((a, b) => a.name.localeCompare(b.name, 'th'))
    .map((u) => ({ id: u.id, name: u.name }));
  const mine = links[me] ?? { ids: [], defaultId: '' };
  const ids = options.filter((o) => mine.ids.includes(o.id)).map((o) => o.id);
  return { options, financeUserIds: ids, defaultFinanceUserId: ids.includes(mine.defaultId) ? mine.defaultId : '' };
}

export const salesFinanceRoutes: Record<string, MockHandler> = {
  'GET /sales-finances': (req) => ok(settings(authorize(req, 'GET /sales-finances', Permissions.SalesFinanceGet).id)),
  'PUT /sales-finances': (req) => {
    const actor = authorize(req, 'PUT /sales-finances', Permissions.SalesFinanceUpdate);
    const body = asRecord(req.body);
    const ids = Array.isArray(body.financeUserIds) ? [...new Set(body.financeUserIds.map(String))] : [];
    let defaultId = str(body.defaultFinanceUserId);
    if (ids.length > 20) throw validation({ financeUserIds: 'เลือกได้ไม่เกิน 20 คน' });
    if (ids.some((id) => id === actor.id || !isFinance(id))) throw validation({ financeUserIds: 'must be active finance users other than yourself' });
    if (defaultId && !ids.includes(defaultId)) throw validation({ defaultFinanceUserId: 'must be one of financeUserIds' });
    if (!defaultId && ids.length === 1) defaultId = ids[0];
    links[actor.id] = { ids, defaultId };
    save();
    return ok(settings(actor.id));
  },
};
