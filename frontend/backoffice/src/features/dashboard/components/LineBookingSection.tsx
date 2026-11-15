import { DollarOutlined, FileTextOutlined, LineChartOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons';
import { Card, Segmented, Table, Tag } from 'antd';
import { useState } from 'react';
import { Link } from 'react-router';
import { useAuth } from '../../../auth/useAuth.ts';
import { FINANCE_LINE_BOOKINGS_PATH, LineBookingStates, lineBookingStateInfo } from '../../../constants/lineBookingStates.ts';
import { Roles, worksAs } from '../../../constants/roles.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { DashboardDeparture, DashboardLineBookings, DashboardLineGroup, DashboardRange } from '../../../types/dashboard.ts';
import { formatCount, NEUTRAL, SERIES } from '../chartTheme.ts';
import { formatDateRange } from '../lib/dates.ts';
import BarList, { type BarListRow } from './BarList.tsx';
import ColumnChart from './ColumnChart.tsx';
import SectionTitle from './SectionTitle.tsx';
import StateBar from './StateBar.tsx';
import StatTile from './StatTile.tsx';

type Measure = 'revenue' | 'count';

const UNIT_TITLE = { day: 'แนวโน้มรายวัน', week: 'แนวโน้มรายสัปดาห์', month: 'แนวโน้มรายเดือน' } as const;

function groupRows(groups: DashboardLineGroup[], none: string, to?: (g: DashboardLineGroup) => string | undefined): BarListRow[] {
  return groups.map((g) => ({
    key: g.key || '__none',
    label: g.label || <span className="text-ink-muted">{none}</span>,
    value: g.revenue,
    detail: `${formatCount(g.count)} ใบจอง · ชำระแล้ว ${formatCount(g.paidCount)} · ${formatCount(g.travellers)} คน`,
    to: to?.(g),
  }));
}

export default function LineBookingSection({ data, range }: { data: DashboardLineBookings; range: DashboardRange }) {
  const { user } = useAuth();
  const [measure, setMeasure] = useState<Measure>('revenue');
  const isFinance = worksAs(user?.roles, Roles.Finance);
  const financeOnly = isFinance && !worksAs(user?.roles, Roles.Sales);
  const listPath = financeOnly ? FINANCE_LINE_BOOKINGS_PATH : '/line-bookings';
  const { totals, previous } = data;

  const points = data.buckets.map((b) => ({ date: b.date, value: measure === 'revenue' ? b.revenue : b.count }));
  const stateCount = (s: number) => data.byState.find((x) => x.state === s)?.count ?? 0;
  const departureColumns = [
    {
      title: 'วันเดินทาง',
      key: 'date',
      width: 150,
      render: (_: unknown, d: DashboardDeparture) => <span className="whitespace-nowrap">{formatDateRange(d.startDate, d.endDate)}</span>,
    },
    {
      title: 'ทัวร์',
      key: 'tour',
      render: (_: unknown, d: DashboardDeparture) => (
        <div className="min-w-0">
          <div className="truncate">{d.tourName}</div>
          <div className="text-[12px] text-ink-muted">{d.tourCode}</div>
        </div>
      ),
    },
    {
      title: 'ลูกค้า',
      key: 'customer',
      render: (_: unknown, d: DashboardDeparture) => (
        <div>
          <Link to={`${listPath}?open=${d.id}`}>{d.customerName}</Link>
          <div className="text-[12px] text-ink-muted">
            {d.bookingNo}
            {data.scopeAll && d.salesUser ? ` · ${d.salesUser.name}` : ''}
          </div>
        </div>
      ),
    },
    { title: 'คน', dataIndex: 'travellers', key: 'travellers', width: 60, align: 'right' as const },
    {
      title: 'สถานะ',
      key: 'state',
      width: 130,
      render: (_: unknown, d: DashboardDeparture) => {
        const info = lineBookingStateInfo(d.state);
        return <Tag color={info.color}>{info.label}</Tag>;
      },
    },
  ];

  return (
    <section className="flex flex-col gap-4">
      <SectionTitle
        icon={<LineChartOutlined />}
        title="ใบจอง LINE"
        subtitle={data.scopeAll ? 'ทุกใบจองในระบบ' : financeOnly ? 'ใบจองที่ส่งมาให้คุณ' : isFinance ? 'ใบจองที่คุณเป็นเซลล์เจ้าของ และที่ส่งมาให้คุณ' : 'เฉพาะใบจองที่คุณเป็นเซลล์เจ้าของ'}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatTile title="ยอดขาย (ชำระแล้ว)" icon={<DollarOutlined />} value={formatBaht(totals.revenue)} delta={{ current: totals.revenue, previous: previous.revenue }} />
        <StatTile
          title="ใบจองที่ชำระแล้ว"
          icon={<FileTextOutlined />}
          value={
            <>
              {formatCount(totals.paidCount)} <span className="text-[15px] text-ink-muted">/ {formatCount(totals.count)} ใบ</span>
            </>
          }
          delta={{ current: totals.paidCount, previous: previous.paidCount }}
        />
        <StatTile title="ผู้เดินทาง" icon={<TeamOutlined />} value={`${formatCount(totals.travellers)} คน`} delta={{ current: totals.travellers, previous: previous.travellers }} />
        <StatTile title="ยอดเฉลี่ยต่อใบจอง" icon={<UserOutlined />} value={formatBaht(totals.averageValue)} delta={{ current: totals.averageValue, previous: previous.averageValue }} />
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card
          className="xl:col-span-2"
          title={UNIT_TITLE[range.unit]}
          extra={
            <Segmented<Measure>
              size="small"
              value={measure}
              onChange={setMeasure}
              options={[
                { label: 'ยอดขาย', value: 'revenue' },
                { label: 'จำนวนใบจอง', value: 'count' },
              ]}
            />
          }
        >
          <ColumnChart
            points={points}
            unit={range.unit}
            label={measure === 'revenue' ? 'ยอดขาย' : 'ใบจอง'}
            format={measure === 'revenue' ? formatBaht : (n) => `${formatCount(n)} ใบ`}
          />
          <p className="m-0 mt-2 text-[12px] text-ink-muted">
            {measure === 'revenue' ? 'ยอดรวมของใบจองที่ชำระมัดจำหรือชำระครบแล้ว นับตามวันที่บันทึกใบจอง' : 'ทุกใบจองที่บันทึก (ไม่รวมที่ลบแล้ว)'}
          </p>
        </Card>
        <Card title="สถานะใบจอง">
          <StateBar
            unit="ใบ"
            segments={[
              { key: LineBookingStates.New, label: 'ใหม่ รอยืนยัน', count: stateCount(LineBookingStates.New), color: SERIES[1] },
              { key: LineBookingStates.DepositPaid, label: 'ชำระมัดจำแล้ว', count: stateCount(LineBookingStates.DepositPaid), color: SERIES[0] },
              { key: LineBookingStates.PaidInFull, label: 'ชำระครบแล้ว', count: stateCount(LineBookingStates.PaidInFull), color: SERIES[2] },
              { key: LineBookingStates.Cancelled, label: 'ยกเลิก', count: stateCount(LineBookingStates.Cancelled), color: NEUTRAL },
            ]}
          />
          {totals.count > 0 && (
            <p className="m-0 mt-4 text-[13px] text-ink-muted">
              อัตราการชำระ <span className="font-semibold text-ink">{Math.round((totals.paidCount / totals.count) * 100)}%</span> · ยกเลิก{' '}
              <span className="font-semibold text-ink">{Math.round((totals.cancelled / totals.count) * 100)}%</span>
            </p>
          )}
        </Card>
      </div>

      <div className={`grid grid-cols-1 gap-4 ${data.scopeAll ? 'xl:grid-cols-3' : 'lg:grid-cols-2'}`}>
        <Card title="ช่องทางการจอง" extra={<span className="text-[12px] text-ink-muted">เรียงตามยอดขาย</span>}>
          <BarList rows={groupRows(data.bySource, 'ไม่ระบุช่องทาง')} format={formatBaht} />
        </Card>
        <Card title="ทัวร์ขายดี" extra={<span className="text-[12px] text-ink-muted">เรียงตามยอดขาย</span>}>
          <BarList rows={groupRows(data.byTour, 'ไม่ระบุ', (g) => (g.key && !financeOnly ? `/line-bookings?q=${encodeURIComponent(g.key)}` : undefined))} format={formatBaht} />
        </Card>
        {data.scopeAll && (
          <Card title="ยอดขายตามเซลล์" extra={<span className="text-[12px] text-ink-muted">เรียงตามยอดขาย</span>}>
            <BarList rows={groupRows(data.bySales, 'ไม่มีเซลล์เจ้าของ')} format={formatBaht} />
          </Card>
        )}
      </div>

      <Card title={`ลูกค้าที่จะเดินทางใน ${data.departuresDays} วันข้างหน้า`} styles={{ body: { padding: 0 } }}>
        <Table<DashboardDeparture>
          rowKey="id"
          size="middle"
          dataSource={data.departures}
          columns={departureColumns}
          pagination={false}
          scroll={{ x: 640 }}
          locale={{ emptyText: 'ไม่มีลูกค้าเดินทางในช่วงนี้' }}
        />
      </Card>
    </section>
  );
}
