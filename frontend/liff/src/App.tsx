import { BookingForm } from './components/BookingForm.tsx'
import { useLiff } from './hooks/useLiff.ts'

export default function App() {
  const liff = useLiff()

  if (liff.status === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-gray-400">กำลังเชื่อมต่อ LINE...</p>
      </div>
    )
  }

  if (liff.status === 'error') {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <div className="text-center">
          <p className="mb-2 text-sm font-medium text-red-600">เกิดข้อผิดพลาด</p>
          <p className="text-xs text-gray-500">{liff.message}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto min-h-screen max-w-lg bg-white">
      <BookingForm
        userId={liff.userId}
        displayName={liff.displayName}
        accessToken={liff.accessToken}
      />
    </div>
  )
}
