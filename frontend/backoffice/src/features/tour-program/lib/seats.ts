// Seats a period gets when its total changes: available follows the total while it still equals it
// (a new departure, or nothing booked yet), otherwise it stays but never above the new total.
export function seatsForTotal(prevTotal: number | null, prevAvailable: number | null, nextTotal: number) {
  const available = prevAvailable === null || prevAvailable === prevTotal ? nextTotal : Math.min(prevAvailable, nextTotal);
  return { seatsTotal: nextTotal, seatsAvailable: available };
}
