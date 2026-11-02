import { EditOutlined, MoreOutlined, PictureOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Dropdown, Input, type MenuProps, Segmented, Select, Table, type TableProps, Tag } from 'antd';
import { Link, useNavigate, useSearchParams } from 'react-router';
import { tourProgramsApi } from '../../../api/tourPrograms.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import { Permissions } from '../../../constants/permissions.ts';
import { TOUR_PROGRAM_STATE_INFO, type TourProgramState, tourProgramStateInfo } from '../../../constants/tourProgramStates.ts';
import { useCountryOptions } from '../../../features/taxonomy/hooks/useCountryOptions.ts';
import { useWholesaleOptions } from '../../../features/tour-program/hooks/useWholesaleOptions.ts';
import { useListQuery } from '../../../hooks/useListQuery.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { TourProgramListItem } from '../../../types/tourProgram.ts';

type StatusFilter = 'active' | 'inactive' | 'all';
type ParamKey = 'q' | 'country' | 'wholesale' | 'state' | 'status' | 'page' | 'pageSize';
const DEFAULT_PAGE_SIZE = 20;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const statusParam = params.get('status');
  const stateParam = Number(params.get('state'));
  const filters = {
    keyword: params.get('q') ?? '',
    countryId: params.get('country') ?? undefined,
    wholesaleId: params.get('wholesale') ?? undefined,
    state: TOUR_PROGRAM_STATE_INFO.some((s) => s.value === stateParam) ? (stateParam as TourProgramState) : undefined,
    status: (statusParam === 'inactive' || statusParam === 'all' ? statusParam : 'active') as StatusFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<ParamKey, string | undefined>>) => {
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

export default function TourProgramListPage() {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const navigate = useNavigate();
  const { filters, update } = useFilters();
  const { countries } = useCountryOptions();
  const { wholesales } = useWholesaleOptions();
  const { data, loading, error, reload } = useListQuery(tourProgramsApi.list, {
    keyword: filters.keyword || undefined,
    countryId: filters.countryId,
    wholesaleId: filters.wholesaleId,
    state: filters.state,
    status: filters.status === 'all' ? undefined : filters.status === 'active',
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const confirmStatus = (tour: TourProgramListItem) => {
    const remove = tour.status;
    modal.confirm({
      title: remove ? `ลบ ${tour.name}?` : `กู้คืน ${tour.name}?`,
      content: remove
        ? 'ทัวร์นี้จะหายจากหน้าเว็บหลังอัปเดตครั้งถัดไป รอบเดินทาง รูป และเอกสารยังอยู่ครบ กู้คืนได้ภายหลัง'
        : 'ทัวร์นี้จะกลับมาแสดงบนหน้าเว็บตามสถานะโปรแกรม',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await tourProgramsApi.updateStatus({ id: tour.id, status: !remove });
          message.success(remove ? 'ลบแล้ว — หน้าเว็บจะอัปเดตภายในไม่กี่นาที' : 'กู้คืนแล้ว — หน้าเว็บจะอัปเดตภายในไม่กี่นาที');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (tour: TourProgramListItem): MenuProps['items'] =>
    [
      { key: 'edit', icon: <EditOutlined />, label: can(Permissions.TourProgramUpdate) ? 'แก้ไข' : 'ดูรายละเอียด', onClick: () => navigate(`/tour-programs/${tour.id}/edit`) },
      can(Permissions.TourProgramStatus) && {
        key: 'status',
        icon: tour.status ? <StopOutlined /> : <UndoOutlined />,
        label: tour.status ? 'ลบ' : 'กู้คืน',
        danger: tour.status,
        onClick: () => confirmStatus(tour),
      },
    ].filter((item) => item !== false);

  const columns: TableProps<TourProgramListItem>['columns'] = [
    {
      title: 'โปรแกรม',
      key: 'name',
      render: (_, t) => (
        <Link to={`/tour-programs/${t.id}/edit`} className="flex items-center gap-3 text-ink no-underline">
          {(t.cardImage ?? t.bannerImage) ? (
            <img src={(t.cardImage ?? t.bannerImage) as string} alt="" className="h-12 w-16 flex-none rounded-md object-cover" loading="lazy" />
          ) : (
            <span className="flex h-12 w-16 flex-none items-center justify-center rounded-md bg-sky-soft text-ink-muted">
              <PictureOutlined />
            </span>
          )}
          <span className="flex min-w-0 flex-col">
            <span className="font-medium">{t.name}</span>
            <span className="font-mono text-[13px] text-ink-muted">
              {t.code}
              {t.wholesaleTourCode && ` · ${t.wholesaleTourCode}`}
            </span>
          </span>
        </Link>
      ),
    },
    {
      title: 'ประเทศ',
      key: 'countries',
      width: 200,
      render: (_, t) => (
        <div className="flex flex-wrap gap-1">
          {t.countries.map((c) => (
            <Tag key={c.id} className="m-0">
              {c.nameTh}
            </Tag>
          ))}
        </div>
      ),
    },
    { title: 'ระยะเวลา', key: 'duration', width: 120, render: (_, t) => `${t.days} วัน ${t.nights} คืน` },
    {
      title: 'โฮลเซลล์',
      key: 'wholesale',
      width: 180,
      render: (_, t) => (
        <span className="flex flex-col">
          <span>{t.wholesale.name}</span>
          <span className="font-mono text-[13px] text-ink-muted">{t.wholesale.code}</span>
        </span>
      ),
    },
    {
      title: 'สถานะโปรแกรม',
      dataIndex: 'state',
      width: 130,
      render: (state: number) => {
        const info = tourProgramStateInfo(state);
        return (
          <Tag color={info.color} className="m-0" title={info.hint}>
            {info.label}
          </Tag>
        );
      },
    },
    { title: 'แก้ไขล่าสุด', dataIndex: 'updateDate', width: 150, render: (v: string) => <span className="text-[13px]">{formatDateTime(v)}</span> },
    {
      title: 'สถานะ',
      dataIndex: 'status',
      width: 90,
      render: (status: boolean) => (status ? <Tag color="green" className="m-0">ใช้งาน</Tag> : <Tag className="m-0">ลบแล้ว</Tag>),
    },
    {
      key: 'actions',
      width: 56,
      align: 'right',
      render: (_, t) => (
        <Dropdown menu={{ items: rowActions(t) }} trigger={['click']} placement="bottomRight">
          <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${t.name}`} />
        </Dropdown>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title="โปรแกรมทัวร์"
        subtitle="โปรแกรมทัวร์ รอบเดินทาง ราคา และเอกสาร"
        actions={
          can(Permissions.TourProgramCreate) && (
            <Link to="/tour-programs/create">
              <Button type="primary" icon={<PlusOutlined />}>
                เพิ่มโปรแกรมทัวร์
              </Button>
            </Link>
          )
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาชื่อ รหัสทัวร์ หรือรหัสโฮลเซลล์"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[320px] flex-[1_1_240px]"
          />
          <Select
            value={filters.countryId}
            placeholder="ทุกประเทศ"
            allowClear
            showSearch
            optionFilterProp="label"
            options={countries.filter((c) => c.status).map((c) => ({ value: c.id, label: c.nameTh }))}
            onChange={(v?: string) => update({ country: v })}
            className="w-44"
          />
          <Select
            value={filters.wholesaleId}
            placeholder="ทุกโฮลเซลล์"
            allowClear
            showSearch
            optionFilterProp="label"
            options={wholesales.map((w) => ({ value: w.id, label: `${w.name} (${w.code})` }))}
            onChange={(v?: string) => update({ wholesale: v })}
            className="w-52"
          />
          <Select<TourProgramState>
            value={filters.state}
            placeholder="ทุกสถานะโปรแกรม"
            allowClear
            options={TOUR_PROGRAM_STATE_INFO.map((s) => ({ value: s.value, label: s.label }))}
            onChange={(state) => update({ state: state ? String(state) : undefined })}
            className="w-44"
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

        <Table<TourProgramListItem>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            pageSizeOptions: [10, 20, 50, 100],
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} โปรแกรม`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 1100 }}
        />
      </Card>
    </>
  );
}
