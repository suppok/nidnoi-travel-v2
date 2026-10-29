import { Tag } from 'antd';
import { Outlet } from 'react-router';
import BrandMark from '../components/BrandMark.tsx';
import { env } from '../config/env.ts';

// /login, /change-password — no sidebar.
export default function AuthLayout() {
  return (
    <div
      className="flex min-h-full flex-col items-center justify-center gap-8 bg-sky px-5 py-12"
      style={{
        backgroundImage: 'radial-gradient(var(--color-sky-dot) 1.6px, transparent 1.7px)',
        backgroundSize: '24px 24px',
      }}
    >
      <div className="flex items-center gap-3">
        <BrandMark size="lg" />
        {env.APP_ENV !== 'prd' && <Tag color="orange">{env.APP_ENV.toUpperCase()}</Tag>}
      </div>
      <Outlet />
      <p className="m-0 text-[13px] text-ink-muted">© nidnoi travel · สำหรับทีมงานเท่านั้น</p>
    </div>
  );
}
