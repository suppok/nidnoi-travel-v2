import { BellOutlined, CloudSyncOutlined, CompassOutlined, FileDoneOutlined, DownOutlined, EnvironmentOutlined, GlobalOutlined, HomeOutlined, MessageOutlined, SendOutlined, ShareAltOutlined, ShopOutlined, KeyOutlined, LinkOutlined, LogoutOutlined, TeamOutlined } from '@ant-design/icons';
import { Avatar, Badge, Button, Dropdown, Layout, Menu, type MenuProps, Tag, Tooltip } from 'antd';
import { type ReactNode, useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router';
import { useAuth } from '../auth/useAuth.ts';
import BrandMark from '../components/BrandMark.tsx';
import { env } from '../config/env.ts';
import { type Permission, Permissions } from '../constants/permissions.ts';
import { BookingStates } from '../constants/bookingStates.ts';
import { roleInfo } from '../constants/roles.ts';
import BookingAlertsProvider from '../features/booking/components/BookingAlertsProvider.tsx';
import { useBookingAlerts } from '../features/booking/hooks/useBookingAlerts.ts';
import LineBookingAlertsProvider from '../features/line-booking/components/LineBookingAlertsProvider.tsx';
import { useLineBookingAlerts } from '../features/line-booking/hooks/useLineBookingAlerts.ts';
import LineLinkModal from '../features/line-link/LineLinkModal.tsx';
import { SALES_TOOLS } from '../features/sales-tool/tools.tsx';
import SiteRebuildButton from '../features/site/components/SiteRebuildButton.tsx';
import { initials } from '../lib/format.ts';

type NavItem = { path: string; label: string; icon: ReactNode; permission?: Permission };
type NavGroup = { key: string; label: string; items: NavItem[] };

// Add an entry when its page exists. Items the user can't access are hidden (UX only).
const NAV: NavGroup[] = [
  { key: 'main', label: 'เมนูหลัก', items: [{ path: '/dashboard', label: 'ภาพรวม', icon: <HomeOutlined /> }] },
  {
    key: 'sales',
    label: 'ฝ่ายขาย',
    items: [
      { path: '/bookings', label: 'ใบจอง', icon: <FileDoneOutlined />, permission: Permissions.BookingList },
      { path: '/line-bookings', label: 'ใบจอง LINE', icon: <MessageOutlined />, permission: Permissions.LineBookingList },
      { path: '/tour-programs', label: 'โปรแกรมทัวร์', icon: <CompassOutlined />, permission: Permissions.TourProgramList },
      { path: '/taxonomy/countries', label: 'ประเทศ', icon: <GlobalOutlined />, permission: Permissions.CountryList },
      { path: '/taxonomy/cities', label: 'เมือง', icon: <EnvironmentOutlined />, permission: Permissions.CityList },
      { path: '/wholesales', label: 'โฮลเซลล์', icon: <ShopOutlined />, permission: Permissions.WholesaleList },
      { path: '/airlines', label: 'สายการบิน', icon: <SendOutlined />, permission: Permissions.AirlineList },
      { path: '/booking-sources', label: 'ช่องทางการจอง', icon: <ShareAltOutlined />, permission: Permissions.BookingSourceList },
    ],
  },
  {
    key: 'sales-tools',
    label: 'เครื่องมือฝ่ายขาย',
    items: SALES_TOOLS.map(({ path, label, icon, permission }) => ({ path, label, icon, permission })),
  },
  {
    key: 'system',
    label: 'ระบบ',
    items: [
      { path: '/users', label: 'ผู้ใช้และสิทธิ์', icon: <TeamOutlined />, permission: Permissions.UserList },
      { path: '/site-builds', label: 'ประวัติอัปเดตหน้าเว็บ', icon: <CloudSyncOutlined />, permission: Permissions.SiteBuildList },
    ],
  },
];

export default function AppLayout() {
  return (
    <BookingAlertsProvider>
      <LineBookingAlertsProvider>
        <Shell />
      </LineBookingAlertsProvider>
    </BookingAlertsProvider>
  );
}

/** Header bell: count of bookings nobody has contacted yet. */
function NewBookingsBell() {
  const { newCount } = useBookingAlerts();
  const navigate = useNavigate();
  if (newCount === null) return null;
  return (
    <Tooltip title={newCount > 0 ? `ใบจองใหม่รอติดต่อ ${newCount} รายการ` : 'ไม่มีใบจองใหม่'}>
      <Badge count={newCount} size="small" offset={[-4, 4]}>
        <Button type="text" shape="circle" icon={<BellOutlined />} aria-label="ใบจองใหม่" onClick={() => navigate(`/bookings?state=${BookingStates.New}`)} />
      </Badge>
    </Tooltip>
  );
}

function Shell() {
  const { user, can, logout } = useAuth();
  const { newCount } = useBookingAlerts();
  const { newCount: lineNewCount } = useLineBookingAlerts();
  const location = useLocation();
  const navigate = useNavigate();
  const [lineLinkOpen, setLineLinkOpen] = useState(false);

  const menuItems: MenuProps['items'] = NAV.map((group) => ({
    type: 'group' as const,
    key: group.key,
    label: group.label,
    children: group.items
      .filter((item) => !item.permission || can(item.permission))
      .map((item) => ({
        key: item.path,
        icon: item.icon,
        label: (
          <Link to={item.path} className="flex items-center justify-between gap-2">
            {item.label}
            {item.path === '/bookings' && !!newCount && <Badge count={newCount} size="small" />}
            {item.path === '/line-bookings' && !!lineNewCount && <Badge count={lineNewCount} size="small" />}
          </Link>
        ),
      })),
  })).filter((group) => group.children.length > 0);

  const selected = NAV.flatMap((g) => g.items)
    .map((i) => i.path)
    .filter((p) => location.pathname === p || location.pathname.startsWith(`${p}/`));

  const userMenu: MenuProps['items'] = [
    { key: 'line', icon: <LinkOutlined />, label: user?.line ? 'บัญชี LINE (เชื่อมแล้ว)' : 'เชื่อมบัญชี LINE', onClick: () => setLineLinkOpen(true) },
    { key: 'password', icon: <KeyOutlined />, label: 'เปลี่ยนรหัสผ่าน', onClick: () => navigate('/change-password') },
    { type: 'divider' },
    { key: 'logout', icon: <LogoutOutlined />, label: 'ออกจากระบบ', danger: true, onClick: () => void logout() },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Layout.Sider width={248} breakpoint="lg" collapsedWidth={0}>
        <div className="flex h-full flex-col gap-7 px-4 py-6">
          <Link to="/dashboard" className="px-2 no-underline">
            <BrandMark caption="BACKOFFICE" />
          </Link>

          <Menu mode="inline" selectedKeys={selected} items={menuItems} style={{ borderInlineEnd: 'none' }} />

        </div>
      </Layout.Sider>

      <Layout>
        <Layout.Header className="flex items-center gap-3 border-b border-line">
          {env.APP_ENV !== 'prd' && <Tag color="orange">{env.APP_ENV.toUpperCase()}</Tag>}
          {env.API_MOCK && <Tag color="purple">MOCK API</Tag>}
          <div className="ml-auto flex items-center gap-3">
            <SiteRebuildButton className="max-md:hidden" />
            <NewBookingsBell />
            {user && (
              <Dropdown menu={{ items: userMenu }} trigger={['click']} placement="bottomRight">
                <button
                  type="button"
                  aria-label="เมนูผู้ใช้"
                  className="flex cursor-pointer items-center gap-2.5 rounded-xl border-0 bg-transparent px-2 py-1 text-left font-sans text-ink hover:bg-sky-soft"
                >
                  <Avatar size={32} className="flex-none font-semibold" style={{ background: 'var(--color-sky-strong)', color: 'var(--color-ink)' }}>
                    {initials(user.name)}
                  </Avatar>
                  <span className="flex min-w-0 flex-col leading-tight max-md:hidden">
                    <span className="max-w-[200px] truncate text-sm font-semibold">{user.name}</span>
                    <span className="max-w-[200px] truncate text-xs text-ink-muted" title={user.email}>
                      {user.roles.map((r) => roleInfo(r).label).join(' · ')}
                    </span>
                  </span>
                  <DownOutlined className="text-[10px] text-ink-muted" />
                </button>
              </Dropdown>
            )}
          </div>
        </Layout.Header>
        <Layout.Content>
          <div className="box-border flex w-full flex-col gap-6 p-8 max-md:p-4">
            <Outlet />
          </div>
        </Layout.Content>
      </Layout>
      <LineLinkModal open={lineLinkOpen} onClose={() => setLineLinkOpen(false)} />
    </Layout>
  );
}
