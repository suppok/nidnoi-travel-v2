// MUST mirror backend/pkg/constants/apicode/. Update both in the same PR.
export const Permissions = {
  TourProgramCreate: 'tour-program:create',
  TourProgramUpdate: 'tour-program:update',
  TourProgramList: 'tour-program:list',
  TourProgramGet: 'tour-program:get',
  TourProgramStatus: 'tour-program:status',
  PeriodUpdate: 'tour-program-period:update',
  CountryCreate: 'country:create',
  CountryUpdate: 'country:update',
  CountryStatus: 'country:status',
  CityCreate: 'city:create',
  CityUpdate: 'city:update',
  CityStatus: 'city:status',
  MediaUpload: 'media:upload',
  SiteRebuild: 'site:rebuild',
  UserCreate: 'user:create',
  UserUpdate: 'user:update',
  UserList: 'user:list',
  UserGet: 'user:get',
  UserStatus: 'user:status',
  UserPasswordReset: 'user:password-reset',
} as const;

export type Permission = (typeof Permissions)[keyof typeof Permissions];
