import { useEffect, useId, useRef, useState } from 'react'

export type SelectOption = { value: string; label: string; hint?: string }

type Props = {
  id?: string
  value: string
  options: SelectOption[]
  onChange: (value: string) => void
  placeholder: string
  disabled?: boolean
}

// A styled dropdown (native <select> popups can't be styled). Click or Enter/Space opens it;
// arrows move, Enter picks, Escape or a click outside closes it.
export function Select({ id, value, options, onChange, placeholder, disabled }: Props) {
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(-1)
  const rootRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLUListElement>(null)
  const listId = useId()
  const selected = options.find((o) => o.value === value) ?? null

  useEffect(() => {
    if (!open) return
    const close = (e: PointerEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('pointerdown', close)
    return () => document.removeEventListener('pointerdown', close)
  }, [open])

  // Keep the highlighted option in view while moving with the keyboard.
  useEffect(() => {
    if (open && active >= 0) listRef.current?.children[active]?.scrollIntoView({ block: 'nearest' })
  }, [open, active])

  function show() {
    setActive(Math.max(0, options.findIndex((o) => o.value === value)))
    setOpen(true)
  }

  function pick(v: string) {
    onChange(v)
    setOpen(false)
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (disabled) return
    if (!open) {
      if (['Enter', ' ', 'ArrowDown', 'ArrowUp'].includes(e.key)) {
        e.preventDefault()
        show()
      }
      return
    }
    if (e.key === 'Escape' || e.key === 'Tab') {
      setOpen(false)
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActive((i) => Math.min(options.length - 1, i + 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive((i) => Math.max(0, i - 1))
    } else if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      if (options[active]) pick(options[active].value)
    }
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        id={id}
        type="button"
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        aria-activedescendant={open && active >= 0 ? `${listId}-${active}` : undefined}
        disabled={disabled}
        onClick={() => (open ? setOpen(false) : show())}
        onKeyDown={onKeyDown}
        className={`flex w-full items-center gap-2 rounded-lg border bg-white px-3 py-2.5 text-left text-sm transition focus:outline-none disabled:bg-gray-100 ${
          open ? 'border-blue-500 ring-2 ring-blue-100' : 'border-gray-300 focus:border-blue-500'
        }`}
      >
        <span className={`min-w-0 flex-1 truncate ${selected ? 'text-gray-900' : 'text-gray-400'}`}>{selected?.label ?? placeholder}</span>
        <svg
          className={`h-4 w-4 shrink-0 text-gray-400 transition-transform ${open ? 'rotate-180' : ''}`}
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden="true"
        >
          <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 0 1 1.06.02L10 11.06l3.71-3.83a.75.75 0 1 1 1.08 1.04l-4.25 4.39a.75.75 0 0 1-1.08 0L5.21 8.27a.75.75 0 0 1 .02-1.06Z" clipRule="evenodd" />
        </svg>
      </button>

      {open && (
        <ul
          ref={listRef}
          id={listId}
          role="listbox"
          className="absolute inset-x-0 top-full z-20 mt-1.5 max-h-64 overflow-auto rounded-xl border border-gray-100 bg-white p-1 shadow-lg ring-1 ring-black/5"
        >
          {options.length === 0 && <li className="px-3 py-2.5 text-sm text-gray-400">ไม่มีตัวเลือก</li>}
          {options.map((o, i) => {
            const isSelected = o.value === value
            return (
              <li
                key={o.value}
                id={`${listId}-${i}`}
                role="option"
                aria-selected={isSelected}
                onPointerEnter={() => setActive(i)}
                onClick={() => pick(o.value)}
                className={`flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2.5 text-sm ${
                  isSelected ? 'font-medium text-blue-700' : 'text-gray-700'
                } ${i === active ? (isSelected ? 'bg-blue-100' : 'bg-gray-100') : isSelected ? 'bg-blue-50' : ''}`}
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate">{o.label}</span>
                  {o.hint && <span className="block truncate text-xs font-normal text-gray-400">{o.hint}</span>}
                </span>
                {isSelected && (
                  <svg className="h-4 w-4 shrink-0 text-blue-600" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fillRule="evenodd" d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.5 7.5a1 1 0 0 1-1.4 0l-3.5-3.5a1 1 0 1 1 1.4-1.4l2.8 2.79 6.8-6.79a1 1 0 0 1 1.4 0Z" clipRule="evenodd" />
                  </svg>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
