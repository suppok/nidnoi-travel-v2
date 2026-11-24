import type { AIDraftPeriod } from '../../../types/aiExtract.ts';

// Draft periods of the create page (read from a PDF by AI), saved after the tour is created.
export type DraftPeriodRow = { key: string; value: AIDraftPeriod; warnings: string[] };

let nextKey = 0;
export const newDraftRowKey = () => `draft:${++nextKey}`;

/** Why the draft periods can't be saved yet, or null when they can. */
export function draftPeriodProblem(rows: DraftPeriodRow[]): string | null {
  for (const [i, { value: p }] of rows.entries()) {
    const n = `รอบที่ ${i + 1}`;
    if (!p.startDate || !p.endDate) return `${n}: เลือกวันเดินทาง`;
    if (p.endDate < p.startDate) return `${n}: วันกลับต้องไม่ก่อนวันไป`;
    if (p.price === null) return `${n}: กรอกราคาผู้ใหญ่`;
    if (p.seatsTotal === null || p.seatsAvailable === null) return `${n}: กรอกที่นั่งว่างและที่นั่งทั้งหมด`;
    if (p.seatsAvailable > p.seatsTotal) return `${n}: ที่ว่างต้องไม่เกินที่นั่งทั้งหมด`;
  }
  return null;
}
