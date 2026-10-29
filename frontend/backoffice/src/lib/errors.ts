import type { FormInstance } from 'antd';
import { ApiError } from '../api/client.ts';
import { ErrorCodes } from '../constants/errorCodes.ts';

const MESSAGES: Record<string, string> = {
  [ErrorCodes.Network]: 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาลองใหม่อีกครั้ง',
  [ErrorCodes.PermissionDenied]: 'คุณไม่มีสิทธิ์ทำรายการนี้',
  [ErrorCodes.NotFound]: 'ไม่พบข้อมูล อาจถูกลบไปแล้ว',
  [ErrorCodes.RateLimited]: 'ลองหลายครั้งเกินไป กรุณารอสักครู่แล้วลองใหม่',
  [ErrorCodes.InvalidCredentials]: 'อีเมลหรือรหัสผ่านไม่ถูกต้อง',
  [ErrorCodes.InvalidPassword]: 'รหัสผ่านปัจจุบันไม่ถูกต้อง',
  [ErrorCodes.PasswordChangeRequired]: 'กรุณาเปลี่ยนรหัสผ่านก่อนใช้งาน',
  [ErrorCodes.EmailTaken]: 'อีเมลนี้มีผู้ใช้อยู่แล้ว',
  [ErrorCodes.CannotModifySelf]: 'ไม่สามารถเปลี่ยนบทบาทหรือสถานะของบัญชีตัวเองได้',
  [ErrorCodes.SuperAdminProtected]: 'ผู้ดูแลระบบสูงสุดแก้ไขได้โดยตัวเองเท่านั้น',
  [ErrorCodes.SuperAdminRequired]: 'เฉพาะผู้ดูแลระบบสูงสุดเท่านั้นที่ให้บทบาทผู้ดูแลระบบหรือแก้ไขผู้ดูแลระบบได้',
  [ErrorCodes.Validation]: 'ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง',
};

// Thai message for an error, chosen by code (never by the backend's message text).
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return MESSAGES[err.code] ?? 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง';
  return 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง';
}

// Puts VALIDATION field messages on the matching Form.Items. Returns true if any were applied.
export function applyFieldErrors(form: FormInstance, err: unknown): boolean {
  if (!(err instanceof ApiError) || err.code !== ErrorCodes.Validation || !err.fields) return false;
  const known = new Set(Object.keys(form.getFieldsValue(true)));
  const fields = Object.entries(err.fields)
    .filter(([name]) => known.has(name))
    .map(([name, message]) => ({ name, errors: [message] }));
  form.setFields(fields);
  return fields.length > 0;
}
