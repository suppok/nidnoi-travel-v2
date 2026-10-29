import { type Permission, Permissions } from '../../constants/permissions.ts';
import { type Role, Roles } from '../../constants/roles.ts';

// Mirrors backend/config/permissions/policy.csv (sales/finance rows are placeholders there too).
const POLICY: Record<Role, readonly Permission[]> = {
  [Roles.SuperAdmin]: Object.values(Permissions),
  [Roles.Admin]: Object.values(Permissions),
  [Roles.Sales]: [
    Permissions.TourProgramCreate,
    Permissions.TourProgramUpdate,
    Permissions.TourProgramList,
    Permissions.TourProgramGet,
    Permissions.PeriodUpdate,
    Permissions.MediaUpload,
  ],
  [Roles.Finance]: [Permissions.TourProgramList, Permissions.TourProgramGet],
};

// A user is allowed an action when any of their roles allows it.
export function permissionsFor(roles: readonly Role[]): Permission[] {
  return [...new Set(roles.flatMap((r) => POLICY[r] ?? []))];
}
