// Backend error codes the UI branches on. A shipped code is never renamed.
export const ErrorCodes = {
  Validation: 'VALIDATION',
  Unauthenticated: 'UNAUTHENTICATED',
  PermissionDenied: 'PERMISSION_DENIED',
  NotFound: 'NOT_FOUND',
  Conflict: 'CONFLICT',
  RateLimited: 'RATE_LIMITED',
  InvalidCredentials: 'INVALID_CREDENTIALS',
  InvalidPassword: 'INVALID_PASSWORD',
  PasswordChangeRequired: 'PASSWORD_CHANGE_REQUIRED',
  EmailTaken: 'EMAIL_TAKEN',
  CannotModifySelf: 'CANNOT_MODIFY_SELF',
  SuperAdminProtected: 'SUPER_ADMIN_PROTECTED',
  SuperAdminRequired: 'SUPER_ADMIN_REQUIRED',
  SlugTaken: 'SLUG_TAKEN',
  RefIdTaken: 'REF_ID_TAKEN',
  CodeTaken: 'CODE_TAKEN',
  NameTaken: 'NAME_TAKEN',
  MediaTypeNotAllowed: 'MEDIA_TYPE_NOT_ALLOWED',
  MediaTooLarge: 'MEDIA_TOO_LARGE',
  ImageTooSmall: 'IMAGE_TOO_SMALL',
  StorageUnavailable: 'STORAGE_UNAVAILABLE',
  SiteRebuildUnavailable: 'SITE_REBUILD_UNAVAILABLE',
  SiteRebuildFailed: 'SITE_REBUILD_FAILED',
  LineNotConfigured: 'LINE_NOT_CONFIGURED',
  PeriodUnavailable: 'PERIOD_UNAVAILABLE', // customer site only (POST /bookings/public)
  // Client-side only: the request never reached the backend.
  Network: 'NETWORK',
  Unknown: 'UNKNOWN',
} as const;

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes];
