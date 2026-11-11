const dateTime = new Intl.DateTimeFormat('th-TH', { dateStyle: 'medium', timeStyle: 'short' });
const longDate = new Intl.DateTimeFormat('th-TH', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });

export function formatDateTime(iso: string | null): string {
  return iso ? dateTime.format(new Date(iso)) : '—';
}

const baht = new Intl.NumberFormat('th-TH', { style: 'currency', currency: 'THB', maximumFractionDigits: 0 });

export function formatBaht(amount: number): string {
  return baht.format(amount);
}

export function formatLongDate(date: Date): string {
  return longDate.format(date);
}

// "สมใจ ขายเก่ง" → "สข", "Admin" → "AD"
export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) return (Array.from(parts[0])[0] + Array.from(parts[1])[0]).toUpperCase();
  return Array.from(parts[0] ?? '?').slice(0, 2).join('').toUpperCase();
}

// Thai phone numbers as stored by the backend (digits only): "0812345678" → "081 234 5678",
// "021234567" → "02 123 4567". Anything else is returned unchanged.
export function formatPhone(phone: string): string {
  if (/^0\d{9}$/.test(phone)) return `${phone.slice(0, 3)} ${phone.slice(3, 6)} ${phone.slice(6)}`;
  if (/^0\d{8}$/.test(phone)) return `${phone.slice(0, 2)} ${phone.slice(2, 5)} ${phone.slice(5)}`;
  return phone;
}
