import { ReloadOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Input, Segmented, Switch, Table, type TableProps, Tag } from 'antd';
import { useSearchParams } from 'react-router';
import { lineBookingsApi } from '../../api/lineBookings.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { LINE_BOOKING_STATE_INFO, type LineBookingState, LineBookingStates, lineBookingStateInfo } from '../../constants/lineBookingStates.ts';
import LineBookingDrawer from '../../features/line-booking/components/LineBookingDrawer.tsx';
import { useLineBookingAlerts } from '../../features/line-booking/hooks/useLineBookingAlerts.ts';
import { travellerSummary, tripDates } from '../../features/line-booking/lib/lineBookingFormat.ts';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatBaht, formatDateTime } from '../../lib/format.ts';
import type { LineBooking } from '../../types/lineBooking.ts';

const DEFAULT_PAGE_SIZE = 20;
const STATES = new Set<number>(LINE_BOOKING_STATE_INFO.map((s) => s.value));

type StateFilter = LineBookingState | 0;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const state = Number(params.get('state'));
  const filters = {
    keyword: params.get('q') ?? '',
    state: (STATES.has(state) ? state : 0) as StateFilter,
    deleted: params.get('deleted') === '1',
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
    open: params.get('open'),
  };

  const update = (patch: Partial<Record<'q' | 'state' | 'deleted' | 'page' | 'pageSize' | 'open', string | undefined>>) => {
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (!('page' in patch) && !('open' in patch)) next.delete('page');
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

export default function LineBookingListPage() {
  const { filters, update } = useFilters();
  const { newCount, refresh } = useLineBookingAlerts();
  const { data, loading, error, reload } = useListQuery(lineBookingsApi.list, {
    keyword: filters.keyword || undefined,
    state: filters.state || undefined,
    status: !filters.deleted,
    page: filters.page,
    pageSize: filters.pageSize,
  });

  const reloadAll = () => {
    reload();
    refresh();
  };

  const columns: TableProps<LineBooking>['columns'] = [
    {
      title: 'ใบจอง',
      key: 'no',
      width: 170,
      render: (_, b) => (
        <div className="flex flex-col">
          <span className="font-mono font-medium">{b.bookingNo}</span>
          <span className="text-xs text-ink-muted">{formatDateTime(b.createDate)}</span>
        </div>
      ),
    },
    {
      title: 'ลูกค้า',
      key: 'customer',
      render: (_, b) => (
        <div className="flex flex-col">
          <span className="font-medium">{b.customerName}</span>
          <span className="text-[13px] text-ink-muted">{b.customerPhone}</span>
        </div>
      ),
    },
    {
      title: 'LINE',
      key: 'line',
      width: 140,
      render: (_, b) => <span className="text-sm text-ink-muted">{b.lineDisplayName}</span>,
    },
    {
      title: 'ทัวร์',
      key: 'tour',
      render: (_, b) => (
        <div className="flex max-w-[300px] flex-col">
          <span className="truncate" title={b.tourName}>
            <span className="font-mono">{b.tourCode}</span> {b.tourName}
          </span>
          <span className="text-[13px] text-ink-muted">{tripDates(b)}</span>
        </div>
      ),
    },
    { title: 'ผู้เดินทาง', key: 'pax', width: 130, render: (_, b) => travellerSummary(b) },
    { title: 'ยอดรวม', dataIndex: 'totalPrice', width: 120, align: 'right', render: (v: number) => formatBaht(v) },
    {
      title: 'สถานะ',
      key: 'state',
      width: 120,
      render: (_, b) => {
        const s = lineBookingStateInfo(b.state);
        return (
          <Tag color={s.color} className="m-0">
            {s.label}
          </Tag>
        );
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="ใบจอง LINE"
        subtitle="ใบจองที่เจ้าหน้าที่ส่งจาก LINE OA พร้อมสลิปโอนเงิน"
        actions={
          <Button icon={<ReloadOutlined />} onClick={reloadAll}>
            รีเฟรช
          </Button>
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input.Search
            key={filters.keyword}
            defaultValue={filters.keyword}
            placeholder="ค้นหาเลขที่ใบจอง ชื่อ เบอร์ รหัสทัวร์ หรือ LINE ชื่อ"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[360px] flex-[1_1_260px]"
          />
          <Segmented<StateFilter>
            value={filters.state}
            options={[
              { value: 0, label: 'ทั้งหมด' },
              ...LINE_BOOKING_STATE_INFO.map((s) => ({
                value: s.value,
                label:
                  s.value === LineBookingStates.New && newCount ? (
                    <span className="inline-flex items-center gap-1.5">
                      {s.label} <Badge count={newCount} size="small" />
                    </span>
                  ) : (
                    s.label
                  ),
              })),
            ]}
            onChange={(state) => update({ state: state ? String(state) : undefined })}
          />
          <label className="ml-auto flex items-center gap-2 text-sm text-ink-muted">
            <Switch size="small" checked={filters.deleted} onChange={(on) => update({ deleted: on ? '1' : undefined })} />
            แสดงที่ลบแล้ว
          </label>
        </div>

        {error !== null && <Alert type="error" showIcon title={errorMessage(error)} className="mb-4" />}

        <Table<LineBooking>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          onRow={(b) => ({ onClick: () => update({ open: b.id }), className: 'cursor-pointer' })}
          rowClassName={(b) => (b.state === LineBookingStates.New ? 'font-medium' : '')}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} ใบจอง`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 1080 }}
        />
      </Card>

      <LineBookingDrawer bookingId={filters.open} onClose={() => update({ open: undefined })} onChanged={reloadAll} />
    </>
  );
}
