import { ReloadOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Input, Segmented, Switch, Table, type TableProps, Tag } from 'antd';
import { useSearchParams } from 'react-router';
import { lineBookingsApi } from '../../api/lineBookings.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { LINE_BOOKING_STATE_INFO, type LineBookingState, LineBookingStates, lineBookingStateInfo } from '../../constants/lineBookingStates.ts';
import LineBookingDrawer from '../../features/line-booking/components/LineBookingDrawer.tsx';
import { useLineBookingAlerts } from '../../features/line-booking/hooks/useLineBookingAlerts.ts';
import { requiredDocuments, travellerSummary, tripDates } from '../../features/line-booking/lib/lineBookingFormat.ts';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatBaht, formatDateTime } from '../../lib/format.ts';
import type { LineBooking } from '../../types/lineBooking.ts';

const DEFAULT_PAGE_SIZE = 20;
const STATES = new Set<number>(LINE_BOOKING_STATE_INFO.map((s) => s.value));
const PAID_STATES: readonly number[] = [LineBookingStates.DepositPaid, LineBookingStates.PaidInFull];

type StateFilter = LineBookingState | 0;
// Finance page: which bookings by documents. Default "waiting".
type DocsFilter = 'waiting' | 'done' | 'all';
const DOCS_FILTERS: { value: DocsFilter; label: string }[] = [
  { value: 'waiting', label: 'รอออกเอกสาร' },
  { value: 'done', label: 'ส่งเอกสารแล้ว' },
  { value: 'all', label: 'ทั้งหมด' },
];

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
    docs: (['done', 'all'].includes(params.get('docs') ?? '') ? params.get('docs') : 'waiting') as DocsFilter,
  };

  const update = (patch: Partial<Record<'q' | 'state' | 'deleted' | 'page' | 'pageSize' | 'open' | 'docs', string | undefined>>) => {
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

export default function LineBookingListPage({ finance = false }: { finance?: boolean }) {
  const { filters, update } = useFilters();
  const { newCount, waitingDocumentsCount, refresh } = useLineBookingAlerts();
  const states = finance ? LINE_BOOKING_STATE_INFO.filter((s) => PAID_STATES.includes(s.value)) : LINE_BOOKING_STATE_INFO;
  const state = finance && !PAID_STATES.includes(filters.state) ? 0 : filters.state;
  const { data, loading, error, reload } = useListQuery(lineBookingsApi.list, {
    keyword: filters.keyword || undefined,
    state: state || undefined,
    paid: finance || undefined,
    documentsDone: finance && filters.docs !== 'all' ? filters.docs === 'done' : undefined,
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
      title: 'เซลล์',
      key: 'sales',
      width: 140,
      render: (_, b) => (
        <div className="flex flex-col">
          <span className="text-sm">{b.salesUser?.name ?? '-'}</span>
          <span className="text-xs text-ink-muted" title="ผู้ส่งผ่าน LINE">
            ส่งโดย {b.lineDisplayName}
          </span>
        </div>
      ),
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
    ...(finance
      ? [
          {
            title: 'ยืนยันเมื่อ',
            key: 'confirmedAt',
            width: 150,
            render: (_: unknown, b: LineBooking) => (b.confirmedAt ? formatDateTime(b.confirmedAt) : '-'),
          },
          {
            title: 'เอกสาร',
            key: 'documents',
            width: 150,
            render: (_: unknown, b: LineBooking) => {
              const codes = requiredDocuments(b)
                .map((d) => d.code)
                .join(' + ');
              if (b.documentsApprovedAt) {
                return (
                  <Tag color="green" className="m-0">
                    {b.customerSentAt ? 'ส่งลูกค้าแล้ว' : 'เซลล์ยืนยันแล้ว'}
                  </Tag>
                );
              }
              if (b.documentsAt) {
                return (
                  <Tag color="blue" className="m-0">
                    ส่ง {codes} แล้ว · รอเซลล์ตรวจ
                  </Tag>
                );
              }
              return b.fixRequestedAt ? (
                <Tag color="red" className="m-0">
                  ต้องแก้ {codes}
                </Tag>
              ) : (
                <Tag color="orange" className="m-0">
                  รอ {codes}
                </Tag>
              );
            },
          },
        ]
      : []),
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
        title={finance ? 'ใบจองรอออกเอกสาร' : 'ใบจอง LINE'}
        subtitle={
          finance
            ? 'ใบจอง LINE ที่เซลล์ยืนยันการชำระเงินแล้ว — มัดจำ: ออก AI ใบรับเงินมัดจำ + IV ใบแจ้งหนี้, ชำระเต็มจำนวน: ออก RE ใบเสร็จรับเงิน'
            : 'ใบจองที่เจ้าหน้าที่ส่งจาก LINE OA พร้อมสลิปโอนเงิน'
        }
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
          {finance && (
            <Segmented<DocsFilter>
              value={filters.docs}
              options={DOCS_FILTERS.map((d) => ({
                value: d.value,
                label:
                  d.value === 'waiting' && waitingDocumentsCount ? (
                    <span className="inline-flex items-center gap-1.5">
                      {d.label} <Badge count={waitingDocumentsCount} size="small" />
                    </span>
                  ) : (
                    d.label
                  ),
              }))}
              onChange={(docs) => update({ docs: docs === 'waiting' ? undefined : docs })}
            />
          )}
          <Segmented<StateFilter>
            value={state}
            options={[
              { value: 0, label: 'ทั้งหมด' },
              ...states.map((s) => ({
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
