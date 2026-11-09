// Booking request sales workflow. MUST mirror backend/pkg/constants/bookingstate (gaps leave room for new states).
export const BookingStates = {
  New: 10,
  Contacted: 20,
  Confirmed: 30,
  Cancelled: 40,
} as const;

export type BookingState = (typeof BookingStates)[keyof typeof BookingStates];

type BookingStateInfo = {
  value: BookingState;
  label: string;
  color: string;
  hint: string;
};

export const BOOKING_STATE_INFO: readonly BookingStateInfo[] = [
  { value: BookingStates.New, label: 'ใหม่', color: 'red', hint: 'ยังไม่มีใครติดต่อลูกค้า' },
  { value: BookingStates.Contacted, label: 'ติดต่อแล้ว', color: 'blue', hint: 'คุยกับลูกค้าแล้ว รอยืนยัน' },
  { value: BookingStates.Confirmed, label: 'ยืนยันแล้ว', color: 'green', hint: 'ลูกค้ายืนยันการเดินทาง' },
  { value: BookingStates.Cancelled, label: 'ยกเลิก', color: 'default', hint: 'ลูกค้าไม่ไป ติดต่อไม่ได้ หรือข้อมูลไม่ถูกต้อง' },
];

// Unknown values (a state added on the backend first) still render.
export function bookingStateInfo(state: number): { label: string; color: string } {
  return BOOKING_STATE_INFO.find((s) => s.value === state) ?? { label: String(state), color: 'default' };
}
