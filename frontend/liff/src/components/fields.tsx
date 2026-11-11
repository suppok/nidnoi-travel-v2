// Small form controls shared by the booking form sections.
import { formatPrice, inputBase } from '../lib/format.ts'

// Whole numbers only; empty = null.
export function NumberInput({
  value,
  onChange,
  placeholder,
  ariaLabel,
  className = 'w-full py-2.5',
  align = 'right',
}: {
  value: number | null
  onChange: (v: number | null) => void
  placeholder?: string
  ariaLabel?: string
  className?: string
  align?: 'left' | 'right'
}) {
  return (
    <input
      type="text"
      inputMode="numeric"
      aria-label={ariaLabel}
      value={value === null ? '' : formatPrice(value)}
      placeholder={placeholder}
      onChange={(e) => {
        const digits = e.target.value.replace(/[^\d]/g, '')
        onChange(digits === '' ? null : Math.min(Number(digits), 10_000_000))
      }}
      className={`${inputBase} ${align === 'left' ? 'text-left' : 'text-right'} ${className}`}
    />
  )
}

export function Label({ children, htmlFor, optional }: { children: React.ReactNode; htmlFor?: string; optional?: boolean }) {
  return (
    <label htmlFor={htmlFor} className="mb-1.5 block text-sm font-medium text-gray-700">
      {children}
      {optional && <span className="font-normal text-gray-400"> (ไม่บังคับ)</span>}
    </label>
  )
}

export function Toggle({
  options,
  value,
  onChange,
}: {
  options: { value: string; label: string }[]
  value: string
  onChange: (v: string) => void
}) {
  return (
    <div className="flex rounded-lg border border-gray-300 p-0.5">
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          onClick={() => onChange(o.value)}
          className={`flex-1 rounded-md py-2 text-sm transition ${value === o.value ? 'bg-blue-600 font-semibold text-white' : 'text-gray-600'}`}
        >
          {o.label}
        </button>
      ))}
    </div>
  )
}
