type Props = {
  bookingNo: string
  totalPrice: number
  onReset: () => void
}

function formatPrice(n: number): string {
  return n.toLocaleString('th-TH')
}

export function SuccessView({ bookingNo, totalPrice, onReset }: Props) {
  return (
    <div className="flex flex-col items-center px-4 py-12 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
        <svg className="h-8 w-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      </div>

      <h2 className="mb-2 text-xl font-semibold text-gray-900">บันทึกสำเร็จ</h2>

      <p className="mb-6 text-sm text-gray-500">ข้อมูลการจองถูกบันทึกเข้าระบบแล้ว</p>

      <div className="mb-8 w-full max-w-xs rounded-lg bg-gray-50 p-4">
        <div className="mb-2 flex justify-between">
          <span className="text-sm text-gray-500">เลขที่จอง</span>
          <span className="font-mono text-sm font-semibold text-gray-900">{bookingNo}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-sm text-gray-500">ยอดรวม</span>
          <span className="text-sm font-semibold text-gray-900">{formatPrice(totalPrice)} บาท</span>
        </div>
      </div>

      <button
        type="button"
        onClick={onReset}
        className="rounded-lg bg-blue-600 px-6 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700"
      >
        บันทึกรายการใหม่
      </button>
    </div>
  )
}
