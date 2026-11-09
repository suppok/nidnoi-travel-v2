import { ReloadOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Input, Segmented, Switch, Table, type TableProps, Tag } from 'antd';
import { useSearchParams } from 'react-router';
import { bookingsApi } from '../../api/bookings.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { BOOKING_STATE_INFO, type BookingState, BookingStates, bookingStateInfo } from '../../constants/bookingStates.ts';
import BookingDrawer from '../../features/booking/components/BookingDrawer.tsx';
import { useBookingAlerts } from '../../features/booking/hooks/useBookingAlerts.ts';
import { travellerSummary, tripDates } from '../../features/booking/lib/bookingFormat.ts';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatBaht, formatDateTime } from '../../lib/format.ts';
import type { Booking } from '../../types/booking.ts';

const DEFAULT_PAGE_SIZE = 20;
const STATES = new Set<number>(BOOKING_STATE_INFO.map((s) => s.value));

type StateFilter = BookingState | 0; // 0 = all

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

  // Any filter change goes back to page 1 unless the page itself is being set.
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

export default function BookingListPage() {
  const { filters, update } = useFilters();
  const { newCount, refresh } = useBookingAlerts();
  const { data, loading, error, reload } = useListQuery(bookingsApi.list, {
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

  const columns: TableProps<Booking>['columns'] = [
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
          <span className="font-medium">{[b.firstName, b.lastName].filter(Boolean).join(' ')}</span>
          <span className="text-[13px] text-ink-muted">{b.phone}</span>
        </div>
      ),
    },
    {
      title: 'ทัวร์',
      key: 'tour',
      render: (_, b) => (
        <div className="flex max-w-[340px] flex-col">
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
      width: 150,
      render: (_, b) => {
        const s = bookingStateInfo(b.state);
        return (
          <div className="flex flex-col items-start gap-1">
            <Tag color={s.color} className="m-0">
              {s.label}
            </Tag>
            {b.assignee && <span className="text-xs text-ink-muted">{b.assignee.name}</span>}
          </div>
        );
      },
    },
  ];

  return (
    <>
      <PageHeader
        title="ใบจอง"
        subtitle="ใบจองที่ลูกค้าส่งจากหน้าเว็บ ยังไม่ได้กันที่นั่ง เจ้าหน้าที่ต้องติดต่อกลับเพื่อยืนยัน"
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
            placeholder="ค้นหาเลขที่ใบจอง ชื่อ อีเมล เบอร์ หรือรหัสทัวร์"
            allowClear
            onSearch={(q) => update({ q: q.trim() || undefined })}
            className="max-w-[360px] flex-[1_1_260px]"
          />
          <Segmented<StateFilter>
            value={filters.state}
            options={[
              { value: 0, label: 'ทั้งหมด' },
              ...BOOKING_STATE_INFO.map((s) => ({
                value: s.value,
                label:
                  s.value === BookingStates.New && newCount ? (
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
            แสดงใบจองที่ลบแล้ว
          </label>
        </div>

        {error !== null && <Alert type="error" showIcon title={errorMessage(error)} className="mb-4" />}

        <Table<Booking>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          onRow={(b) => ({ onClick: () => update({ open: b.id }), className: 'cursor-pointer' })}
          rowClassName={(b) => (b.state === BookingStates.New ? 'font-medium' : '')}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} ใบจอง`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 980 }}
        />
      </Card>

      <BookingDrawer bookingId={filters.open} onClose={() => update({ open: undefined })} onChanged={reloadAll} />
    </>
  );
}
