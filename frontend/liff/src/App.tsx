import { BookingForm } from './components/BookingForm.tsx'
import { useLiff } from './hooks/useLiff.ts'
import { type StaffBlock, useStaff } from './hooks/useStaff.ts'

const HOW_TO_LINK = 'เข้าระบบหลังบ้าน → เมนูชื่อผู้ใช้มุมขวาบน → “เชื่อมบัญชี LINE” แล้วสแกน QR code ด้วยมือถือ'

const BLOCK_TEXT: Record<StaffBlock, { title: string; body: string }> = {
  'not-linked': { title: 'บัญชี LINE นี้ยังไม่ได้เชื่อมกับผู้ใช้ระบบ', body: HOW_TO_LINK },
  'no-permission': { title: 'บัญชีของคุณไม่มีสิทธิ์ส่งใบจอง', body: 'ใช้ได้เฉพาะฝ่ายขายและผู้ดูแลระบบ' },
  'password-change': { title: 'กรุณาเปลี่ยนรหัสผ่านก่อน', body: 'เข้าระบบหลังบ้านเพื่อตั้งรหัสผ่านใหม่ แล้วเปิดแอปนี้อีกครั้ง' },
  'link-invalid': { title: 'QR code หมดอายุหรือถูกใช้ไปแล้ว', body: `สร้าง QR code ใหม่: ${HOW_TO_LINK}` },
  'already-linked': { title: 'บัญชี LINE นี้เชื่อมกับผู้ใช้อื่นอยู่แล้ว', body: 'ให้ผู้ใช้คนนั้นยกเลิกการเชื่อมในระบบหลังบ้านก่อน' },
  error: { title: 'เกิดข้อผิดพลาด', body: 'กรุณาลองใหม่อีกครั้ง' },
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center px-6">
      <div className="max-w-sm text-center">{children}</div>
    </div>
  )
}

export default function App() {
  const liff = useLiff()
  const ready = liff.status === 'ready'
  const staff = useStaff(ready)

  if (liff.status === 'loading' || (ready && staff.status === 'loading')) {
    return (
      <Centered>
        <p className="text-sm text-gray-400">กำลังเชื่อมต่อ LINE...</p>
      </Centered>
    )
  }

  if (liff.status === 'error') {
    return (
      <Centered>
        <p className="mb-2 text-sm font-medium text-red-600">เกิดข้อผิดพลาด</p>
        <p className="text-xs text-gray-500">{liff.message}</p>
      </Centered>
    )
  }

  if (staff.status === 'blocked') {
    const text = BLOCK_TEXT[staff.reason]
    return (
      <Centered>
        <p className="mb-2 text-sm font-medium text-gray-900">{text.title}</p>
        <p className="text-xs text-gray-500">{staff.message ?? text.body}</p>
        {liff.status === 'ready' && <p className="mt-4 text-xs text-gray-400">LINE: {liff.displayName}</p>}
      </Centered>
    )
  }

  if (staff.status !== 'ready') return null

  return (
    <div className="mx-auto min-h-screen max-w-lg bg-white">
      {staff.justLinked && (
        <p className="bg-green-50 px-4 py-2 text-center text-sm text-green-700">
          เชื่อมบัญชี LINE กับ {staff.me.name} แล้ว
        </p>
      )}
      <BookingForm staffName={staff.me.name} />
    </div>
  )
}
