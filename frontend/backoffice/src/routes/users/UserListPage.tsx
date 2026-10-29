import { EditOutlined, KeyOutlined, MoreOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Dropdown, Input, type MenuProps, Segmented, Select, Table, type TableProps, Tag } from 'antd';
import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { usersApi } from '../../api/users.ts';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import RoleTags from '../../components/RoleTags.tsx';
import { Permissions } from '../../constants/permissions.ts';
import { canManageUser, type Role, ROLE_INFO } from '../../constants/roles.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';
import ResetPasswordModal from '../../features/user/components/ResetPasswordModal.tsx';
import UserFormModal from '../../features/user/components/UserFormModal.tsx';
import { useUserList } from '../../features/user/hooks/useUserList.ts';
import type { User } from '../../types/user.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
const DEFAULT_PAGE_SIZE = 20;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const statusParam = params.get('status');
  const filters = {
    keyword: params.get('q') ?? '',
    role: (ROLE_INFO.some((r) => r.value === params.get('role')) ? params.get('role') : undefined) as Role | undefined,
    status: (statusParam === 'inactive' || statusParam === 'all' ? statusParam : 'active') as StatusFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<'q' | 'role' | 'status' | 'page' | 'pageSize', string | undefined>>) => {
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (!('page' in patch)) next.delete('page');
        for (const [key, value] of Object.entries(patch)) {
          if (value) next.set(key, value);
          else next.delete(key);
        }
        return next;
      },
      { replace: true },
    );
  };

  return { filters, update };
}

export default function UserListPage() {
  const { message, modal } = App.useApp();
  const { user: me, can } = useAuth();
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useUserList({
    keyword: filters.keyword || undefined,
    role: filters.role,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<User | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [resetting, setResetting] = useState<User | null>(null);

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };
  const openEdit = (user: User) => {
    setEditing(user);
    setFormOpen(true);
  };

  const confirmStatus = (user: User) => {
    const deactivate = user.status;
    modal.confirm({
      title: deactivate ? `ปิดใช้งาน ${user.name}?` : `เปิดใช้งาน ${user.name} อีกครั้ง?`,
      content: deactivate
        ? 'ผู้ใช้จะถูกออกจากระบบทุกเครื่องและเข้าสู่ระบบไม่ได้จนกว่าจะเปิดใช้งานอีกครั้ง ข้อมูลที่เคยบันทึกยังอยู่ครบ'
        : 'ผู้ใช้จะเข้าสู่ระบบได้อีกครั้งด้วยรหัสผ่านเดิม',
      okText: deactivate ? 'ปิดใช้งาน' : 'เปิดใช้งาน',
      okButtonProps: { danger: deactivate },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await usersApi.updateStatus({ id: user.id, status: !deactivate });
          message.success(deactivate ? 'ปิดใช้งานแล้ว' : 'เปิดใช้งานแล้ว');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (user: User): MenuProps['items'] => {
    const isSelf = user.id === me?.id;
    // Backend: SUPER_ADMIN_PROTECTED / SUPER_ADMIN_REQUIRED. Self-edits are limited to name and email.
    const manageable = canManageUser(me?.roles ?? [], user.roles);
    const locked = isSelf || !manageable;
    return [
      can(Permissions.UserUpdate) && {
        key: 'edit',
        icon: <EditOutlined />,
        label: 'แก้ไข',
        disabled: !isSelf && !manageable,
        onClick: () => openEdit(user),
      },
      can(Permissions.UserPasswordReset) &&
        user.status && { key: 'reset', icon: <KeyOutlined />, label: 'รีเซ็ตรหัสผ่าน', disabled: locked, onClick: () => setResetting(user) },
      can(Permissions.UserStatus) && {
        key: 'status',
        icon: user.status ? <StopOutlined /> : <UndoOutlined />,
        label: user.status ? 'ปิดใช้งาน' : 'เปิดใช้งาน',
        danger: user.status,
        disabled: locked,
        onClick: () => confirmStatus(user),
      },
    ].filter((item) => item !== false);
  };

  const columns: TableProps<User>['columns'] = [
    {
      title: 'ผู้ใช้',
      key: 'user',
      render: (_, user) => (
        <div className="flex flex-col">
          <span className="font-medium">
            {user.name}
            {user.id === me?.id && <span className="ml-1.5 text-[13px] text-ink-muted">(คุณ)</span>}
          </span>
          <span className="text-[13px] text-ink-muted">{user.email}</span>
        </div>
      ),
    },
    { title: 'บทบาท', dataIndex: 'roles', render: (roles: Role[]) => <RoleTags roles={roles} /> },
    {
      title: 'สถานะ',
      key: 'status',
      render: (_, user) => (
        <span className="inline-flex flex-wrap gap-1">
          {user.status ? <Tag color="green" className="m-0">ใช้งาน</Tag> : <Tag className="m-0">ปิดใช้งาน</Tag>}
          {user.status && user.mustChangePassword && <Tag color="orange" className="m-0">รอตั้งรหัสผ่าน</Tag>}
        </span>
      ),
    },
    { title: 'เข้าสู่ระบบล่าสุด', dataIndex: 'lastLoginAt', render: (v: string | null) => formatDateTime(v) },
    {
      key: 'actions',
      width: 56,
      align: 'right',
      render: (_, user) => {
        const items = rowActions(user);
        return items && items.length > 0 ? (
          <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
            <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${user.name}`} />
          </Dropdown>
        ) : null;
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="ผู้ใช้และสิทธิ์"
        subtitle="ทุกคนต้องมีอย่างน้อย 1 บทบาท สิทธิ์ของผู้ใช้คือสิทธิ์รวมของทุกบทบาทที่มี"
        actions={
          can(Permissions.UserCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              เพิ่มผู้ใช้
            </Button>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อหรืออีเมล"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[320px] flex-[1_1_240px]"
          />
          <Select<Role>
            value={filters.role}
            placeholder="ทุกบทบาท"
            allowClear
            options={ROLE_INFO.map((r) => ({ value: r.value, label: r.label }))}
            onChange={(role) => update({ role })}
            className="w-44"
          />
          <Segmented<StatusFilter>
            value={filters.status}
            options={[
              { value: 'active', label: 'ใช้งาน' },
              { value: 'inactive', label: 'ปิดใช้งาน' },
              { value: 'all', label: 'ทั้งหมด' },
            ]}
            onChange={(status) => update({ status: status === 'active' ? undefined : status })}
          />
        </div>

        {error !== null && <Alert type="error" showIcon title={errorMessage(error)} className="mb-4" />}

        <Table<User>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} คน`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 720 }}
        />
      </Card>

      <UserFormModal
        open={formOpen}
        user={editing}
        isSelf={editing !== null && editing.id === me?.id}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
      <ResetPasswordModal
        user={resetting}
        onClose={() => setResetting(null)}
        onSaved={() => {
          setResetting(null);
          reload();
        }}
      />
    </>
  );
}
