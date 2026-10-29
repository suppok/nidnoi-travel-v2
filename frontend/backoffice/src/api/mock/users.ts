import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import { canGrantRole, isSuperAdmin, type Role, ROLE_INFO } from '../../constants/roles.ts';
import { passwordProblem } from '../../lib/password.ts';
import type { ListResponse } from '../../types/api.ts';
import type { User } from '../../types/user.ts';
import { type MockUser, save, toUser, users } from './db.ts';
import { asRecord, authorize, EMAIL_RE, type MockHandler, MockFail, ok, str, validation } from './http.ts';

const ROLE_VALUES = new Set<string>(ROLE_INFO.map((r) => r.value));

function findUser(id: unknown): MockUser {
  const user = users.find((u) => u.id === id);
  if (!user) throw new MockFail(404, ErrorCodes.NotFound, 'User not found');
  return user;
}

function parseProfile(body: Record<string, unknown>, selfId: string | null) {
  const email = str(body.email).toLowerCase();
  const name = str(body.name);
  const roles = Array.isArray(body.roles) ? [...new Set(body.roles)] : [];
  const fields: Record<string, string> = {};

  if (!EMAIL_RE.test(email)) fields.email = 'อีเมลไม่ถูกต้อง';
  if (!name) fields.name = 'กรุณากรอกชื่อ';
  else if (name.length > 100) fields.name = 'ชื่อยาวได้ไม่เกิน 100 ตัวอักษร';
  if (roles.length === 0) fields.roles = 'ต้องมีอย่างน้อย 1 บทบาท';
  else if (roles.some((r) => typeof r !== 'string' || !ROLE_VALUES.has(r))) fields.roles = 'บทบาทไม่ถูกต้อง';
  if (Object.keys(fields).length) throw validation(fields);

  if (users.some((u) => u.email === email && u.id !== selfId)) {
    throw new MockFail(409, ErrorCodes.EmailTaken, 'Email is already in use');
  }
  return { email, name, roles: roles as Role[] };
}

// Only the super admin may grant the admin role.
function ensureCanGrant(actor: MockUser, roles: readonly Role[]) {
  if (roles.some((r) => r !== 'super_admin' && !canGrantRole(actor.roles, r))) {
    throw new MockFail(403, ErrorCodes.SuperAdminRequired, 'Only the super admin can grant the admin role or manage admins');
  }
}

// Nobody else may change the super admin; only the super admin may change an admin.
function ensureCanManage(actor: MockUser, user: MockUser) {
  if (isSuperAdmin(user.roles)) throw new MockFail(409, ErrorCodes.SuperAdminProtected, 'The super admin can only be changed by themselves');
  ensureCanGrant(actor, user.roles);
}

const list: MockHandler = (req) => {
  authorize(req, 'POST /users/list', Permissions.UserList);
  const body = asRecord(req.body);
  const keyword = str(body.keyword).toLowerCase();
  const page = Math.max(1, Number(body.page) || 1);
  const pageSize = Math.min(100, Math.max(1, Number(body.pageSize) || 20));

  const filtered = users
    .filter((u) => !keyword || u.name.toLowerCase().includes(keyword) || u.email.includes(keyword))
    .filter((u) => !body.role || u.roles.includes(body.role as Role))
    .filter((u) => typeof body.status !== 'boolean' || u.status === body.status)
    .sort((a, b) => a.name.localeCompare(b.name, 'th'));

  return ok({
    items: filtered.slice((page - 1) * pageSize, page * pageSize).map(toUser),
    total: filtered.length,
    page,
    pageSize,
  } satisfies ListResponse<User>);
};

const get: MockHandler = (req) => {
  authorize(req, 'GET /users/:id', Permissions.UserGet);
  return ok(toUser(findUser(req.params.id)));
};

const create: MockHandler = (req) => {
  const actor = authorize(req, 'POST /users', Permissions.UserCreate);
  const body = asRecord(req.body);
  const profile = parseProfile(body, null);
  if (isSuperAdmin(profile.roles)) throw validation({ roles: 'super_admin cannot be assigned' });
  ensureCanGrant(actor, profile.roles);
  const temporaryPassword = typeof body.temporaryPassword === 'string' ? body.temporaryPassword : '';
  const problem = passwordProblem(temporaryPassword);
  if (problem) throw validation({ temporaryPassword: problem });

  const now = new Date().toISOString();
  const user: MockUser = {
    id: crypto.randomUUID(),
    ...profile,
    status: true,
    mustChangePassword: true,
    lastLoginAt: null,
    createDate: now,
    updateDate: now,
    password: temporaryPassword,
    sessionVersion: 1,
  };
  users.push(user);
  save();
  return ok(toUser(user), 201);
};

const update: MockHandler = (req) => {
  const actor = authorize(req, 'PUT /users', Permissions.UserUpdate);
  const body = asRecord(req.body);
  const user = findUser(body.id);
  const profile = parseProfile(body, user.id);
  if (user.id !== actor.id) {
    ensureCanManage(actor, user);
    ensureCanGrant(actor, profile.roles);
  }

  const rolesChanged = profile.roles.length !== user.roles.length || profile.roles.some((r) => !user.roles.includes(r));
  if (rolesChanged && user.id === actor.id) {
    throw new MockFail(409, ErrorCodes.CannotModifySelf, 'You cannot change your own roles');
  }
  if (rolesChanged && isSuperAdmin(profile.roles)) throw validation({ roles: 'super_admin cannot be assigned' });

  Object.assign(user, profile, { updateDate: new Date().toISOString() });
  if (rolesChanged) user.sessionVersion += 1;
  save();
  return ok(toUser(user));
};

const updateStatus: MockHandler = (req) => {
  const actor = authorize(req, 'PUT /users/status', Permissions.UserStatus);
  const body = asRecord(req.body);
  if (typeof body.status !== 'boolean') throw validation({ status: 'required' });
  const user = findUser(body.id);

  if (user.id === actor.id) throw new MockFail(409, ErrorCodes.CannotModifySelf, 'You cannot change your own status');
  ensureCanManage(actor, user);

  user.status = body.status;
  user.updateDate = new Date().toISOString();
  if (!body.status) user.sessionVersion += 1;
  save();
  return ok(null);
};

const resetPassword: MockHandler = (req) => {
  const actor = authorize(req, 'PUT /users/password', Permissions.UserPasswordReset);
  const body = asRecord(req.body);
  const user = findUser(body.id);
  if (user.id === actor.id) throw new MockFail(409, ErrorCodes.CannotModifySelf, 'Use PUT /auth/password for your own password');
  ensureCanManage(actor, user);
  const temporaryPassword = typeof body.temporaryPassword === 'string' ? body.temporaryPassword : '';
  const problem = passwordProblem(temporaryPassword);
  if (problem) throw validation({ temporaryPassword: problem });

  user.password = temporaryPassword;
  user.mustChangePassword = true;
  user.sessionVersion += 1; // signs the user out everywhere
  user.updateDate = new Date().toISOString();
  save();
  return ok(null);
};

export const userRoutes: Record<string, MockHandler> = {
  'POST /users/list': list,
  'GET /users/:id': get,
  'POST /users': create,
  'PUT /users': update,
  'PUT /users/status': updateStatus,
  'PUT /users/password': resetPassword,
};
