import { EditOutlined, MoreOutlined, PhoneOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Dropdown, Input, type MenuProps, Segmented, Select, Table, type TableProps, Tag } from 'antd';
import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { wholesalesApi } from '../../api/wholesales.ts';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { Permissions } from '../../constants/permissions.ts';
import { WHOLESALE_STATE_INFO, type WholesaleState, wholesaleStateInfo } from '../../constants/wholesaleStates.ts';
import WholesaleFormModal from '../../features/wholesale/components/WholesaleFormModal.tsx';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import type { Wholesale, WholesaleContact } from '../../types/wholesale.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
const DEFAULT_PAGE_SIZE = 20;
const SHOWN_CONTACTS = 2;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const statusParam = params.get('status');
  const stateParam = Number(params.get('state'));
  const filters = {
    keyword: params.get('q') ?? '',
    state: WHOLESALE_STATE_INFO.some((s) => s.value === stateParam) ? (stateParam as WholesaleState) : undefined,
    status: (statusParam === 'inactive' || statusParam === 'all' ? statusParam : 'active') as StatusFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<'q' | 'state' | 'status' | 'page' | 'pageSize', string | undefined>>) => {
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

function ContactLine({ contact }: { contact: WholesaleContact }) {
  return (
    <span className="flex items-center gap-1.5 text-[13px]">
      <span className="truncate">{contact.name}</span>
      <a href={`tel:${contact.phone.replace(/[^\d+]/g, '')}`} className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap">
        <PhoneOutlined />
        {contact.phone}
      </a>
    </span>
  );
}

export default function WholesaleListPage() {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useListQuery(wholesalesApi.list, {
    keyword: filters.keyword || undefined,
    state: filters.state,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<Wholesale | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const openForm = (wholesale: Wholesale | null) => {
    setEditing(wholesale);
    setFormOpen(true);
  };

  const confirmStatus = (wholesale: Wholesale) => {
    const remove = wholesale.status;
    modal.confirm({
      title: remove ? `ลบ ${wholesale.name}?` : `กู้คืน ${wholesale.name}?`,
      content: remove ? 'โฮลเซลล์นี้จะถูกซ่อนจากรายการและเลือกให้ทัวร์ใหม่ไม่ได้ ข้อมูลผู้ติดต่อยังอยู่ครบ กู้คืนได้ภายหลัง' : 'โฮลเซลล์นี้จะกลับมาใช้งานได้อีกครั้ง',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await wholesalesApi.updateStatus({ id: wholesale.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (wholesale: Wholesale): MenuProps['items'] =>
    [
      can(Permissions.WholesaleUpdate) && { key: 'edit', icon: <EditOutlined />, label: 'แก้ไข', onClick: () => openForm(wholesale) },
      can(Permissions.WholesaleStatus) && {
        key: 'status',
        icon: wholesale.status ? <StopOutlined /> : <UndoOutlined />,
        label: wholesale.status ? 'ลบ' : 'กู้คืน',
        danger: wholesale.status,
        onClick: () => confirmStatus(wholesale),
      },
    ].filter((item) => item !== false);

  const columns: TableProps<Wholesale>['columns'] = [
    {
      title: 'โฮลเซลล์',
      key: 'name',
      render: (_, w) => (
        <div className="flex flex-col">
          <span className="font-medium">{w.name}</span>
          <span className="font-mono text-[13px] text-ink-muted">{w.code}</span>
        </div>
      ),
    },
    {
      title: 'สถานะการเชื่อมต่อ',
      dataIndex: 'state',
      width: 150,
      render: (state: number) => {
        const info = wholesaleStateInfo(state);
        return (
          <Tag color={info.color} className="m-0">
            {info.label}
          </Tag>
        );
      },
    },
    {
      title: 'ผู้ติดต่อ',
      key: 'contacts',
      render: (_, w) =>
        w.contacts.length === 0 ? (
          <span className="text-ink-muted">—</span>
        ) : (
          <div className="flex max-w-[320px] flex-col gap-0.5">
            {w.contacts.slice(0, SHOWN_CONTACTS).map((c) => (
              <ContactLine key={c.id} contact={c} />
            ))}
            {w.contacts.length > SHOWN_CONTACTS && <span className="text-[13px] text-ink-muted">และอีก {w.contacts.length - SHOWN_CONTACTS} คน</span>}
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
      render: (_, w) => {
        const items = rowActions(w);
        return items && items.length > 0 ? (
          <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
            <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${w.name}`} />
          </Dropdown>
        ) : null;
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="โฮลเซลล์"
        subtitle="บริษัททัวร์ที่เรารับทัวร์มาขาย พร้อมรายชื่อผู้ติดต่อ"
        actions={
          can(Permissions.WholesaleCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm(null)}>
              เพิ่มโฮลเซลล์
            </Button>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อ รหัส ผู้ติดต่อ หรือเบอร์โทร"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[320px] flex-[1_1_240px]"
          />
          <Select<WholesaleState>
            value={filters.state}
            placeholder="ทุกสถานะการเชื่อมต่อ"
            allowClear
            options={WHOLESALE_STATE_INFO.map((s) => ({ value: s.value, label: s.label }))}
            onChange={(state) => update({ state: state ? String(state) : undefined })}
            className="w-52"
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

        <Table<Wholesale>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} ราย`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 860 }}
        />
      </Card>

      <WholesaleFormModal
        open={formOpen}
        wholesale={editing}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
    </>
  );
}
