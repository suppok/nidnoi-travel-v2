// LINE booking workflow states. MUST mirror backend/pkg/constants/linebookingstate.
export const LineBookingStates = {
  New: 10,
  Confirmed: 20,
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
  { value: LineBookingStates.New, label: 'ใหม่', color: 'red', hint: 'ยังไม่ได้ตรวจสอบ' },
  { value: LineBookingStates.Confirmed, label: 'ยืนยันแล้ว', color: 'green', hint: 'ตรวจสอบข้อมูลและสลิปแล้ว' },
  { value: LineBookingStates.Cancelled, label: 'ยกเลิก', color: 'default', hint: 'ข้อมูลไม่ถูกต้อง หรือลูกค้ายกเลิก' },
];

export function lineBookingStateInfo(state: number): { label: string; color: string } {
  return LINE_BOOKING_STATE_INFO.find((s) => s.value === state) ?? { label: String(state), color: 'default' };
}
