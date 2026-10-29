import { Roles } from '../../constants/roles.ts';
import type { User } from '../../types/user.ts';
import { MOCK_PASSWORD } from './credentials.ts';

// Mock-only persistence. The real backend owns users in Postgres; nothing here is a token store
// for the real app.
export type MockUser = User & {
  password: string;
  // Bumped on password reset, deactivation and role change to revoke refresh "cookies".
  sessionVersion: number;
};

const DB_KEY = 'nidnoi.mock.users.v3';

const now = '2026-10-01T09:00:00.000Z';

function seed(): MockUser[] {
  const base = { status: true, mustChangePassword: false, lastLoginAt: null, createDate: now, updateDate: now, sessionVersion: 1 };
  return [
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000000', email: 'owner@nidnoi.test', name: 'เจ้าของระบบ', roles: [Roles.SuperAdmin], password: MOCK_PASSWORD },
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000001', email: 'admin@nidnoi.test', name: 'แอดมิน', roles: [Roles.Admin], password: MOCK_PASSWORD },
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000002', email: 'sales@nidnoi.test', name: 'สมใจ ขายเก่ง', roles: [Roles.Sales], password: MOCK_PASSWORD },
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000003', email: 'finance@nidnoi.test', name: 'มานี บัญชีดี', roles: [Roles.Sales, Roles.Finance], password: MOCK_PASSWORD },
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000004', email: 'new@nidnoi.test', name: 'พนักงานใหม่', roles: [Roles.Sales], password: MOCK_PASSWORD, mustChangePassword: true },
    { ...base, id: '0b9c6a52-0000-4000-8000-000000000005', email: 'old@nidnoi.test', name: 'อดีตพนักงาน', roles: [Roles.Sales], password: MOCK_PASSWORD, status: false },
  ];
}

function load(): MockUser[] {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as MockUser[];
  } catch {
    // fall through to seed
  }
  return seed();
}

export const users: MockUser[] = load();

export function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(users));
  } catch {
    // storage unavailable: keep in memory only
  }
}

export function toUser({ password: _password, sessionVersion: _sessionVersion, ...user }: MockUser): User {
  return user;
}
