import { useState } from 'react'
import { ApiError, createLineBooking } from '../api/client.ts'
import type { TourPeriodItem, TourPublicItem } from '../types/index.ts'
import { PeriodSelector } from './PeriodSelector.tsx'
import { SlipUpload } from './SlipUpload.tsx'
import { SuccessView } from './SuccessView.tsx'
import { TourSelector } from './TourSelector.tsx'

type Props = {
  // The linked backoffice user who submits the booking.
  staffName: string
}

type SuccessData = { bookingNo: string; totalPrice: number }

function formatPrice(n: number): string {
  return n.toLocaleString('th-TH')
}

function calcTotal(period: TourPeriodItem | null, counts: Counts): number {
  if (!period) return 0
  let total = (counts.adultTwin + counts.adultSingle) * period.price
  if (counts.adultSingle > 0 && period.singleSupplement != null) {
    total += counts.adultSingle * period.singleSupplement
  }
  if (counts.childWithBed > 0 && period.childWithBedPrice != null) {
    total += counts.childWithBed * period.childWithBedPrice
  } else if (counts.childWithBed > 0) {
    total += counts.childWithBed * period.price
  }
  if (counts.childNoBed > 0 && period.childNoBedPrice != null) {
    total += counts.childNoBed * period.childNoBedPrice
  }
  return total
}

type Counts = {
  adultTwin: number
  adultSingle: number
  childWithBed: number
  childNoBed: number
}

const defaultCounts: Counts = { adultTwin: 0, adultSingle: 0, childWithBed: 0, childNoBed: 0 }

export function BookingForm({ staffName }: Props) {
  const [tour, setTour] = useState<TourPublicItem | null>(null)
  const [period, setPeriod] = useState<TourPeriodItem | null>(null)
  const [counts, setCounts] = useState<Counts>(defaultCounts)
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [slipImage, setSlipImage] = useState('')
  const [note, setNote] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState<SuccessData | null>(null)

  const totalPax = counts.adultTwin + counts.adultSingle + counts.childWithBed + counts.childNoBed
  const totalPrice = calcTotal(period, counts)

  function updateCount(key: keyof Counts, delta: number) {
    setCounts((prev) => ({
      ...prev,
      [key]: Math.max(0, prev[key] + delta),
    }))
  }

  async function handleSubmit() {
    setError('')

    if (!tour || !period) {
      setError('กรุณาเลือกทัวร์และช่วงเดินทาง')
      return
    }
    if (totalPax === 0) {
      setError('กรุณาระบุจำนวนผู้เดินทางอย่างน้อย 1 ท่าน')
      return
    }
    if (!customerName.trim()) {
      setError('กรุณาระบุชื่อลูกค้า')
      return
    }
    if (!customerPhone.trim()) {
      setError('กรุณาระบุเบอร์โทรลูกค้า')
      return
    }
    if (!slipImage) {
      setError('กรุณาอัพโหลดสลิปโอนเงิน')
      return
    }

    setSubmitting(true)
    try {
      const result = await createLineBooking({
        tourCode: tour.code,
        periodId: period.id,
        adultTwin: counts.adultTwin,
        adultSingle: counts.adultSingle,
        childWithBed: counts.childWithBed,
        childNoBed: counts.childNoBed,
        customerName: customerName.trim(),
        customerPhone: customerPhone.trim(),
        slipImage,
        note: note.trim(),
      })
      setSuccess(result)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('เกิดข้อผิดพลาด กรุณาลองอีกครั้ง')
      }
    } finally {
      setSubmitting(false)
    }
  }

  function handleReset() {
    setTour(null)
    setPeriod(null)
    setCounts(defaultCounts)
    setCustomerName('')
    setCustomerPhone('')
    setSlipImage('')
    setNote('')
    setError('')
    setSuccess(null)
  }

  if (success) {
    return (
      <SuccessView
        bookingNo={success.bookingNo}
        totalPrice={success.totalPrice}
        onReset={handleReset}
      />
    )
  }

  return (
    <div className="space-y-6 px-4 pb-8 pt-4">
      <div className="text-center">
        <h1 className="text-lg font-semibold text-gray-900">บันทึกการจอง</h1>
        <p className="text-xs text-gray-500">โดย {staffName}</p>
      </div>

      {/* Tour */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">โปรแกรมทัวร์</label>
        <TourSelector selected={tour} onSelect={setTour} />
      </section>

      {/* Period */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">ช่วงเดินทาง</label>
        <PeriodSelector tourCode={tour?.code ?? null} selected={period} onSelect={setPeriod} />
      </section>

      {/* Traveler counts */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">จำนวนผู้เดินทาง</label>
        <div className="space-y-3 rounded-lg border border-gray-200 p-3">
          <CounterRow
            label="ผู้ใหญ่ (พักคู่)"
            value={counts.adultTwin}
            onChange={(d) => updateCount('adultTwin', d)}
          />
          <CounterRow
            label="ผู้ใหญ่ (พักเดี่ยว)"
            value={counts.adultSingle}
            onChange={(d) => updateCount('adultSingle', d)}
          />
          <CounterRow
            label="เด็ก (มีเตียง)"
            value={counts.childWithBed}
            onChange={(d) => updateCount('childWithBed', d)}
          />
          <CounterRow
            label="เด็ก (ไม่มีเตียง)"
            value={counts.childNoBed}
            onChange={(d) => updateCount('childNoBed', d)}
          />
        </div>
        {period && totalPax > 0 && (
          <p className="mt-2 text-right text-sm font-medium text-gray-700">
            รวม {totalPax} ท่าน / {formatPrice(totalPrice)} บาท
          </p>
        )}
      </section>

      {/* Customer info */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">ชื่อลูกค้า</label>
        <input
          type="text"
          value={customerName}
          onChange={(e) => setCustomerName(e.target.value)}
          placeholder="ชื่อ-นามสกุล"
          className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none"
        />
      </section>

      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">เบอร์โทรลูกค้า</label>
        <input
          type="tel"
          value={customerPhone}
          onChange={(e) => setCustomerPhone(e.target.value)}
          placeholder="0xx-xxx-xxxx"
          className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none"
        />
      </section>

      {/* Slip upload */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">สลิปโอนเงิน</label>
        <SlipUpload value={slipImage} onChange={setSlipImage} />
      </section>

      {/* Note */}
      <section>
        <label className="mb-1.5 block text-sm font-medium text-gray-700">หมายเหตุ</label>
        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={3}
          placeholder="รายละเอียดเพิ่มเติม (ถ้ามี)"
          className="w-full rounded-lg border border-gray-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none"
        />
      </section>

      {/* Error */}
      {error && (
        <div className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>
      )}

      {/* Submit */}
      <button
        type="button"
        onClick={handleSubmit}
        disabled={submitting}
        className="w-full rounded-lg bg-green-600 py-3 text-sm font-semibold text-white transition hover:bg-green-700 disabled:opacity-50"
      >
        {submitting ? 'กำลังบันทึก...' : 'บันทึกการจอง'}
      </button>
    </div>
  )
}

function CounterRow({
  label,
  value,
  onChange,
}: {
  label: string
  value: number
  onChange: (delta: number) => void
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-gray-600">{label}</span>
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={() => onChange(-1)}
          disabled={value === 0}
          className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-300 text-gray-600 transition hover:bg-gray-100 disabled:opacity-30"
        >
          -
        </button>
        <span className="w-6 text-center text-sm font-medium">{value}</span>
        <button
          type="button"
          onClick={() => onChange(1)}
          className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-300 text-gray-600 transition hover:bg-gray-100"
        >
          +
        </button>
      </div>
    </div>
  )
}
