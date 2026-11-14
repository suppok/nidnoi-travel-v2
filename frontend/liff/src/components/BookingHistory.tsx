import { useEffect, useState } from 'react'
import { getBookingHistory } from '../api/client.ts'
import type { LineBookingEvent } from '../types/index.ts'

// Labels mirror the backoffice (frontend/backoffice/src/constants/lineBookingEvents.ts); actions and
// field keys MUST match backend pkg/constants/linebookingevent and dealSnapshot.
const FIELD_LABELS: Record<string, string> = {
  adultTwin: 'ผู้ใหญ่ (พักคู่)',
  adultSingle: 'ผู้ใหญ่ (พักเดี่ยว)',
  childWithBed: 'เด็ก',
  childNoBed: 'เด็ก (ไม่เสริมเตียง)',
  infant: 'ทารก',
  adultPrice: 'ราคาผู้ใหญ่',
  adultSinglePrice: 'ราคาผู้ใหญ่พักเดี่ยว',
  childWithBedPrice: 'ราคาเด็ก',
  childNoBedPrice: 'ราคาเด็ก (ไม่เสริมเตียง)',
  infantPrice: 'ราคาทารก',
  adultPriceBeforeDiscount: 'ราคาก่อนลด',
  discountPerPerson: 'ส่วนลด/ท่าน',
  charges: 'ค่าใช้จ่ายเพิ่ม',
  paymentMethod: 'วิธีชำระ',
  creditFeePercent: 'ค่าธรรมเนียมบัตร (%)',
  depositPerPerson: 'มัดจำ/ท่าน',
  depositDueDate: 'ชำระมัดจำภายใน',
  depositDueTime: 'เวลาชำระมัดจำ',
  remainingDueDate: 'ชำระส่วนเหลือภายใน',
  fullDueDate: 'ชำระเต็มจำนวนภายใน',
  fullDueTime: 'เวลาชำระเต็มจำนวน',
  totalPrice: 'ยอดรวม',
  customerName: 'ชื่อลูกค้า',
  customerPhone: 'เบอร์โทร',
  customerAddress: 'ที่อยู่',
  bookingSource: 'ช่องทางการจอง',
  refName: 'ชื่อในช่องทาง',
  refBookingNo: 'เลข booking ช่องทาง',
  lineContact: 'แชท LINE ลูกค้า',
  note: 'หมายเหตุ',
  staffNote: 'บันทึกของเจ้าหน้าที่',
}

const MONEY_FIELDS = new Set([
  'adultPrice',
  'adultSinglePrice',
  'childWithBedPrice',
  'childNoBedPrice',
  'infantPrice',
  'adultPriceBeforeDiscount',
  'discountPerPerson',
  'depositPerPerson',
  'totalPrice',
])

const DOT: Record<string, string> = {
  created: 'bg-blue-500',
  confirmed: 'bg-green-500',
  cancelled: 'bg-red-500',
  documents_sent: 'bg-blue-500',
  documents_approved: 'bg-green-500',
  documents_rejected: 'bg-red-500',
  edited: 'bg-orange-500',
  balance_submitted: 'bg-yellow-500',
  balance_confirmed: 'bg-green-500',
  comment: 'bg-purple-500',
  deleted: 'bg-red-500',
}

function title(e: LineBookingEvent): string {
  switch (e.action) {
    case 'created':
      return 'ส่งใบจองจาก LINE'
    case 'confirmed':
      return `ยืนยันการชำระเงิน · ส่งให้ ${e.target || 'ฝ่ายการเงิน'}`
    case 'cancelled':
      return 'ยกเลิกใบจอง'
    case 'documents_sent':
      return `ฝ่ายการเงินส่งเอกสาร ${e.target}`
    case 'documents_updated':
      return `ฝ่ายการเงินเปลี่ยนไฟล์เอกสาร ${e.target}`
    case 'documents_approved':
      return e.target ? `ส่งเอกสาร ${e.note} ให้ลูกค้าทาง LINE (${e.target})` : `ยืนยันเอกสาร ${e.note} (ไม่ได้ส่ง LINE)`
    case 'documents_rejected':
      return `ส่งเอกสาร ${e.target} กลับให้ฝ่ายการเงินแก้ไข`
    case 'line_contact_set':
      return 'เปลี่ยนแชท LINE ของลูกค้า'
    case 'staff_note_updated':
      return 'แก้บันทึกของเจ้าหน้าที่'
    case 'edited':
      return 'แก้ไขใบจอง'
    case 'balance_submitted':
      return 'แนบสลิปชำระส่วนที่เหลือ'
    case 'balance_confirmed':
      return `ยืนยันยอดคงเหลือ · ชำระครบแล้ว`
    case 'deleted':
      return 'ลบใบจอง'
    case 'restored':
      return 'กู้คืนใบจอง'
    case 'comment':
      return 'ความคิดเห็น'
    default:
      return e.action
  }
}

function value(field: string, v: string): string {
  if (v === '') return '–'
  if (MONEY_FIELDS.has(field) && /^-?\d+$/.test(v)) return Number(v).toLocaleString('th-TH')
  if (field === 'paymentMethod') return v === 'credit' ? 'บัตรเครดิต' : 'โอนเงิน'
  return v
}

const when = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })

// "ประวัติ" of a reopened booking, newest first, collapsed until tapped.
export function BookingHistory({ bookingNo }: { bookingNo: string }) {
  const [events, setEvents] = useState<LineBookingEvent[] | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let cancelled = false
    getBookingHistory(bookingNo)
      .then((items) => !cancelled && setEvents(items))
      .catch(() => !cancelled && setFailed(true))
    return () => {
      cancelled = true
    }
  }, [bookingNo])

  const newestFirst = events ? [...events].reverse() : []

  return (
    <details className="group border-t border-blue-100 bg-white px-4 py-2">
      <summary className="flex cursor-pointer list-none items-center justify-between text-xs font-medium text-gray-700">
        <span>ประวัติ{events ? ` (${events.length})` : ''}</span>
        <span className="text-gray-400 transition group-open:rotate-180">▾</span>
      </summary>
      <div className="pt-2">
        {failed && <p className="text-xs text-red-600">โหลดประวัติไม่สำเร็จ</p>}
        {!failed && events === null && <p className="text-xs text-gray-400">กำลังโหลด...</p>}
        {events?.length === 0 && <p className="text-xs text-gray-400">ยังไม่มีประวัติ</p>}
        <ol className="space-y-3">
          {newestFirst.map((e) => (
            <li key={e.id} className="flex gap-2.5">
              <span className={`mt-1.5 h-2 w-2 shrink-0 rounded-full ${DOT[e.action] ?? 'bg-gray-400'}`} />
              <div className="min-w-0 flex-1 text-xs">
                <p className="font-medium text-gray-900">{title(e)}</p>
                <p className="text-gray-500">
                  {when.format(new Date(e.createDate))} · {e.user?.name ?? '-'}
                </p>
                {e.action === 'documents_rejected' && e.note && <p className="mt-0.5 whitespace-pre-wrap text-red-700">{e.note}</p>}
                {e.action === 'comment' && <p className="mt-1 whitespace-pre-wrap rounded-lg bg-gray-100 px-2.5 py-1.5 text-gray-800">{e.note}</p>}
                {e.changes.length > 0 && (
                  <ul className="mt-1 space-y-0.5">
                    {e.changes.map((c) => (
                      <li key={c.field} className="text-gray-700">
                        <span className="text-gray-500">{FIELD_LABELS[c.field] ?? c.field}:</span>{' '}
                        <span className="text-gray-400 line-through">{value(c.field, c.from)}</span> →{' '}
                        <span className="font-medium">{value(c.field, c.to)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </li>
          ))}
        </ol>
      </div>
    </details>
  )
}
