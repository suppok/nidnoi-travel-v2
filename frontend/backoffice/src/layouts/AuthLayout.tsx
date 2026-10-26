import { Tag } from 'antd';
import { Outlet } from 'react-router';
import { env } from '../config/env.ts';

// /login — no sidebar.
export default function AuthLayout() {
  return (
    <div className="flex min-h-full flex-col items-center justify-center gap-4 bg-slate-100 px-4">
      {env.APP_ENV !== 'prd' && <Tag color="orange">{env.APP_ENV.toUpperCase()}</Tag>}
      <Outlet />
    </div>
  );
}
