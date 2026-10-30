import { DownOutlined, EnvironmentOutlined, GlobalOutlined, HomeOutlined, ShopOutlined, KeyOutlined, LogoutOutlined, TeamOutlined } from '@ant-design/icons';
import { Avatar, Dropdown, Layout, Menu, type MenuProps, Tag } from 'antd';
import type { ReactNode } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router';
import { useAuth } from '../auth/useAuth.ts';
import BrandMark from '../components/BrandMark.tsx';
import { env } from '../config/env.ts';
import { type Permission, Permissions } from '../constants/permissions.ts';
import { roleInfo } from '../constants/roles.ts';
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
      { path: '/taxonomy/countries', label: 'ประเทศ', icon: <GlobalOutlined />, permission: Permissions.CountryList },
      { path: '/taxonomy/cities', label: 'เมือง', icon: <EnvironmentOutlined />, permission: Permissions.CityList },
      { path: '/wholesales', label: 'โฮลเซลล์', icon: <ShopOutlined />, permission: Permissions.WholesaleList },
    ],
  },
  {
    key: 'system',
    label: 'ระบบ',
    items: [{ path: '/users', label: 'ผู้ใช้และสิทธิ์', icon: <TeamOutlined />, permission: Permissions.UserList }],
  },
];

export default function AppLayout() {
  const { user, can, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const menuItems: MenuProps['items'] = NAV.map((group) => ({
    type: 'group' as const,
    key: group.key,
    label: group.label,
    children: group.items
      .filter((item) => !item.permission || can(item.permission))
      .map((item) => ({ key: item.path, icon: item.icon, label: <Link to={item.path}>{item.label}</Link> })),
  })).filter((group) => group.children.length > 0);

  const selected = NAV.flatMap((g) => g.items)
    .map((i) => i.path)
    .filter((p) => location.pathname === p || location.pathname.startsWith(`${p}/`));

  const userMenu: MenuProps['items'] = [
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
          {user && (
            <Dropdown menu={{ items: userMenu }} trigger={['click']} placement="bottomRight">
              <button
                type="button"
                aria-label="เมนูผู้ใช้"
                className="ml-auto flex cursor-pointer items-center gap-2.5 rounded-xl border-0 bg-transparent px-2 py-1 text-left font-sans text-ink hover:bg-sky-soft"
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
        </Layout.Header>
        <Layout.Content>
          <div className="box-border flex w-full flex-col gap-6 p-8 max-md:p-4">
            <Outlet />
          </div>
        </Layout.Content>
      </Layout>
    </Layout>
  );
}
