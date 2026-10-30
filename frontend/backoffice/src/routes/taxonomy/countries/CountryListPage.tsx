import { EditOutlined, MoreOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Dropdown, Input, type MenuProps, Segmented, Table, type TableProps, Tag, Typography } from 'antd';
import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router';
import { countriesApi } from '../../../api/taxonomy.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import { Permissions } from '../../../constants/permissions.ts';
import TaxonomyFormModal from '../../../features/taxonomy/components/TaxonomyFormModal.tsx';
import { useConfirmCountryStatus } from '../../../features/taxonomy/hooks/useConfirmCountryStatus.ts';
import { useListQuery } from '../../../hooks/useListQuery.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { Country } from '../../../types/taxonomy.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
const DEFAULT_PAGE_SIZE = 50;

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

export default function CountryListPage() {
  const { can } = useAuth();
  const navigate = useNavigate();
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useListQuery(countriesApi.list, {
    keyword: filters.keyword || undefined,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<Country | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const openForm = (country: Country | null) => {
    setEditing(country);
    setFormOpen(true);
  };

  const confirmStatus = useConfirmCountryStatus(reload);

  const rowActions = (country: Country): MenuProps['items'] =>
    [
      can(Permissions.CountryUpdate) && { key: 'edit', icon: <EditOutlined />, label: 'แก้ไข', onClick: () => openForm(country) },
      can(Permissions.CountryStatus) && {
        key: 'status',
        icon: country.status ? <StopOutlined /> : <UndoOutlined />,
        label: country.status ? 'ลบ' : 'กู้คืน',
        danger: country.status,
        onClick: () => confirmStatus(country),
      },
    ].filter((item) => item !== false);

  const columns: TableProps<Country>['columns'] = [
    {
      title: 'ประเทศ',
      key: 'name',
      render: (_, country) => (
        <div className="flex flex-col">
          <Link
            to={`/taxonomy/countries/${country.id}`}
            onClick={(e) => e.stopPropagation()} // the row handler would navigate a second time
            className="font-medium text-ink no-underline hover:text-brand"
          >
            {country.nameTh}
          </Link>
          <span className="text-[13px] text-ink-muted">{country.nameEn}</span>
        </div>
      ),
    },
    { title: 'Slug', dataIndex: 'slug', render: (slug: string) => <Typography.Text code>{slug}</Typography.Text> },
    { title: 'Ref ID', dataIndex: 'refId', width: 100, render: (v: number | null) => v ?? '—' },
    { title: 'ลำดับ', dataIndex: 'sortOrder', width: 80 },
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
      render: (_, country) => {
        const items = rowActions(country);
        // stopPropagation: menu clicks bubble (through the portal) to the row's click handler.
        return items && items.length > 0 ? (
          <span onClick={(e) => e.stopPropagation()}>
            <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
              <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${country.nameTh}`} />
            </Dropdown>
          </span>
        ) : null;
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="ประเทศ"
        subtitle="ใช้จัดหมวดทัวร์และสร้างหน้า landing ของแต่ละประเทศ (รวมทวีปและภูมิภาค เช่น ยุโรป สแกนดิเนเวีย)"
        actions={
          can(Permissions.CountryCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm(null)}>
              เพิ่มประเทศ
            </Button>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อไทย อังกฤษ หรือ slug"
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

        <Table<Country>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            pageSizeOptions: [20, 50, 100],
            showTotal: (total) => `ทั้งหมด ${total} รายการ`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          onRow={(country) => ({ onClick: () => navigate(`/taxonomy/countries/${country.id}`), className: 'cursor-pointer' })}
          scroll={{ x: 720 }}
        />
      </Card>

      <TaxonomyFormModal
        kind="country"
        open={formOpen}
        item={editing}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
    </>
  );
}
