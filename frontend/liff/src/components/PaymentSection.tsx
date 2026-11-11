import { useState } from 'react'
import type { Deal, Payment, Totals } from '../lib/deal.ts'
import { formatPrice, inputBase } from '../lib/format.ts'
import { Label, NumberInput, Toggle } from './fields.tsx'

type Props = {
  deal: Deal
  totals: Totals
  onChange: (deal: Deal) => void
}

// Payment method, full or deposit, due dates, and the totals the customer is told.
export function PaymentSection({ deal, totals, onChange }: Props) {
  const p = deal.payment
  const set = (patch: Partial<Payment>) => onChange({ ...deal, payment: { ...p, ...patch } })

  return (
    <section>
      <Label>การชำระเงิน</Label>
      <div className="space-y-3 rounded-lg border border-gray-200 p-3">
        <Toggle
          options={[
            { value: 'transfer', label: 'โอนเงิน' },
            { value: 'credit', label: 'บัตรเครดิต' },
          ]}
          value={p.method}
          onChange={(v) => set({ method: v === 'credit' ? 'credit' : 'transfer' })}
        />
        {p.method === 'credit' && (
          <div className="flex items-center gap-2">
            <span className="min-w-0 flex-1 text-sm text-gray-600">ค่าธรรมเนียมบัตร (%)</span>
            <PercentInput value={p.creditFeePercent} onChange={(v) => set({ creditFeePercent: v })} />
          </div>
        )}

        <Toggle
          options={[
            { value: 'full', label: 'ชำระเต็มจำนวน' },
            { value: 'deposit', label: 'ชำระมัดจำ' },
          ]}
          value={p.isDeposit ? 'deposit' : 'full'}
          onChange={(v) => set({ isDeposit: v === 'deposit' })}
        />

        {p.isDeposit ? (
          <>
            <div className="flex items-center gap-2">
              <span className="min-w-0 flex-1 text-sm text-gray-600">มัดจำ/ท่าน</span>
              <NumberInput ariaLabel="มัดจำ/ท่าน" value={p.depositPerPerson} onChange={(v) => set({ depositPerPerson: v })} className="w-32 shrink-0 py-2" />
            </div>
            <DueRow label="ชำระมัดจำภายใน" date={p.depositDueDate} time={p.depositDueTime} onChange={(date, time) => set({ depositDueDate: date, depositDueTime: time })} />
            <DueRow label="ชำระส่วนเหลือภายใน" date={p.remainingDueDate} onChange={(date) => set({ remainingDueDate: date })} />
          </>
        ) : (
          <DueRow label="ชำระเต็มจำนวนภายใน" date={p.fullDueDate} time={p.fullDueTime} onChange={(date, time) => set({ fullDueDate: date, fullDueTime: time })} />
        )}

        <dl className="space-y-1 border-t border-gray-100 pt-3 text-sm">
          <TotalRow label={`รวม ${totals.travellers} ท่าน`} value={totals.subtotal} />
          {p.method === 'credit' && <TotalRow label={`ค่าธรรมเนียมบัตร (${p.creditFeePercent ?? 0}%)`} value={totals.fee} />}
          <TotalRow label="รวมเป็นเงิน" value={totals.grandTotal} strong />
          {p.isDeposit && (
            <>
              <TotalRow label="ยอดมัดจำ" value={totals.depositToPay} />
              <TotalRow label="ยอดส่วนเหลือ" value={totals.remaining} />
            </>
          )}
        </dl>
      </div>
    </section>
  )
}

// Keeps the typed text ("2." while typing 2.6) and follows outside changes (a pasted summary).
function PercentInput({ value, onChange }: { value: number | null; onChange: (v: number | null) => void }) {
  const [text, setText] = useState(value === null ? '' : String(value))
  const parsed = text === '' ? null : Number(text)
  if (parsed !== value) setText(value === null ? '' : String(value))
  return (
    <input
      type="text"
      inputMode="decimal"
      aria-label="ค่าธรรมเนียมบัตร (%)"
      value={text}
      onChange={(e) => {
        const t = e.target.value.replace(/[^\d.]/g, '').replace(/(\..*)\./g, '$1').replace(/(\.\d{2})\d+$/, '$1')
        const v = t === '' ? null : Math.min(Number(t), 20)
        setText(v !== null && v !== Number(t) ? String(v) : t)
        onChange(v)
      }}
      className={`${inputBase} w-24 shrink-0 py-2 text-right`}
    />
  )
}

function DueRow({
  label,
  date,
  time,
  onChange,
}: {
  label: string
  date: string
  time?: string
  onChange: (date: string, time: string) => void
}) {
  return (
    <div>
      <span className="mb-1 block text-xs text-gray-500">{label}</span>
      <div className="flex gap-2">
        <input type="date" aria-label={label} value={date} onChange={(e) => onChange(e.target.value, time ?? '')} className={`${inputBase} flex-1 py-2`} />
        {time !== undefined && (
          <input type="time" aria-label={`${label} (เวลา)`} value={time} onChange={(e) => onChange(date, e.target.value)} className={`${inputBase} w-32 shrink-0 py-2`} />
        )}
      </div>
    </div>
  )
}

function TotalRow({ label, value, strong }: { label: string; value: number; strong?: boolean }) {
  return (
    <div className={`flex justify-between ${strong ? 'font-semibold text-gray-900' : 'text-gray-600'}`}>
      <dt>{label}</dt>
      <dd>{formatPrice(value)} บาท</dd>
    </div>
  )
}
