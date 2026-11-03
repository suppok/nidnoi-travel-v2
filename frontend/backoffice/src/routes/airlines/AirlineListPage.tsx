import { EditOutlined, MoreOutlined, PictureOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Dropdown, Input, type MenuProps, Segmented, Table, type TableProps, Tag } from 'antd';
import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { airlinesApi } from '../../api/airlines.ts';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { Permissions } from '../../constants/permissions.ts';
import AirlineFormModal from '../../features/airline/components/AirlineFormModal.tsx';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import type { Airline } from '../../types/airline.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
const DEFAULT_PAGE_SIZE = 20;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const statusParam = params.get('status');
  const filters = {
    keyword: params.get('q') ?? '',
    status: (statusParam === 'inactive' || statusParam === 'all' ? statusParam : 'active') as StatusFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<'q' | 'status' | 'page' | 'pageSize', string | undefined>>) => {
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

function Logo({ airline }: { airline: Airline }) {
  return (
    <div className="flex h-10 w-24 shrink-0 items-center justify-center overflow-hidden rounded-md border border-line bg-white">
      {airline.logo ? (
        <img src={airline.logo} alt={airline.name} loading="lazy" className="max-h-full max-w-full object-contain p-1" />
      ) : (
        <PictureOutlined className="text-ink-muted" aria-label="ไม่มีโลโก้" />
      )}
    </div>
  );
}

export default function AirlineListPage() {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useListQuery(airlinesApi.list, {
    keyword: filters.keyword || undefined,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<Airline | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const openForm = (airline: Airline | null) => {
    setEditing(airline);
    setFormOpen(true);
  };

  const confirmStatus = (airline: Airline) => {
    const remove = airline.status;
    modal.confirm({
      title: remove ? `ลบ ${airline.name}?` : `กู้คืน ${airline.name}?`,
      content: remove ? 'สายการบินนี้จะถูกซ่อนจากรายการ กู้คืนได้ภายหลัง' : 'สายการบินนี้จะกลับมาใช้งานได้อีกครั้ง',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await airlinesApi.updateStatus({ id: airline.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (airline: Airline): MenuProps['items'] =>
    [
      can(Permissions.AirlineUpdate) && { key: 'edit', icon: <EditOutlined />, label: 'แก้ไข', onClick: () => openForm(airline) },
      can(Permissions.AirlineStatus) && {
        key: 'status',
        icon: airline.status ? <StopOutlined /> : <UndoOutlined />,
        label: airline.status ? 'ลบ' : 'กู้คืน',
        danger: airline.status,
        onClick: () => confirmStatus(airline),
      },
    ].filter((item) => item !== false);

  const columns: TableProps<Airline>['columns'] = [
    { title: 'โลโก้', key: 'logo', width: 128, render: (_, a) => <Logo airline={a} /> },
    {
      title: 'สายการบิน',
      key: 'name',
      render: (_, a) => (
        <div className="flex flex-col">
          <span className="font-medium">{a.name}</span>
          <span className="font-mono text-[13px] text-ink-muted">{a.code}</span>
        </div>
      ),
    },
    { title: 'Ref ID', dataIndex: 'refId', width: 100, render: (v: number | null) => v ?? '—' },
    {
      title: 'สถานะ',
      dataIndex: 'status',
      width: 100,
      render: (status: boolean) => (status ? <Tag color="green" className="m-0">ใช้งาน</Tag> : <Tag className="m-0">ลบแล้ว</Tag>),
    },
    {
      key: 'actions',
      width: 56,
      align: 'right',
      render: (_, a) => {
        const items = rowActions(a);
        return items && items.length > 0 ? (
          <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
            <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${a.name}`} />
          </Dropdown>
        ) : null;
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="สายการบิน"
        subtitle="สายการบินที่ใช้ในโปรแกรมทัวร์ พร้อมรหัสและโลโก้"
        actions={
          can(Permissions.AirlineCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm(null)}>
              เพิ่มสายการบิน
            </Button>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อหรือรหัส"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[320px] flex-[1_1_240px]"
          />
          <Segmented<StatusFilter>
            value={filters.status}
            options={[
              { value: 'active', label: 'ใช้งาน' },
              { value: 'inactive', label: 'ลบแล้ว' },
              { value: 'all', label: 'ทั้งหมด' },
            ]}
            onChange={(status) => update({ status: status === 'active' ? undefined : status })}
          />
        </div>

        {error !== null && <Alert type="error" showIcon title={errorMessage(error)} className="mb-4" />}

        <Table<Airline>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} สายการบิน`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 640 }}
        />
      </Card>

      <AirlineFormModal
        open={formOpen}
        airline={editing}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
    </>
  );
}
