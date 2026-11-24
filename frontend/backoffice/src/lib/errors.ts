import type { FormInstance } from 'antd';
import { ApiError } from '../api/client.ts';
import { ErrorCodes } from '../constants/errorCodes.ts';

const MESSAGES: Record<string, string> = {
  [ErrorCodes.Network]: 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาลองใหม่อีกครั้ง',
  [ErrorCodes.PermissionDenied]: 'คุณไม่มีสิทธิ์ทำรายการนี้',
  [ErrorCodes.NotFound]: 'ไม่พบข้อมูล อาจถูกลบไปแล้ว',
  [ErrorCodes.RateLimited]: 'ลองหลายครั้งเกินไป กรุณารอสักครู่แล้วลองใหม่',
  [ErrorCodes.InvalidCredentials]: 'อีเมลหรือรหัสผ่านไม่ถูกต้อง',
  [ErrorCodes.LineNotLinked]: 'บัญชี LINE นี้ยังไม่ได้เชื่อมกับผู้ใช้ในระบบหลังบ้าน',
  [ErrorCodes.InvalidLineLogin]: 'เข้าสู่ระบบด้วย LINE ไม่สำเร็จหรือหมดเวลา กรุณาลองใหม่',
  [ErrorCodes.InvalidPassword]: 'รหัสผ่านปัจจุบันไม่ถูกต้อง',
  [ErrorCodes.PasswordChangeRequired]: 'กรุณาเปลี่ยนรหัสผ่านก่อนใช้งาน',
  [ErrorCodes.EmailTaken]: 'อีเมลนี้มีผู้ใช้อยู่แล้ว',
  [ErrorCodes.CannotModifySelf]: 'ไม่สามารถเปลี่ยนบทบาทหรือสถานะของบัญชีตัวเองได้',
  [ErrorCodes.SuperAdminProtected]: 'ผู้ดูแลระบบสูงสุดแก้ไขได้โดยตัวเองเท่านั้น',
  [ErrorCodes.SuperAdminRequired]: 'เฉพาะผู้ดูแลระบบสูงสุดเท่านั้นที่ให้บทบาทผู้ดูแลระบบหรือแก้ไขผู้ดูแลระบบได้',
  [ErrorCodes.SlugTaken]: 'Slug นี้ถูกใช้แล้ว',
  [ErrorCodes.RefIdTaken]: 'Ref ID นี้ถูกใช้แล้ว',
  [ErrorCodes.CodeTaken]: 'รหัสนี้ถูกใช้แล้ว',
  [ErrorCodes.NameTaken]: 'ชื่อนี้ถูกใช้แล้ว',
  [ErrorCodes.MediaTypeNotAllowed]: 'ไฟล์ประเภทนี้อัปโหลดในช่องนี้ไม่ได้',
  [ErrorCodes.MediaTooLarge]: 'ไฟล์ใหญ่เกินไป (ไม่เกิน 10 MB)',
  [ErrorCodes.ImageTooSmall]: 'รูปมีขนาดเล็กเกินไป',
  [ErrorCodes.StorageUnavailable]: 'ระบบจัดเก็บไฟล์ยังไม่พร้อมใช้งาน กรุณาติดต่อผู้ดูแลระบบ',
  [ErrorCodes.SiteRebuildUnavailable]: 'ยังไม่ได้ตั้งค่าการอัปเดตหน้าเว็บ กรุณาติดต่อผู้ดูแลระบบ',
  [ErrorCodes.SiteRebuildFailed]: 'สั่งอัปเดตหน้าเว็บไม่สำเร็จ กรุณาลองใหม่อีกครั้ง',
  [ErrorCodes.LineNotConfigured]: 'ยังไม่ได้ตั้งค่าการเชื่อมต่อ LINE กรุณาติดต่อผู้ดูแลระบบ',
  [ErrorCodes.LinePushFailed]: 'ส่งข้อความ LINE ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง',
  [ErrorCodes.LineQuotaExceeded]: 'LINE OA ส่งข้อความครบโควตาของเดือนนี้แล้ว กรุณาติดต่อผู้ดูแลระบบ',
  [ErrorCodes.LineBookingNotNew]: 'ใบจองนี้ถูกยืนยันหรือยกเลิกไปแล้ว',
  [ErrorCodes.LineBookingNotEditable]: 'เปิดได้เฉพาะใบจองที่เซลล์ยืนยันการชำระเงินแล้ว (ชำระมัดจำหรือชำระครบ)',
  [ErrorCodes.LineAlreadyLinked]: 'บัญชี LINE นี้เชื่อมกับผู้ใช้อื่นอยู่แล้ว',
  [ErrorCodes.PeriodUnavailable]: 'รอบเดินทางนี้ปิดรับจองหรือที่นั่งเต็มแล้ว กรุณาเลือกรอบใหม่',
  [ErrorCodes.LineBookingDocumentsNotReady]: 'เอกสารนี้ไม่ได้รอตรวจแล้ว (อาจถูกยืนยันหรือส่งกลับไปแล้ว)',
  [ErrorCodes.LineBookingBalanceNotPending]: 'ไม่มียอดคงเหลือรอยืนยันแล้ว (อาจถูกยืนยันไปแล้ว)',
  [ErrorCodes.AINotConfigured]: 'ยังไม่ได้ตั้งค่า AI หรือปิดใช้งานอยู่ กรุณาติดต่อผู้ดูแลระบบ',
  [ErrorCodes.AIInvalidApiKey]: 'API key ของ Anthropic ไม่ถูกต้องหรือถูกยกเลิกแล้ว',
  [ErrorCodes.AIModelUnavailable]: 'API key นี้ใช้โมเดลที่เลือกไม่ได้',
  [ErrorCodes.AIBusy]: 'ระบบ AI ไม่ว่างในขณะนี้ กรุณาลองใหม่ในอีกสักครู่',
  [ErrorCodes.AIRequestFailed]: 'AI อ่านไฟล์ไม่สำเร็จ กรุณาลองใหม่ หรือกรอกข้อมูลเอง',
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
