import { DEAL_ROWS, type Deal, type Line, travellers } from '../lib/deal.ts'
import { inputBase } from '../lib/format.ts'
import { NumberInput } from './fields.tsx'

type Props = {
  deal: Deal
  onChange: (deal: Deal) => void
}

// Column widths shared by the header and every row.
const COUNT_COL = 'w-16 shrink-0'
const PRICE_COL = 'w-28 shrink-0'
const cell = 'py-2'

// Travellers and extra charges as one table (label, count, unit price), like the backoffice
// "สรุปยอดโอน" tool, then the discount.
export function DealSection({ deal, onChange }: Props) {
  const pax = travellers(deal)

  function getLine(key: (typeof DEAL_ROWS)[number]['key']): Line {
    return key === 'adult' || key === 'child' || key === 'infant' || key === 'singleRoom' ? deal[key] : deal.charges[key]
  }
  function setLine(key: (typeof DEAL_ROWS)[number]['key'], patch: Partial<Line>) {
    if (key === 'adult' || key === 'child' || key === 'infant' || key === 'singleRoom') {
      onChange({ ...deal, [key]: { ...deal[key], ...patch } })
    } else {
      onChange({ ...deal, charges: { ...deal.charges, [key]: { ...deal.charges[key], ...patch } } })
    }
  }
  const setOption = (i: number, patch: Partial<Deal['options'][number]>) =>
    onChange({ ...deal, options: deal.options.map((o, j) => (j === i ? { ...o, ...patch } : o)) })

  return (
    <section className="overflow-hidden rounded-xl border border-gray-200 bg-white">
      <h2 className="border-b border-gray-100 px-3 py-2.5 text-sm font-semibold text-gray-900">ผู้เดินทางและราคา</h2>
      <div className="space-y-2 p-3">
        <div className="flex items-end gap-2 text-xs text-gray-500">
          <span className="flex-1" />
          <span className={COUNT_COL}>จำนวน</span>
          <span className={PRICE_COL}>ราคา/หน่วย (บาท)</span>
        </div>

        {DEAL_ROWS.map((row) => {
          const line = getLine(row.key)
          return (
            <div key={row.key} className="flex items-center gap-2">
              <span className="min-w-0 flex-1 text-sm text-gray-700">{row.label}</span>
              <NumberInput
                ariaLabel={`จำนวน ${row.label}`}
                value={line.count === 0 ? null : line.count}
                onChange={(v) => setLine(row.key, { count: v })}
                placeholder={row.perTraveller ? String(pax) : '0'}
                align="left"
                className={`${COUNT_COL} ${cell}`}
              />
              <NumberInput
                ariaLabel={`ราคา ${row.label}`}
                value={line.price}
                onChange={(v) => setLine(row.key, { price: v })}
                placeholder="0"
                align="left"
                className={`${PRICE_COL} ${cell}`}
              />
            </div>
          )
        })}

        {deal.options.map((o, i) => (
          <div key={i} className="flex items-center gap-2">
            <div className="flex min-w-0 flex-1 items-center gap-1">
              <button
                type="button"
                onClick={() => onChange({ ...deal, options: deal.options.filter((_, j) => j !== i) })}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-gray-400 transition hover:bg-red-50 hover:text-red-500"
                aria-label="ลบรายการ"
              >
                ×
              </button>
              <input
                value={o.name}
                onChange={(e) => setOption(i, { name: e.target.value })}
                placeholder="ชื่อรายการ"
                aria-label="ชื่อรายการ"
                className={`${inputBase} w-full ${cell}`}
              />
            </div>
            <NumberInput ariaLabel="จำนวน รายการเพิ่มเติม" value={o.count} onChange={(v) => setOption(i, { count: v })} placeholder="0" align="left" className={`${COUNT_COL} ${cell}`} />
            <NumberInput ariaLabel="ราคา รายการเพิ่มเติม" value={o.price} onChange={(v) => setOption(i, { price: v })} placeholder="0" align="left" className={`${PRICE_COL} ${cell}`} />
          </div>
        ))}

        <button
          type="button"
          onClick={() => onChange({ ...deal, options: [...deal.options, { name: '', count: 1, price: null }] })}
          className="mt-1 inline-flex items-center gap-1.5 rounded-lg border border-dashed border-gray-300 px-3 py-2 text-sm text-gray-700 transition hover:border-blue-400 hover:text-blue-700"
        >
          <span className="text-base leading-none">+</span> เพิ่มรายการ
        </button>

        <div className="grid grid-cols-2 gap-3 pt-2">
          <div>
            <span className="mb-1 block text-sm text-gray-700">ส่วนลด/ท่าน (บาท)</span>
            <NumberInput
              ariaLabel="ส่วนลด/ท่าน"
              value={deal.discountPerPerson}
              onChange={(v) => onChange({ ...deal, discountPerPerson: v })}
              placeholder="0"
              align="left"
              className={`w-full ${cell}`}
            />
            <span className="mt-1 block text-xs text-gray-400">หักจากราคาผู้ใหญ่ เด็ก และทารก</span>
          </div>
          <div>
            <span className="mb-1 block text-sm text-gray-700">ราคาก่อนลด (บาท)</span>
            <NumberInput
              ariaLabel="ราคาผู้ใหญ่ก่อนลด"
              value={deal.adultPriceBeforeDiscount}
              onChange={(v) => onChange({ ...deal, adultPriceBeforeDiscount: v })}
              placeholder="-"
              align="left"
              className={`w-full ${cell}`}
            />
            <span className="mt-1 block text-xs text-gray-400">แสดงเป็น "(จาก …)" ข้างราคาผู้ใหญ่</span>
          </div>
        </div>
      </div>
    </section>
  )
}
