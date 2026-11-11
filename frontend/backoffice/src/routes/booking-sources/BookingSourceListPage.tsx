import { EditOutlined, MoreOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Dropdown, Input, type MenuProps, Segmented, Table, type TableProps, Tag } from 'antd';
import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { bookingSourcesApi } from '../../api/bookingSources.ts';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { Permissions } from '../../constants/permissions.ts';
import BookingSourceFormModal from '../../features/booking-source/components/BookingSourceFormModal.tsx';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import type { BookingSource } from '../../types/bookingSource.ts';

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

export default function BookingSourceListPage() {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useListQuery(bookingSourcesApi.list, {
    keyword: filters.keyword || undefined,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<BookingSource | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const openForm = (source: BookingSource | null) => {
    setEditing(source);
    setFormOpen(true);
  };

  const confirmStatus = (source: BookingSource) => {
    const remove = source.status;
    modal.confirm({
      title: remove ? `ลบ ${source.name}?` : `กู้คืน ${source.name}?`,
      content: remove ? 'ช่องทางนี้จะไม่แสดงในเครื่องมือฝ่ายขาย กู้คืนได้ภายหลัง' : 'ช่องทางนี้จะกลับมาใช้งานได้อีกครั้ง',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await bookingSourcesApi.updateStatus({ id: source.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (source: BookingSource): MenuProps['items'] =>
    [
      can(Permissions.BookingSourceUpdate) && { key: 'edit', icon: <EditOutlined />, label: 'แก้ไข', onClick: () => openForm(source) },
      can(Permissions.BookingSourceStatus) && {
        key: 'status',
        icon: source.status ? <StopOutlined /> : <UndoOutlined />,
        label: source.status ? 'ลบ' : 'กู้คืน',
        danger: source.status,
        onClick: () => confirmStatus(source),
      },
    ].filter((item) => item !== false);

  const columns: TableProps<BookingSource>['columns'] = [
    { title: 'ชื่อช่องทาง', dataIndex: 'name', render: (v: string) => <span className="font-medium">{v}</span> },
    { title: 'ชื่อย่อ', dataIndex: 'shortName', render: (v: string) => v || '—' },
    { title: 'ลำดับ', dataIndex: 'sortOrder', width: 100 },
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
      render: (_, s) => {
        const items = rowActions(s);
        return items && items.length > 0 ? (
          <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
            <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${s.name}`} />
          </Dropdown>
        ) : null;
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="ช่องทางการจอง"
        subtitle="ช่องทางที่ลูกค้าติดต่อจองเข้ามา ใช้ในเครื่องมือสรุปยอดโอน"
        actions={
          can(Permissions.BookingSourceCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm(null)}>
              เพิ่มช่องทาง
            </Button>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อ"
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

        <Table<BookingSource>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} ช่องทาง`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 560 }}
        />
      </Card>

      <BookingSourceFormModal
        open={formOpen}
        source={editing}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
    </>
  );
}
