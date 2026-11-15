import { useRef, useState } from 'react'
import { ApiError, getEditableBooking, listBookingsByChat } from '../api/client.ts'
import { formatPrice, inputClass } from '../lib/format.ts'
import type { LineBookingItem, LineContact } from '../types/index.ts'
import { CustomerChatPicker } from './CustomerChatPicker.tsx'

const ERRORS: Record<string, string> = {
  NOT_FOUND: 'ไม่พบใบจองเลขนี้',
  LINE_BOOKING_NOT_EDITABLE: 'แก้ไขได้เฉพาะใบจองที่เซลล์ยืนยันการชำระเงินแล้ว (ชำระมัดจำหรือชำระครบ)',
}

const shortDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit', timeZone: 'UTC' })

function dateRange(b: LineBookingItem): string {
  return `${shortDate.format(new Date(`${b.startDate}T00:00:00Z`))} – ${shortDate.format(new Date(`${b.endDate}T00:00:00Z`))}`
}

function stateLabel(b: LineBookingItem): string {
  if (b.state === 25) return 'ชำระครบแล้ว'
  if (b.balanceSubmittedAt && !b.balanceConfirmedAt) return 'รอยืนยันยอดส่วนที่เหลือ'
  return 'ชำระมัดจำแล้ว'
}

// Reopens an earlier paid booking, to change the deal (e.g. travellers) or, on a deposit booking,
// to record the balance the customer paid. Found by its booking no, or by the customer's chat when
// the booking was linked to it.
export function OpenBooking({ onOpen }: { onOpen: (booking: LineBookingItem) => void }) {
  const [bookingNo, setBookingNo] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [contact, setContact] = useState<LineContact | null>(null)
  const [chatBookings, setChatBookings] = useState<LineBookingItem[] | null>(null)
  const [chatLoading, setChatLoading] = useState(false)
  const [chatError, setChatError] = useState('')

  const chatSeq = useRef(0)

  async function selectContact(c: LineContact | null) {
    const current = ++chatSeq.current
    setContact(c)
    setChatBookings(null)
    setChatError('')
    if (!c) return
    setChatLoading(true)
    try {
      const items = await listBookingsByChat(c.id)
      if (current === chatSeq.current) setChatBookings(items)
    } catch {
      if (current === chatSeq.current) setChatError('โหลดใบจองไม่สำเร็จ กรุณาลองอีกครั้ง')
    } finally {
      if (current === chatSeq.current) setChatLoading(false)
    }
  }

  async function open() {
    const no = bookingNo.trim().toUpperCase()
    if (!no) return
    setError('')
    setLoading(true)
    try {
      onOpen(await getEditableBooking(no))
    } catch (err) {
      setError(err instanceof ApiError ? (ERRORS[err.code] ?? err.message) : 'เกิดข้อผิดพลาด กรุณาลองอีกครั้ง')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-2 border-t border-gray-200 pt-5">
      <label htmlFor="open-booking" className="block text-sm font-medium text-gray-700">
        แก้ไขใบจองเดิม / ลูกค้าชำระส่วนที่เหลือ
      </label>
      <p className="text-xs text-gray-500">กรอกเลขที่จองของใบจองที่ยืนยันการชำระเงินแล้ว (มัดจำหรือชำระเต็มจำนวน)</p>
      <form
        className="flex gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          void open()
        }}
      >
        <input
          id="open-booking"
          value={bookingNo}
          onChange={(e) => setBookingNo(e.target.value)}
          placeholder="เช่น 26X8K9AQ"
          autoCapitalize="characters"
          autoComplete="off"
          maxLength={32}
          className={`${inputClass} font-mono uppercase`}
        />
        <button
          type="submit"
          disabled={loading || !bookingNo.trim()}
          className="shrink-0 rounded-lg bg-gray-800 px-4 text-sm font-medium text-white disabled:opacity-50"
        >
          {loading ? 'กำลังเปิด...' : 'เปิดใบจอง'}
        </button>
      </form>
      {error && <p className="text-xs text-red-600">{error}</p>}

      <p className="pt-2 text-xs text-gray-500">หรือค้นหาจากแชทลูกค้า (ใบจองที่ผูกแชทไว้ตอนสร้าง)</p>
      <CustomerChatPicker selected={contact} onSelect={(c) => void selectContact(c)} />
      {chatLoading && <p className="text-xs text-gray-400">กำลังโหลดใบจอง...</p>}
      {chatError && <p className="text-xs text-red-600">{chatError}</p>}
      {chatBookings && chatBookings.length === 0 && (
        <p className="text-xs text-gray-500">ไม่พบใบจองที่ชำระเงินแล้วของแชทนี้ — ลองกรอกเลขที่จองแทน</p>
      )}
      {chatBookings && chatBookings.length > 0 && (
        <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200">
          {chatBookings.map((b) => (
            <li key={b.id}>
              <button
                type="button"
                onClick={() => onOpen(b)}
                className="flex w-full items-start gap-3 px-3 py-2.5 text-left hover:bg-blue-50"
              >
                <span className="min-w-0 flex-1">
                  <span className="block font-mono text-sm font-medium text-gray-900">{b.bookingNo}</span>
                  <span className="block truncate text-xs text-gray-700">
                    {b.tourCode} {b.tourName}
                  </span>
                  <span className="block text-xs text-gray-500">
                    {dateRange(b)} · {b.customerName}
                  </span>
                </span>
                <span className="shrink-0 text-right">
                  <span className="block text-xs text-gray-700">{formatPrice(b.totalPrice)} บาท</span>
                  <span className={`block text-xs ${b.state === 25 ? 'text-green-700' : 'text-amber-700'}`}>{stateLabel(b)}</span>
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
