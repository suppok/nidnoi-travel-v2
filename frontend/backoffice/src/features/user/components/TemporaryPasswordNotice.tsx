import { Alert, Typography } from 'antd';

// Shown once after create/reset so the admin can hand the temporary password over.
export default function TemporaryPasswordNotice({ email, password }: { email: string; password: string }) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <span className="text-ink-muted">อีเมล</span>
        <Typography.Text copyable strong>
          {email}
        </Typography.Text>
      </div>
      <div className="flex flex-col gap-1">
        <span className="text-ink-muted">รหัสผ่านชั่วคราว</span>
        <Typography.Text copyable code className="text-base">
          {password}
        </Typography.Text>
      </div>
      <Alert
        type="warning"
        showIcon
        title="ส่งรหัสผ่านนี้ให้ผู้ใช้ทางช่องทางส่วนตัว หน้านี้จะไม่แสดงรหัสผ่านอีก ผู้ใช้ต้องตั้งรหัสผ่านใหม่เมื่อเข้าสู่ระบบครั้งแรก"
      />
    </div>
  );
}
