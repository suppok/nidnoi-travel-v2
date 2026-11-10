import { useCallback, useEffect, useRef, useState } from 'react'
import { searchTours } from '../api/client.ts'
import type { TourPublicItem } from '../types/index.ts'

type Props = {
  onSelect: (tour: TourPublicItem) => void
  selected: TourPublicItem | null
}

export function TourSelector({ onSelect, selected }: Props) {
  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<TourPublicItem[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const search = useCallback(async (q: string) => {
    if (q.length < 2) {
      setResults([])
      return
    }
    setLoading(true)
    try {
      const items = await searchTours(q)
      setResults(items.filter((t) => t.isSelling))
    } catch {
      setResults([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => search(keyword), 300)
    return () => clearTimeout(debounceRef.current)
  }, [keyword, search])

  if (selected) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
        <div className="flex items-center justify-between">
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-gray-900">{selected.name}</p>
            <p className="text-xs text-gray-500">{selected.code}</p>
          </div>
          <button
            type="button"
            onClick={() => {
              onSelect(null as unknown as TourPublicItem)
              setKeyword('')
              setResults([])
            }}
            className="ml-2 text-sm text-blue-600"
          >
            เปลี่ยน
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="relative">
      <input
        type="text"
        value={keyword}
        onChange={(e) => {
          setKeyword(e.target.value)
          setOpen(true)
        }}
        onFocus={() => setOpen(true)}
        placeholder="ค้นหาทัวร์ (ชื่อ หรือ รหัส)"
        className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none"
      />
      {loading && (
        <div className="absolute right-3 top-3 text-xs text-gray-400">กำลังค้นหา...</div>
      )}
      {open && results.length > 0 && (
        <ul className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg">
          {results.map((tour) => (
            <li key={tour.id}>
              <button
                type="button"
                className="w-full px-3 py-2.5 text-left hover:bg-blue-50"
                onClick={() => {
                  onSelect(tour)
                  setOpen(false)
                }}
              >
                <p className="text-sm font-medium text-gray-900">{tour.name}</p>
                <p className="text-xs text-gray-500">
                  {tour.code}
                  {tour.countries.length > 0 &&
                    ` - ${tour.countries.map((c) => c.nameTh).join(', ')}`}
                </p>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
