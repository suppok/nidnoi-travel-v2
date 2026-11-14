import { useEffect, useRef, useState } from 'react'
import { ApiError, createLineBooking, getBookingSources, getSalesUsers, searchTours, updateLineBooking } from '../api/client.ts'
import { applyPeriod, computeTotals, type Deal, dealError, dealFromBooking, dealFromSummary, dealRequest, emptyDeal } from '../lib/deal.ts'
import type { ParsedSummary } from '../lib/parseSummary.ts'
import type { BookingSource, LineBookingItem, LineContact, SalesUser, TourPeriodItem, TourPublicItem } from '../types/index.ts'
import { CustomerChatPicker } from './CustomerChatPicker.tsx'
import { BookingHistory } from './BookingHistory.tsx'
import { DealSection } from './DealSection.tsx'
import { OpenBooking } from './OpenBooking.tsx'
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

type SuccessData = { bookingNo: string; totalPrice: number; message?: string }

type Purpose = 'balance' | 'edit'

// line_booking.state 20: deposit paid, the balance may still be recorded. MUST match the backend.
const DEPOSIT_PAID = 20

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
  // A booking reopened by its booking no (edit mode), and the balance slip to record with it.
  const [editing, setEditing] = useState<LineBookingItem | null>(null)
  const [balanceSlip, setBalanceSlip] = useState('')
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
  // What sales reopened the booking for: record the balance, or edit it. Asked only when the
  // booking can take a balance; otherwise it is an edit.
  const [purpose, setPurpose] = useState<Purpose | null>(null)
  const balanceMode = editing !== null && purpose === 'balance'
  // Finance hasn't finished this round's documents: saving an edit asks for a second confirm.
  const waitingFinance = editing !== null && editing.documentsApprovedAt === null
  const [confirmingWarning, setConfirmingWarning] = useState(false)
  // A source deleted since the booking was made still shows on it.
  const sourceOptions = [...(sources ?? [])]
  if (editing?.bookingSource && !sourceOptions.some((s) => s.id === editing.bookingSource?.id)) sourceOptions.push(editing.bookingSource)

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

  function openBooking(b: LineBookingItem) {
    setEditing(b)
    setDeal(dealFromBooking(b))
    setBookingSourceId(b.bookingSource?.id ?? '')
    setRefName(b.refName)
    setRefBookingNo(b.refBookingNo)
    setLineContact(b.lineContact ? { ...b.lineContact, lastMessageAt: null } : null)
    setCustomerName(b.customerName)
    setCustomerPhone(b.customerPhone)
    setCustomerAddress(b.customerAddress)
    setNote(b.note)
    setBalanceSlip('')
    setError('')
    setConfirmingWarning(false)
    setPurpose(canRecordBalance(b) ? null : 'edit')
    setStarted(true)
  }

  async function handleUpdate(b: LineBookingItem) {
    setError('')
    if (balanceMode) {
      if (!balanceSlip) return setError('กรุณาอัพโหลดสลิปชำระส่วนที่เหลือ')
    } else {
      const invalid = dealError(deal)
      if (invalid) return setError(invalid)
      if (!customerName.trim()) return setError('กรุณาระบุชื่อลูกค้า')
      if (!customerPhone.trim()) return setError('กรุณาระบุเบอร์โทรลูกค้า')
      if (waitingFinance && !confirmingWarning) return setConfirmingWarning(true)
    }

    setConfirmingWarning(false)
    setSubmitting(true)
    try {
      const result = await updateLineBooking({
        bookingNo: b.bookingNo,
        lineContactId: lineContact?.id ?? null,
        bookingSourceId: bookingSourceId || null,
        refName: refName.trim(),
        refBookingNo: refBookingNo.trim(),
        ...dealRequest(deal),
        customerName: customerName.trim(),
        customerPhone: customerPhone.trim(),
        customerAddress: customerAddress.trim(),
        note: note.trim(),
        balanceSlipImage: balanceMode ? balanceSlip : '',
      })
      setSuccess({
        bookingNo: result.bookingNo,
        totalPrice: result.totalPrice,
        message: balanceMode
          ? `บันทึกยอดคงเหลือแล้ว รอ ${b.salesUser?.name ?? 'เซลล์เจ้าของการจอง'} ยืนยันในระบบหลังบ้าน แล้วฝ่ายการเงินจะออกใบเสร็จ`
          : 'บันทึกการแก้ไขใบจองแล้ว',
      })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'เกิดข้อผิดพลาด กรุณาลองอีกครั้ง')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleSubmit() {
    if (editing) return handleUpdate(editing)
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
    setEditing(null)
    setBalanceSlip('')
    setConfirmingWarning(false)
    setPurpose(null)
    keepDealRef.current = false
    setStarted(false)
  }

  if (success) {
    return <SuccessView bookingNo={success.bookingNo} totalPrice={success.totalPrice} message={success.message} onReset={handleReset} />
  }

  const header = (
    <div className="text-center">
      <h1 className="text-lg font-semibold text-gray-900">
        {editing ? `${balanceMode ? 'ชำระส่วนที่เหลือ' : 'แก้ไขใบจอง'} ${editing.bookingNo}` : 'บันทึกการจอง'}
      </h1>
      <p className="text-xs text-gray-500">โดย {staffName}</p>
    </div>
  )

  if (!started) {
    return (
      <div className="space-y-6 px-4 pb-8 pt-4">
        {header}
        <SummaryPasteStart onApply={applySummary} onDone={() => setStarted(true)} />
        <OpenBooking onOpen={openBooking} />
      </div>
    )
  }

  // A reopened booking that can take a balance: first choose what to do.
  if (editing && purpose === null) {
    return (
      <div className="space-y-6 px-4 pb-8 pt-4">
        {header}
        <EditingSummary booking={editing} onCancel={handleReset} />
        <PurposeChoice remaining={totals.remaining} onChoose={setPurpose} />
      </div>
    )
  }

  // Recording the balance: only the slip (and a note); the deal stays as it is.
  if (editing && balanceMode) {
    return (
      <div className="space-y-6 px-4 pb-8 pt-4">
        {header}
        <EditingSummary booking={editing} onCancel={handleReset} onChangePurpose={() => setPurpose(null)} />

        <section className="rounded-xl border border-green-200 bg-green-50 px-4 py-3">
          <p className="text-sm text-green-800">ยอดคงเหลือที่ลูกค้าต้องชำระ</p>
          <p className="text-2xl font-semibold text-green-900">{formatPrice(totals.remaining)} บาท</p>
          {editing.remainingDueDate && <p className="text-xs text-green-800">กำหนดชำระภายใน {thaiDate.format(new Date(`${editing.remainingDueDate}T00:00:00Z`))}</p>}
        </section>

        <section>
          <Label>สลิปชำระส่วนที่เหลือ</Label>
          {editing.balanceSubmittedAt && !balanceSlip && (
            <p className="mb-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800">
              มีสลิปส่วนที่เหลือรอเซลล์ยืนยันอยู่แล้ว อัพโหลดใหม่เพื่อแทนที่{' '}
              <a href={editing.balanceSlipImage} target="_blank" rel="noreferrer" className="underline">
                ดูสลิปเดิม
              </a>
            </p>
          )}
          <SlipUpload value={balanceSlip} onChange={setBalanceSlip} />
          <p className="mt-1 text-xs text-gray-500">{editing.salesUser?.name ?? 'เซลล์เจ้าของการจอง'} จะตรวจสลิปและยืนยันในระบบหลังบ้าน</p>
        </section>

        <section>
          <Label htmlFor="note">หมายเหตุ</Label>
          <textarea id="note" value={note} onChange={(e) => setNote(e.target.value)} rows={3} className={inputClass} />
        </section>

        {error && <div className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>}

        <button
          type="button"
          onClick={handleSubmit}
          disabled={submitting || !balanceSlip}
          className="w-full rounded-lg bg-green-600 py-3 text-sm font-semibold text-white transition hover:bg-green-700 disabled:opacity-50"
        >
          {submitting ? 'กำลังบันทึก...' : 'ส่งยอดคงเหลือให้เซลล์ยืนยัน'}
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6 px-4 pb-8 pt-4">
      {header}

      {editing ? (
        <EditingSummary
          booking={editing}
          onCancel={handleReset}
          onChangePurpose={canRecordBalance(editing) ? () => setPurpose(null) : undefined}
        />
      ) : (
        <SummaryPaste onApply={applySummary} report={pasteReport} />
      )}

      {/* Sales owner (fixed on a reopened booking) */}
      {!editing && (
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
      )}

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
            ...sourceOptions.map((s) => ({ value: s.id, label: s.name })),
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

      {/* Tour and period are fixed on a reopened booking (shown in its summary above). */}
      {!editing && (
        <>
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
        </>
      )}

      <DealSection deal={deal} onChange={setDeal} />

      <PaymentSection deal={deal} totals={totals} onChange={setDeal} depositLocked={editing !== null} />
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

      {/* Slip upload (a new booking only; an edit takes no slip, the balance has its own screen) */}
      {!editing && (
        <section>
          <Label>สลิปโอนเงิน</Label>
          <SlipUpload value={slipImage} onChange={setSlipImage} />
        </section>
      )}

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

      {editing && confirmingWarning && (
        <div className="space-y-3 rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900">
          <p className="font-medium">{financeWarning(editing)}</p>
          <p className="text-xs">แจ้งฝ่ายการเงินด้วยว่ามีการแก้ไขใบจอง เพื่อออกเอกสารตามยอดใหม่</p>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setConfirmingWarning(false)}
              disabled={submitting}
              className="flex-1 rounded-lg border border-gray-300 bg-white py-2 text-sm text-gray-700 disabled:opacity-50"
            >
              กลับไปแก้ไข
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting}
              className="flex-1 rounded-lg bg-amber-600 py-2 text-sm font-semibold text-white disabled:opacity-50"
            >
              {submitting ? 'กำลังบันทึก...' : 'ยืนยันบันทึก'}
            </button>
          </div>
        </div>
      )}

      <button
        type="button"
        onClick={handleSubmit}
        disabled={submitting || confirmingWarning}
        className="w-full rounded-lg bg-green-600 py-3 text-sm font-semibold text-white transition hover:bg-green-700 disabled:opacity-50"
      >
        {submitting
          ? 'กำลังบันทึก...'
          : `${editing ? 'บันทึกการแก้ไข' : 'บันทึกการจอง'}${totals.grandTotal > 0 ? ` · ${formatPrice(totals.grandTotal)} บาท` : ''}`}
      </button>
    </div>
  )
}

// The fixed part of a reopened booking: tour, departure, sales owner, payment and the slips.
const thaiDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'UTC' })

// "30 พ.ย. 69 – 2 ธ.ค. 69 (3 วัน)" from YYYY-MM-DD dates.
function tripRange(start: string, end: string): string {
  const s = new Date(`${start.slice(0, 10)}T00:00:00Z`)
  const e = new Date(`${end.slice(0, 10)}T00:00:00Z`)
  const days = Math.round((e.getTime() - s.getTime()) / 86_400_000) + 1
  return `${thaiDate.format(s)} – ${thaiDate.format(e)} (${days} วัน)`
}

function SummaryRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-3 py-1.5">
      <dt className="shrink-0 text-gray-500">{label}</dt>
      <dd className="text-right font-medium text-gray-900">{children}</dd>
    </div>
  )
}

function SlipLink({ href, label }: { href: string; label: string }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="inline-flex items-center gap-1.5 rounded-full border border-blue-200 bg-white px-3 py-1 text-xs font-medium text-blue-700"
    >
      <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.6-4.6a2 2 0 012.8 0L16 16m-2-2l1.6-1.6a2 2 0 012.8 0L20 14M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2zm8-12h.01" />
      </svg>
      {label}
    </a>
  )
}

function EditingSummary({ booking, onCancel, onChangePurpose }: { booking: LineBookingItem; onCancel: () => void; onChangePurpose?: () => void }) {
  const depositPaid = booking.state === DEPOSIT_PAID
  return (
    <section className="overflow-hidden rounded-xl border border-blue-200 bg-blue-50/60 text-sm">
      <div className="space-y-1.5 border-b border-blue-100 bg-white px-4 py-3">
        <div className="flex items-center justify-between gap-2">
          <span className="rounded bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-700">{booking.tourCode}</span>
          <span
            className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${depositPaid ? 'bg-amber-100 text-amber-800' : 'bg-green-100 text-green-800'}`}
          >
            {depositPaid ? 'ชำระมัดจำแล้ว' : 'ชำระครบแล้ว'}
          </span>
        </div>
        <p className="line-clamp-2 font-medium leading-snug text-gray-900">{booking.tourName}</p>
      </div>

      <dl className="divide-y divide-blue-100 px-4 py-1">
        <SummaryRow label="วันเดินทาง">{tripRange(booking.startDate, booking.endDate)}</SummaryRow>
        <SummaryRow label="เซลล์">{booking.salesUser?.name ?? '-'}</SummaryRow>
        <SummaryRow label="ยอดเดิม">{formatPrice(booking.totalPrice)} บาท</SummaryRow>
      </dl>

      <div className="flex flex-wrap gap-2 px-4 pb-3">
        <SlipLink href={booking.slipImage} label={booking.isDeposit ? 'สลิปมัดจำ' : 'สลิปโอนเงิน'} />
        {booking.balanceConfirmedAt && booking.balanceSlipImage && <SlipLink href={booking.balanceSlipImage} label="สลิปส่วนที่เหลือ" />}
      </div>

      {!booking.documentsApprovedAt && (
        <p className="mx-4 mb-3 rounded-lg bg-amber-100 px-3 py-2 text-xs leading-relaxed text-amber-900">⚠️ {financeWarning(booking)}</p>
      )}

      <BookingHistory bookingNo={booking.bookingNo} />

      <div className="flex justify-between gap-3 border-t border-blue-100 bg-white px-4 py-2">
        {onChangePurpose ? (
          <button type="button" onClick={onChangePurpose} className="text-xs text-blue-700 underline">
            เปลี่ยนสิ่งที่ต้องการทำ
          </button>
        ) : (
          <span />
        )}
        <button type="button" onClick={onCancel} className="text-xs text-gray-500 underline">
          ยกเลิก
        </button>
      </div>
    </section>
  )
}

// Why saving a booking whose documents aren't approved yet needs care: finance may issue them for
// the old amounts.
function financeWarning(b: LineBookingItem): string {
  const finance = b.financeUser?.name ?? 'ฝ่ายการเงิน'
  return b.documentsAt
    ? `${finance} ส่งเอกสารการเงินแล้วแต่ยังรอเซลล์ตรวจ เอกสารอาจไม่ตรงกับข้อมูลที่แก้ไข`
    : `ใบจองนี้รอ ${finance} ออกเอกสารการเงิน การแก้ไขตอนนี้อาจทำให้เอกสารไม่ตรงกับยอดใหม่`
}

// A deposit booking not yet paid in full whose deposit documents were sent can take the balance.
function canRecordBalance(b: LineBookingItem): boolean {
  return b.isDeposit && b.state === DEPOSIT_PAID && b.documentsApprovedAt !== null
}

function PurposeChoice({ remaining, onChoose }: { remaining: number; onChoose: (p: Purpose) => void }) {
  const card = 'w-full rounded-xl border-2 px-4 py-4 text-left transition active:scale-[0.99]'
  return (
    <section className="space-y-3">
      <p className="text-sm font-medium text-gray-700">ต้องการทำอะไรกับใบจองนี้?</p>
      <button type="button" onClick={() => onChoose('balance')} className={`${card} border-green-300 bg-green-50 hover:border-green-500`}>
        <p className="font-semibold text-green-900">💰 ลูกค้าชำระส่วนที่เหลือ</p>
        <p className="mt-1 text-xs text-green-800">แนบสลิปยอดคงเหลือ {formatPrice(remaining)} บาท ให้เซลล์ยืนยัน แล้วฝ่ายการเงินออกใบเสร็จ</p>
      </button>
      <button type="button" onClick={() => onChoose('edit')} className={`${card} border-gray-200 bg-white hover:border-blue-400`}>
        <p className="font-semibold text-gray-900">✏️ แก้ไขข้อมูลใบจอง</p>
        <p className="mt-1 text-xs text-gray-500">เช่น เพิ่ม/ลดผู้เดินทาง เปลี่ยนราคา ข้อมูลลูกค้า หรือกำหนดชำระ</p>
      </button>
    </section>
  )
}
