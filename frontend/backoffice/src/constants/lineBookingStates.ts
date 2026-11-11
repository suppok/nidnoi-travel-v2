// LINE booking workflow states. MUST mirror backend/pkg/constants/linebookingstate.
export const LineBookingStates = {
  New: 10,
  DepositPaid: 20,
  PaidInFull: 25,
  Cancelled: 30,
} as const;

export type LineBookingState = (typeof LineBookingStates)[keyof typeof LineBookingStates];

type LineBookingStateInfo = {
  value: LineBookingState;
  label: string;
  color: string;
  hint: string;
};

export const LINE_BOOKING_STATE_INFO: readonly LineBookingStateInfo[] = [
  { value: LineBookingStates.New, label: 'ใหม่', color: 'red', hint: 'รอเซลล์เจ้าของตรวจสอบและยืนยัน' },
  { value: LineBookingStates.DepositPaid, label: 'ชำระมัดจำแล้ว', color: 'blue', hint: 'ยืนยันแล้ว ลูกค้าชำระมัดจำ ส่งให้ฝ่ายการเงินออกใบเสร็จ' },
  { value: LineBookingStates.PaidInFull, label: 'ชำระครบแล้ว', color: 'green', hint: 'ยืนยันแล้ว ลูกค้าชำระเต็มจำนวน ส่งให้ฝ่ายการเงินออกใบเสร็จ' },
  { value: LineBookingStates.Cancelled, label: 'ยกเลิก', color: 'default', hint: 'ข้อมูลไม่ถูกต้อง หรือลูกค้ายกเลิก' },
];

export function lineBookingStateInfo(state: number): { label: string; color: string } {
  return LINE_BOOKING_STATE_INFO.find((s) => s.value === state) ?? { label: String(state), color: 'default' };
}

// The finance menu's page: confirmed LINE bookings (sent to the caller; admins see all), newest
// confirmation first. Same page component as /line-bookings.
export const FINANCE_LINE_BOOKINGS_PATH = '/finance/line-bookings';
