import { CompassOutlined } from '@ant-design/icons';
import { Card, Table, Tag, Tooltip } from 'antd';
import { Link } from 'react-router';
import { TourProgramStates } from '../../../constants/tourProgramStates.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { DashboardPeriod, DashboardTours } from '../../../types/dashboard.ts';
import { formatCount, SERIES } from '../chartTheme.ts';
import { formatDateRange } from '../lib/dates.ts';
import BarList from './BarList.tsx';
import SectionTitle from './SectionTitle.tsx';
import StatTile from './StatTile.tsx';

// Seats sold across the upcoming departures: a single meter with its numbers spelled out.
function SeatMeter({ seats }: { seats: DashboardTours['seats'] }) {
  const sold = Math.max(0, seats.seatsTotal - seats.seatsAvailable);
  const pct = seats.seatsTotal ? Math.round((sold / seats.seatsTotal) * 100) : 0;
  return (
    <Card className="h-full" styles={{ body: { padding: 20 } }}>
      <div className="text-[14px] text-ink-muted">ที่นั่งขายแล้ว · {seats.days} วันข้างหน้า</div>
      <div className="mt-2 font-display text-[28px] leading-tight text-ink">{pct}%</div>
      <Tooltip title={`ขายแล้ว ${formatCount(sold)} จาก ${formatCount(seats.seatsTotal)} ที่นั่ง`}>
        <div className="mt-3 h-2.5 w-full rounded-full bg-sky-soft" role="meter" aria-valuenow={pct} aria-valuemin={0} aria-valuemax={100} aria-label="ที่นั่งขายแล้ว">
          <div className="h-2.5 rounded-full" style={{ width: `${pct}%`, background: SERIES[0] }} />
        </div>
      </Tooltip>
      <div className="mt-2 flex justify-between text-[12.5px] text-ink-muted">
        <span>ว่าง {formatCount(seats.seatsAvailable)} ที่นั่ง</span>
        <span>ทั้งหมด {formatCount(seats.seatsTotal)}</span>
      </div>
    </Card>
  );
}

export default function TourSection({ data }: { data: DashboardTours }) {
  const count = (s: number) => data.byState.find((x) => x.state === s)?.count ?? 0;
  const { seats } = data;

  const columns = [
    {
      title: 'วันเดินทาง',
      key: 'date',
      width: 150,
      render: (_: unknown, p: DashboardPeriod) => <span className="whitespace-nowrap">{formatDateRange(p.startDate, p.endDate)}</span>,
    },
    {
      title: 'ทัวร์',
      key: 'tour',
      render: (_: unknown, p: DashboardPeriod) => (
        <div className="min-w-0">
          <Link to={`/tour-programs/${p.tourProgramId}/edit#periods`} className="block truncate">
            {p.tourName}
          </Link>
          <div className="text-[12px] text-ink-muted">{p.tourCode}</div>
        </div>
      ),
    },
    { title: 'ราคา', key: 'price', width: 110, align: 'right' as const, render: (_: unknown, p: DashboardPeriod) => formatBaht(p.price) },
    {
      title: 'ที่นั่งว่าง',
      key: 'seats',
      width: 120,
      align: 'right' as const,
      render: (_: unknown, p: DashboardPeriod) => (
        <span className="whitespace-nowrap">
          {p.seatsAvailable <= seats.lowSeatsThreshold ? <Tag color="volcano">เหลือ {p.seatsAvailable}</Tag> : p.seatsAvailable}
          <span className="text-ink-muted"> / {p.seatsTotal}</span>
        </span>
      ),
    },
  ];

  return (
    <section className="flex flex-col gap-4">
      <SectionTitle icon={<CompassOutlined />} title="โปรแกรมทัวร์และที่นั่ง" subtitle="ข้อมูล ณ ตอนนี้ (ไม่ขึ้นกับช่วงเวลาที่เลือก)" />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatTile
          title="ทัวร์ที่เปิดขาย"
          value={`${formatCount(count(TourProgramStates.Selling))} โปรแกรม`}
          footer={`ร่าง ${formatCount(count(TourProgramStates.Draft))} · ปิดรับ/เต็ม ${formatCount(count(TourProgramStates.Closed))}`}
        />
        <StatTile title={`รอบเดินทาง ${seats.days} วันข้างหน้า`} value={`${formatCount(seats.periods)} รอบ`} footer="เฉพาะทัวร์ที่แสดงบนหน้าเว็บ" />
        <StatTile
          title="รอบที่ที่นั่งใกล้เต็ม"
          value={`${formatCount(seats.lowSeats)} รอบ`}
          footer={`เหลือไม่เกิน ${seats.lowSeatsThreshold} ที่ · เต็มแล้ว ${formatCount(seats.soldOut)} รอบ`}
        />
        <SeatMeter seats={seats} />
      </div>
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card title="ทัวร์ที่เปิดขายตามประเทศ">
          <BarList
            rows={data.byCountry.map((c) => ({ key: c.countryId, label: c.nameTh, value: c.count, to: `/tour-programs?country=${c.countryId}` }))}
            format={(n) => `${formatCount(n)} โปรแกรม`}
            empty="ยังไม่มีทัวร์ที่เปิดขาย"
          />
        </Card>
        <Card className="xl:col-span-2" title={`รอบที่ที่นั่งเหลือน้อย · ${seats.days} วันข้างหน้า`} styles={{ body: { padding: 0 } }}>
          <Table<DashboardPeriod>
            rowKey="periodId"
            size="middle"
            dataSource={data.lowSeatPeriods}
            columns={columns}
            pagination={false}
            scroll={{ x: 560 }}
            locale={{ emptyText: 'ไม่มีรอบเดินทางที่ยังมีที่ว่างในช่วงนี้' }}
          />
        </Card>
      </div>
    </section>
  );
}
