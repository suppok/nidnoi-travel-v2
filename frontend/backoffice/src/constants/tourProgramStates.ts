// Tour program workflow states. MUST mirror backend/pkg/constants/tourprogramstate (gaps leave room for new states).
export const TourProgramStates = {
  Draft: 10,
  Selling: 20,
  Closed: 30,
  Cancelled: 40,
} as const;

export type TourProgramState = (typeof TourProgramStates)[keyof typeof TourProgramStates];

type TourProgramStateInfo = {
  value: TourProgramState;
  label: string;
  color: string;
  hint: string;
};

export const TOUR_PROGRAM_STATE_INFO: readonly TourProgramStateInfo[] = [
  { value: TourProgramStates.Draft, label: 'ร่าง', color: 'default', hint: 'ยังไม่แสดงบนหน้าเว็บ' },
  { value: TourProgramStates.Selling, label: 'เปิดขาย', color: 'green', hint: 'แสดงบนหน้าเว็บและรับสอบถาม' },
  { value: TourProgramStates.Closed, label: 'ปิดรับ / เต็ม', color: 'gold', hint: 'แสดงบนหน้าเว็บว่าปิดรับแล้ว' },
  { value: TourProgramStates.Cancelled, label: 'ยกเลิก', color: 'red', hint: 'ไม่แสดงบนหน้าเว็บ' },
];

// Unknown values (a state added on the backend first) still render.
export function tourProgramStateInfo(state: number): { label: string; color: string; hint: string } {
  return TOUR_PROGRAM_STATE_INFO.find((s) => s.value === state) ?? { label: String(state), color: 'default', hint: '' };
}
