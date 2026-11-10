import { useEffect, useRef, useState } from 'react'
import { searchLineContacts } from '../api/client.ts'
import type { LineContact } from '../types/index.ts'

type Props = {
  selected: LineContact | null
  onSelect: (contact: LineContact | null) => void
}

const SEARCH_DEBOUNCE_MS = 300

const timeFormat = new Intl.DateTimeFormat('th-TH', {
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Bangkok',
})

function ContactAvatar({ contact }: { contact: LineContact }) {
  if (contact.pictureUrl) {
    return <img src={contact.pictureUrl} alt="" className="h-8 w-8 flex-none rounded-full object-cover" />
  }
  return (
    <span className="flex h-8 w-8 flex-none items-center justify-center rounded-full bg-gray-200 text-xs text-gray-500">
      {contact.displayName.slice(0, 1) || '?'}
    </span>
  )
}

// Optional: the customer's chat with the OA, so the receipt can be pushed to it later. Customers
// appear once they have messaged or followed the OA; skip it when the customer chats on a
// personal LINE account.
export function CustomerChatPicker({ selected, onSelect }: Props) {
  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<LineContact[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const seq = useRef(0)

  // Search while the list is open; an empty keyword lists the most recent chats.
  useEffect(() => {
    if (!open || selected) return
    const current = ++seq.current
    const timer = setTimeout(async () => {
      setLoading(true)
      setError('')
      try {
        const items = await searchLineContacts(keyword.trim())
        if (current === seq.current) setResults(items)
      } catch {
        if (current === seq.current) setError('ค้นหาไม่สำเร็จ กรุณาลองอีกครั้ง')
      } finally {
        if (current === seq.current) setLoading(false)
      }
    }, SEARCH_DEBOUNCE_MS)
    return () => clearTimeout(timer)
  }, [keyword, open, selected])

  if (selected) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
        <div className="flex items-center gap-3">
          <ContactAvatar contact={selected} />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-gray-900">{selected.displayName || '(ไม่มีชื่อ)'}</p>
            {!selected.following && <p className="text-xs text-red-600">ลูกค้าบล็อก OA — ส่งข้อความทาง LINE ไม่ได้</p>}
          </div>
          <button
            type="button"
            onClick={() => {
              onSelect(null)
              setKeyword('')
            }}
            className="text-sm text-blue-600"
          >
            เปลี่ยน
          </button>
        </div>
      </div>
    )
  }

  const showList = open && !loading && (results.length > 0 || error !== '')
  const showEmpty = open && !loading && error === '' && results.length === 0 && keyword.trim() !== ''

  return (
    <div className="relative">
      <input
        type="search"
        value={keyword}
        onChange={(e) => {
          setKeyword(e.target.value)
          setOpen(true)
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        placeholder="ค้นหาชื่อ LINE ของลูกค้า"
        className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none"
      />
      {loading && <div className="absolute right-3 top-3 text-xs text-gray-400">กำลังค้นหา...</div>}
      {showList && (
        <ul className="absolute z-10 mt-1 max-h-72 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg">
          {error !== '' && <li className="px-3 py-2.5 text-sm text-red-600">{error}</li>}
          {results.map((c) => (
            <li key={c.id}>
              <button
                type="button"
                className="flex w-full items-center gap-3 px-3 py-2.5 text-left hover:bg-blue-50"
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => {
                  onSelect(c)
                  setOpen(false)
                }}
              >
                <ContactAvatar contact={c} />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium text-gray-900">{c.displayName || '(ไม่มีชื่อ)'}</span>
                  {c.lastMessageAt && (
                    <span className="block text-xs text-gray-500">แชทล่าสุด {timeFormat.format(new Date(c.lastMessageAt))}</span>
                  )}
                </span>
                {!c.following && <span className="text-xs text-red-600">บล็อก</span>}
              </button>
            </li>
          ))}
        </ul>
      )}
      {showEmpty && <p className="mt-1 text-xs text-gray-500">ไม่พบแชท — ลูกค้าต้องทักแชท OA ก่อน</p>}
    </div>
  )
}
