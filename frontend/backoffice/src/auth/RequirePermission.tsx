import { Button, Result } from 'antd';
import { Link, Outlet } from 'react-router';
import type { Permission } from '../constants/permissions.ts';
import { useAuth } from './useAuth.ts';

export default function RequirePermission({ code }: { code: Permission }) {
  const { can } = useAuth();

  if (!can(code)) {
    return (
      <Result
        status="403"
        title="ไม่มีสิทธิ์เข้าถึง"
        subTitle="บทบาทของคุณยังไม่มีสิทธิ์ใช้หน้านี้ ติดต่อผู้ดูแลระบบหากต้องการสิทธิ์เพิ่ม"
        extra={
          <Link to="/dashboard">
            <Button type="primary">กลับหน้าภาพรวม</Button>
          </Link>
        }
      />
    );
  }
  return <Outlet />;
}
