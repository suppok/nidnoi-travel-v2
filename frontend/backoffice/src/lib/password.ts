// Must match the backend password rule.
export const PASSWORD_MIN_LENGTH = 8;
export const PASSWORD_MAX_LENGTH = 72; // bcrypt input limit

export function passwordProblem(value: string): string | null {
  if (value.length < PASSWORD_MIN_LENGTH) return `รหัสผ่านต้องมีอย่างน้อย ${PASSWORD_MIN_LENGTH} ตัวอักษร`;
  if (value.length > PASSWORD_MAX_LENGTH) return `รหัสผ่านยาวได้ไม่เกิน ${PASSWORD_MAX_LENGTH} ตัวอักษร`;
  if (!/[A-Za-z]/.test(value) || !/[0-9]/.test(value)) return 'รหัสผ่านต้องมีทั้งตัวอักษรภาษาอังกฤษและตัวเลข';
  return null;
}

// No look-alike characters (0/O, 1/l/I) — staff read these out or type them from a chat message.
const LETTERS = 'abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ';
const DIGITS = '23456789';

export function generateTemporaryPassword(length = 12): string {
  const alphabet = LETTERS + DIGITS;
  for (;;) {
    const bytes = crypto.getRandomValues(new Uint32Array(length));
    const value = Array.from(bytes, (b) => alphabet[b % alphabet.length]).join('');
    if (passwordProblem(value) === null) return value;
  }
}
