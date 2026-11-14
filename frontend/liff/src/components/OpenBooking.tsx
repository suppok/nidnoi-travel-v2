import { useState } from 'react'
import { ApiError, getEditableBooking } from '../api/client.ts'
import { inputClass } from '../lib/format.ts'
import type { LineBookingItem } from '../types/index.ts'

const ERRORS: Record<string, string> = {
  NOT_FOUND: 'ไม่พบใบจองเลขนี้',
  LINE_BOOKING_NOT_EDITABLE: 'แก้ไขได้เฉพาะใบจองที่เซลล์ยืนยันการชำระเงินแล้ว (ชำระมัดจำหรือชำระครบ)',
}

// Reopens an earlier paid booking by its booking no, to change the deal (e.g. travellers) or, on a
// deposit booking, to record the balance the customer paid.
export function OpenBooking({ onOpen }: { onOpen: (booking: LineBookingItem) => void }) {
  const [bookingNo, setBookingNo] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

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
    </div>
  )
}
