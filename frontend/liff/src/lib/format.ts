// Input look without width or vertical padding, so callers can set those without class clashes
// (two width utilities on one element: the CSS order decides, not the class order).
export const inputBase =
  'min-w-0 rounded-lg border border-gray-300 bg-white px-3 text-sm focus:border-blue-500 focus:outline-none disabled:bg-gray-100'

// A full-width input.
export const inputClass = `${inputBase} w-full py-2.5`

export function formatPrice(n: number): string {
  return n.toLocaleString('th-TH')
}
