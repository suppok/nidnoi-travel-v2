import { PlusOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Input, Segmented, Select } from 'antd';
import { useState } from 'react';
import { useSearchParams } from 'react-router';
import { citiesApi } from '../../../api/taxonomy.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import { Permissions } from '../../../constants/permissions.ts';
import CityTable from '../../../features/taxonomy/components/CityTable.tsx';
import TaxonomyFormModal from '../../../features/taxonomy/components/TaxonomyFormModal.tsx';
import { useCountryOptions } from '../../../features/taxonomy/hooks/useCountryOptions.ts';
import { useListQuery } from '../../../hooks/useListQuery.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { City } from '../../../types/taxonomy.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
const DEFAULT_PAGE_SIZE = 50;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const statusParam = params.get('status');
  const filters = {
    keyword: params.get('q') ?? '',
    countryId: params.get('country') ?? undefined,
    status: (statusParam === 'inactive' || statusParam === 'all' ? statusParam : 'active') as StatusFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<'q' | 'country' | 'status' | 'page' | 'pageSize', string | undefined>>) => {
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

export default function CityListPage() {
  const { can } = useAuth();
  const { filters, update } = useFilters();
  const { countries, loading: countriesLoading } = useCountryOptions();
  const { data, loading, error, reload } = useListQuery(citiesApi.list, {
    keyword: filters.keyword || undefined,
    countryId: filters.countryId,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const [editing, setEditing] = useState<City | null>(null);
  const [formOpen, setFormOpen] = useState(false);

  const openForm = (city: City | null) => {
    setEditing(city);
    setFormOpen(true);
  };

  return (
    <>
      <PageHeader
        title="เมือง"
        subtitle="เมือง จังหวัด และภูมิภาคย่อย (เช่น ฮอกไกโด ภาคเหนือ) แต่ละเมืองอยู่ภายใต้ประเทศเดียว"
        actions={
          can(Permissions.CityCreate) && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm(null)}>
              เพิ่มเมือง
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
          <Select<string>
            value={filters.countryId}
            placeholder="ทุกประเทศ"
            allowClear
            showSearch
            optionFilterProp="label"
            loading={countriesLoading}
            options={countries.map((c) => ({ value: c.id, label: c.status ? c.nameTh : `${c.nameTh} (ลบแล้ว)` }))}
            onChange={(country) => update({ country })}
            className="w-56"
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

        <CityTable
          cities={data?.items ?? []}
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
          onEdit={openForm}
          onChanged={reload}
        />
      </Card>

      <TaxonomyFormModal
        kind="city"
        open={formOpen}
        item={editing}
        countries={countries}
        defaultCountryId={filters.countryId}
        onClose={() => setFormOpen(false)}
        onSaved={() => {
          setFormOpen(false);
          reload();
        }}
      />
    </>
  );
}
