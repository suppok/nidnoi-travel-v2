import { useEffect, useRef, useState } from 'react'
import { ApiError, createLineBooking, getBookingSources, getSalesUsers, searchTours } from '../api/client.ts'
import { applyPeriod, computeTotals, type Deal, dealError, dealFromSummary, dealRequest, emptyDeal } from '../lib/deal.ts'
import type { ParsedSummary } from '../lib/parseSummary.ts'
import type { BookingSource, LineContact, SalesUser, TourPeriodItem, TourPublicItem } from '../types/index.ts'
import { CustomerChatPicker } from './CustomerChatPicker.tsx'
import { DealSection } from './DealSection.tsx'
import { formatPrice, inputClass } from '../lib/format.ts'
import { Label } from './fields.tsx'
import { PaymentSection } from './PaymentSection.tsx'
import { PeriodSelector } from './PeriodSelector.tsx'
import { Select } from './Select.tsx'
import { SlipUpload } from './SlipUpload.tsx'
import { SuccessView } from './SuccessView.tsx'
import { type PasteReport, SummaryPaste, SummaryPasteStart } from './SummaryPaste.tsx'
import { TourSelector } from './TourSelector.tsx'

type Props = {
  // The linked backoffice user who submits the booking.
  staffName: string
}

type SuccessData = { bookingNo: string; totalPrice: number }

export function BookingForm({ staffName }: Props) {
  const [tour, setTour] = useState<TourPublicItem | null>(null)
  const [period, setPeriod] = useState<TourPeriodItem | null>(null)
  const [deal, setDeal] = useState<Deal>(emptyDeal)
  const [salesUsers, setSalesUsers] = useState<SalesUser[] | null>(null)
  const [salesUsersError, setSalesUsersError] = useState('')
  const [salesUserId, setSalesUserId] = useState('')
  const [sources, setSources] = useState<BookingSource[] | null>(null)
  const [bookingSourceId, setBookingSourceId] = useState('')
  const [refName, setRefName] = useState('')
  const [refBookingNo, setRefBookingNo] = useState('')
  const [lineContact, setLineContact] = useState<LineContact | null>(null)
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [customerAddress, setCustomerAddress] = useState('')
  const [slipImage, setSlipImage] = useState('')
  const [note, setNote] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState<SuccessData | null>(null)
  // Prefill from a pasted sales-tool summary.
  const [pasteReport, setPasteReport] = useState<PasteReport | null>(null)
  const [autoPeriod, setAutoPeriod] = useState<{ startDate: string; endDate: string } | null>(null)
  const [pastedTotal, setPastedTotal] = useState<number | null>(null)
  const [pasteSeq, setPasteSeq] = useState(0) // remounts PeriodSelector so it reloads and auto-selects
  // True while the period picked by a paste is pending: its prices must not replace the pasted ones.
  const keepDealRef = useRef(false)
  // The first screen asks for a summary to paste; skipping (or pasting) shows the form.
  const [started, setStarted] = useState(false)
  // Sales users and booking sources: a paste right after opening waits for them.
  const lookupsRef = useRef<Promise<{ sales: SalesUser[]; sources: BookingSource[] }> | null>(null)

  const totals = computeTotals(deal)
  const source = sources?.find((s) => s.id === bookingSourceId) ?? null

  // Sales owner options (the submitter is preselected when they are sales) and booking sources.
  useEffect(() => {
    let cancelled = false
    const sales = getSalesUsers()
      .then((res) => {
        if (!cancelled) {
          setSalesUsers(res.items)
          setSalesUserId((current) => current || res.defaultId)
        }
        return res.items
      })
      .catch(() => {
        if (!cancelled) setSalesUsersError('โหลดรายชื่อเซลล์ไม่สำเร็จ กรุณาเปิดแอปใหม่')
        return [] as SalesUser[]
      })
    const bookingSources = getBookingSources()
      .then((items) => {
        if (!cancelled) setSources(items)
        return items
      })
      .catch(() => {
        if (!cancelled) setSources([])
        return [] as BookingSource[]
      })
    lookupsRef.current = Promise.all([sales, bookingSources]).then(([s, b]) => ({ sales: s, sources: b }))
    return () => {
      cancelled = true
    }
  }, [])

  function handlePeriodSelect(p: TourPeriodItem | null) {
    setPeriod(p)
    if (!p) return
    if (!keepDealRef.current) setDeal((d) => applyPeriod(d, p))
    keepDealRef.current = false
  }

  async function applySummary(s: ParsedSummary) {
    const filled: string[] = []
    const warnings: string[] = []
    const lookups = (await lookupsRef.current) ?? { sales: [], sources: [] }

    if (s.salesName) {
      const name = s.salesName
      // The summary writes "<name> <phone>"; prefer the longest matching name.
      const match = [...lookups.sales]
        .sort((a, b) => b.name.length - a.name.length)
        .find((u) => name === u.name || name.startsWith(`${u.name} `))
      if (match) {
        setSalesUserId(match.id)
        filled.push('เซลล์')
      } else {
        warnings.push(`ไม่พบเซลล์ "${name}" กรุณาเลือกเอง`)
      }
    }
    if (s.source) {
      const match = lookups.sources.find((x) => x.name.toLowerCase() === s.source?.toLowerCase())
      if (match) {
        setBookingSourceId(match.id)
        filled.push('ช่องทางการจอง')
      } else {
        warnings.push(`ไม่พบช่องทาง "${s.source}" กรุณาเลือกเอง`)
      }
    }
    if (s.refName) setRefName(s.refName)
    if (s.refBookingNo) setRefBookingNo(s.refBookingNo)
    if (s.customerName) {
      setCustomerName(s.customerName)
      filled.push('ชื่อลูกค้า')
    }
    if (s.customerPhone) {
      setCustomerPhone(s.customerPhone)
      filled.push('เบอร์โทร')
    }
    if (s.customerAddress) {
      setCustomerAddress(s.customerAddress)
      filled.push('ที่อยู่')
    }

    if (s.travellers) {
      setDeal(dealFromSummary(s))
      keepDealRef.current = true
      filled.push('ผู้เดินทาง ราคา และการชำระเงิน')
    }
    setPastedTotal(s.total)

    const keyword = s.tourCode ?? s.tourName
    if (keyword) {
      try {
        const items = await searchTours(keyword)
        const tourMatch = s.tourCode
          ? items.find((t) => t.code.toLowerCase() === s.tourCode?.toLowerCase())
          : items.find((t) => t.name === s.tourName)
        if (tourMatch) {
          setTour(tourMatch)
          setPeriod(null)
          setAutoPeriod(s.startDate && s.endDate ? { startDate: s.startDate, endDate: s.endDate } : null)
          setPasteSeq((n) => n + 1)
          filled.push('โปรแกรมทัวร์')
          if (!s.startDate || !s.endDate) warnings.push('ไม่พบวันเดินทางในข้อความ กรุณาเลือกช่วงเดินทางเอง')
        } else {
          warnings.push(`ไม่พบโปรแกรม "${keyword}" ที่เปิดขาย กรุณาเลือกเอง`)
        }
      } catch {
        warnings.push('ค้นหาโปรแกรมทัวร์ไม่สำเร็จ กรุณาเลือกเอง')
      }
    }

    setPasteReport({ filled, warnings })
  }

  function handleAutoPeriod(found: boolean) {
    const want = autoPeriod
    setAutoPeriod(null)
    keepDealRef.current = false
    setPasteReport((prev) => {
      if (!prev) return prev
      return found
        ? { ...prev, filled: [...prev.filled, 'ช่วงเดินทาง'] }
        : { ...prev, warnings: [...prev.warnings, `ไม่พบรอบ ${want?.startDate ?? ''} – ${want?.endDate ?? ''} ที่ยังว่าง กรุณาเลือกเอง`] }
    })
  }

  async function handleSubmit() {
    setError('')

    if (!salesUserId) return setError('กรุณาเลือกเซลล์เจ้าของการจอง')
    if (!tour || !period) return setError('กรุณาเลือกทัวร์และช่วงเดินทาง')
    const invalid = dealError(deal)
    if (invalid) return setError(invalid)
    if (!customerName.trim()) return setError('กรุณาระบุชื่อลูกค้า')
    if (!customerPhone.trim()) return setError('กรุณาระบุเบอร์โทรลูกค้า')
    if (!slipImage) return setError('กรุณาอัพโหลดสลิปโอนเงิน')

    setSubmitting(true)
    try {
      const result = await createLineBooking({
        salesUserId,
        lineContactId: lineContact?.id ?? null,
        bookingSourceId: bookingSourceId || null,
        refName: refName.trim(),
        refBookingNo: refBookingNo.trim(),
        tourCode: tour.code,
        periodId: period.id,
        ...dealRequest(deal),
        customerName: customerName.trim(),
        customerPhone: customerPhone.trim(),
        customerAddress: customerAddress.trim(),
        slipImage,
        note: note.trim(),
      })
      setSuccess(result)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'เกิดข้อผิดพลาด กรุณาลองอีกครั้ง')
    } finally {
      setSubmitting(false)
    }
  }

  function handleReset() {
    setTour(null)
    setPeriod(null)
    setDeal(emptyDeal())
    setBookingSourceId('')
    setRefName('')
    setRefBookingNo('')
    setLineContact(null)
    setCustomerName('')
    setCustomerPhone('')
    setCustomerAddress('')
    setSlipImage('')
    setNote('')
    setError('')
    setSuccess(null)
    setPasteReport(null)
    setAutoPeriod(null)
    setPastedTotal(null)
    keepDealRef.current = false
    setStarted(false)
  }

  if (success) {
    return <SuccessView bookingNo={success.bookingNo} totalPrice={success.totalPrice} onReset={handleReset} />
  }

  const header = (
    <div className="text-center">
      <h1 className="text-lg font-semibold text-gray-900">บันทึกการจอง</h1>
      <p className="text-xs text-gray-500">โดย {staffName}</p>
    </div>
  )

  if (!started) {
    return (
      <div className="space-y-6 px-4 pb-8 pt-4">
        {header}
        <SummaryPasteStart onApply={applySummary} onDone={() => setStarted(true)} />
      </div>
    )
  }

  return (
    <div className="space-y-6 px-4 pb-8 pt-4">
      {header}

      <SummaryPaste onApply={applySummary} report={pasteReport} />

      {/* Sales owner */}
      <section>
        <Label htmlFor="sales-owner">เซลล์เจ้าของการจอง</Label>
        <Select
          id="sales-owner"
          value={salesUserId}
          onChange={setSalesUserId}
          disabled={salesUsers === null}
          placeholder={salesUsers === null && !salesUsersError ? 'กำลังโหลด...' : 'เลือกเซลล์'}
          options={(salesUsers ?? []).map((u) => ({ value: u.id, label: u.name }))}
        />
        {salesUsersError && <p className="mt-1 text-xs text-red-600">{salesUsersError}</p>}
        {salesUsers?.length === 0 && <p className="mt-1 text-xs text-gray-500">ยังไม่มีผู้ใช้บทบาทเซลล์ในระบบ</p>}
      </section>

      {/* Booking source */}
      <section className="space-y-2">
        <Label htmlFor="booking-source" optional>
          ช่องทางการจอง
        </Label>
        <Select
          id="booking-source"
          value={bookingSourceId}
          onChange={setBookingSourceId}
          disabled={sources === null}
          placeholder={sources === null ? 'กำลังโหลด...' : 'ไม่ระบุ'}
          options={[
            { value: '', label: 'ไม่ระบุ' },
            ...(sources ?? []).map((s) => ({ value: s.id, label: s.name })),
          ]}
        />
        <div className="grid grid-cols-2 gap-2">
          <input
            value={refName}
            onChange={(e) => setRefName(e.target.value)}
            placeholder={source?.shortName ? `ชื่อ ${source.shortName}` : 'ชื่อลูกค้าในช่องทาง'}
            aria-label="ชื่อลูกค้าในช่องทางการจอง"
            className={inputClass}
          />
          <input
            value={refBookingNo}
            onChange={(e) => setRefBookingNo(e.target.value)}
            placeholder="เลข booking"
            aria-label="เลข booking ของช่องทาง"
            className={inputClass}
          />
        </div>
      </section>

      {/* Tour */}
      <section>
        <Label>โปรแกรมทัวร์</Label>
        <TourSelector
          selected={tour}
          onSelect={(t) => {
            setTour(t)
            setAutoPeriod(null)
            keepDealRef.current = false
          }}
        />
      </section>

      {/* Period */}
      <section>
        <Label>ช่วงเดินทาง</Label>
        <PeriodSelector
          key={pasteSeq}
          tourCode={tour?.code ?? null}
          selected={period}
          onSelect={handlePeriodSelect}
          autoSelect={autoPeriod}
          onAutoSelect={handleAutoPeriod}
        />
      </section>

      <DealSection deal={deal} onChange={setDeal} />

      <PaymentSection deal={deal} totals={totals} onChange={setDeal} />
      {pastedTotal !== null && pastedTotal !== totals.grandTotal && (
        <p className="-mt-4 text-right text-xs text-amber-700">
          ข้อความสรุป {formatPrice(pastedTotal)} บาท ไม่ตรงกับยอดในฟอร์ม ตรวจสอบราคาอีกครั้ง
        </p>
      )}

      {/* Customer info */}
      <section>
        <Label optional>แชท LINE ของลูกค้า</Label>
        <CustomerChatPicker selected={lineContact} onSelect={setLineContact} />
        <p className="mt-1 text-xs text-gray-400">
          เลือกเมื่อลูกค้าคุยกับ OA เพื่อส่งใบเสร็จทาง LINE ได้ ถ้าคุยผ่าน LINE ส่วนตัวไม่ต้องเลือก
        </p>
      </section>

      <section>
        <Label htmlFor="customer-name">ชื่อลูกค้า</Label>
        <input
          id="customer-name"
          type="text"
          value={customerName}
          onChange={(e) => setCustomerName(e.target.value)}
          placeholder="ชื่อ-นามสกุล"
          className={inputClass}
        />
      </section>

      <section>
        <Label htmlFor="customer-phone">เบอร์โทรลูกค้า</Label>
        <input
          id="customer-phone"
          type="tel"
          value={customerPhone}
          onChange={(e) => setCustomerPhone(e.target.value)}
          placeholder="0xx-xxx-xxxx"
          className={inputClass}
        />
      </section>

      <section>
        <Label htmlFor="customer-address" optional>
          ที่อยู่ลูกค้า
        </Label>
        <textarea
          id="customer-address"
          value={customerAddress}
          onChange={(e) => setCustomerAddress(e.target.value)}
          rows={2}
          maxLength={500}
          className={inputClass}
        />
      </section>

      {/* Slip upload */}
      <section>
        <Label>สลิปโอนเงิน</Label>
        <SlipUpload value={slipImage} onChange={setSlipImage} />
      </section>

      {/* Note */}
      <section>
        <Label htmlFor="note">หมายเหตุ</Label>
        <textarea
          id="note"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={3}
          placeholder="รายละเอียดเพิ่มเติม (ถ้ามี)"
          className={inputClass}
        />
      </section>

      {error && <div className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>}

      <button
        type="button"
        onClick={handleSubmit}
        disabled={submitting}
        className="w-full rounded-lg bg-green-600 py-3 text-sm font-semibold text-white transition hover:bg-green-700 disabled:opacity-50"
      >
        {submitting ? 'กำลังบันทึก...' : `บันทึกการจอง${totals.grandTotal > 0 ? ` · ${formatPrice(totals.grandTotal)} บาท` : ''}`}
      </button>
    </div>
  )
}
