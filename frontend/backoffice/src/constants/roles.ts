// Fixed roles. MUST mirror the role names in backend/config/permissions/policy.csv.
// Every user has one or more roles; permissions are the union of their roles' policies.
export const Roles = {
  // The first user, created by the backend bootstrap command. Never assignable; only they can edit themselves.
  SuperAdmin: 'super_admin',
  Admin: 'admin',
  Sales: 'sales',
  Finance: 'finance',
} as const;

export type Role = (typeof Roles)[keyof typeof Roles];

type RoleInfo = {
  value: Role;
  label: string;
  description: string;
  color: string;
};

// sales/finance descriptions follow the placeholder policy rows — confirm with the business before prd.
export const ROLE_INFO: readonly RoleInfo[] = [
  {
    value: Roles.SuperAdmin,
    label: 'ผู้ดูแลระบบสูงสุด',
    description: 'ผู้ใช้คนแรกของระบบ มีสิทธิ์ทุกอย่าง และแก้ไขได้โดยตัวเองเท่านั้น',
    color: 'purple',
  },
  {
    value: Roles.Admin,
    label: 'ผู้ดูแลระบบ',
    description: 'เข้าถึงได้ทุกเมนู รวมถึงจัดการผู้ใช้และสิทธิ์',
    color: 'magenta',
  },
  {
    value: Roles.Sales,
    label: 'ฝ่ายขาย',
    description: 'สร้างและแก้ไขทัวร์ รอบเดินทาง ราคา ที่นั่ง และอัปโหลดรูป',
    color: 'blue',
  },
  {
    value: Roles.Finance,
    label: 'ฝ่ายบัญชี',
    description: 'ดูข้อมูลทัวร์',
    color: 'gold',
  },
];

// Roles staff can grant in the user form (super_admin is never assignable).
export const ASSIGNABLE_ROLES = ROLE_INFO.filter((r) => r.value !== Roles.SuperAdmin);

export function isSuperAdmin(roles: readonly Role[]): boolean {
  return roles.includes(Roles.SuperAdmin);
}

// Mirrors the backend user rules: nobody else may change the super admin, and only the super admin
// may grant the admin role or change an admin.
export function canGrantRole(actorRoles: readonly Role[], role: Role): boolean {
  if (role === Roles.SuperAdmin) return false;
  return role !== Roles.Admin || isSuperAdmin(actorRoles);
}

export function canManageUser(actorRoles: readonly Role[], targetRoles: readonly Role[]): boolean {
  return !isSuperAdmin(targetRoles) && targetRoles.every((r) => canGrantRole(actorRoles, r));
}

export function roleInfo(role: Role): RoleInfo {
  return ROLE_INFO.find((r) => r.value === role) ?? { value: role, label: role, description: '', color: 'default' };
}
