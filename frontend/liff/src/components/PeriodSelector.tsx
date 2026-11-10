import { useEffect, useState } from 'react'
import { getTourPeriods } from '../api/client.ts'
import type { TourPeriodItem } from '../types/index.ts'

type Props = {
  tourCode: string | null
  onSelect: (period: TourPeriodItem | null) => void
  selected: TourPeriodItem | null
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString('th-TH', { day: 'numeric', month: 'short', year: '2-digit' })
}

function formatPrice(n: number): string {
  return n.toLocaleString('th-TH')
}

export function PeriodSelector({ tourCode, onSelect, selected }: Props) {
  const [periods, setPeriods] = useState<TourPeriodItem[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!tourCode) {
      setPeriods([])
      onSelect(null)
      return
    }

    let cancelled = false
    setLoading(true)

    getTourPeriods(tourCode)
      .then((items) => {
        if (cancelled) return
        // Show only active periods with seats
        setPeriods(items.filter((p) => p.status && p.seatsAvailable > 0))
      })
      .catch(() => {
        if (!cancelled) setPeriods([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tourCode])

  if (!tourCode) {
    return <p className="text-sm text-gray-400">เลือกทัวร์ก่อน</p>
  }

  if (loading) {
    return <p className="text-sm text-gray-400">กำลังโหลดช่วงเดินทาง...</p>
  }

  if (periods.length === 0) {
    return <p className="text-sm text-gray-400">ไม่มีช่วงเดินทางที่เปิดขาย</p>
  }

  return (
    <div className="space-y-2">
      {periods.map((p) => {
        const isSelected = selected?.id === p.id
        return (
          <button
            key={p.id}
            type="button"
            onClick={() => onSelect(isSelected ? null : p)}
            className={`w-full rounded-lg border p-3 text-left transition ${
              isSelected
                ? 'border-blue-500 bg-blue-50'
                : 'border-gray-200 bg-white hover:border-blue-300'
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">
                {formatDate(p.startDate)} - {formatDate(p.endDate)}
              </span>
              <span className="text-xs text-gray-500">ว่าง {p.seatsAvailable} ที่นั่ง</span>
            </div>
            <div className="mt-1 text-sm text-gray-600">
              ราคา {formatPrice(p.price)} บาท/ท่าน
              {p.singleSupplement != null && (
                <span className="ml-2 text-xs text-gray-400">
                  (พักเดี่ยว +{formatPrice(p.singleSupplement)})
                </span>
              )}
            </div>
          </button>
        )
      })}
    </div>
  )
}
