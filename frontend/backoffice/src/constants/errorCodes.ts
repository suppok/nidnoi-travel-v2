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
  // Client-side only: the request never reached the backend.
  Network: 'NETWORK',
  Unknown: 'UNKNOWN',
} as const;

export type ErrorCode = (typeof ErrorCodes)[keyof typeof ErrorCodes];
