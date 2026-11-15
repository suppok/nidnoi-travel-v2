import { GlobalOutlined } from '@ant-design/icons';
import { Card } from 'antd';
import { BookingStates } from '../../../constants/bookingStates.ts';
import type { DashboardBookings, DashboardRange } from '../../../types/dashboard.ts';
import { formatCount, NEUTRAL, SERIES } from '../chartTheme.ts';
import ColumnChart from './ColumnChart.tsx';
import SectionTitle from './SectionTitle.tsx';
import StateBar from './StateBar.tsx';
import StatTile from './StatTile.tsx';

// Booking requests from the customer site (leads; no seats, no payment).
export default function WebBookingSection({ data, range }: { data: DashboardBookings; range: DashboardRange }) {
  const count = (s: number) => data.byState.find((x) => x.state === s)?.count ?? 0;
  const confirmed = count(BookingStates.Confirmed);

  return (
    <section className="flex flex-col gap-4">
      <SectionTitle icon={<GlobalOutlined />} title="ใบจองจากหน้าเว็บ" subtitle="คำขอจองที่ลูกค้าส่งจากเว็บไซต์ (ยังไม่การันตีที่นั่ง)" />
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-1">
          <StatTile title="คำขอจองในช่วงนี้" value={`${formatCount(data.total)} รายการ`} delta={{ current: data.total, previous: data.previousTotal }} />
          <StatTile
            title="ลูกค้ายืนยันการเดินทาง"
            value={`${formatCount(confirmed)} รายการ`}
            footer={data.total ? `คิดเป็น ${Math.round((confirmed / data.total) * 100)}% ของคำขอในช่วงนี้` : 'ยังไม่มีคำขอในช่วงนี้'}
          />
        </div>
        <Card title="คำขอจองตามวันที่ส่ง" className="xl:col-span-2">
          <ColumnChart
            points={data.buckets.map((b) => ({ date: b.date, value: b.count }))}
            unit={range.unit}
            label="คำขอจอง"
            format={(n) => `${formatCount(n)} รายการ`}
            color={SERIES[0]}
            height={200}
          />
          <div className="mt-4">
            <StateBar
              segments={[
                { key: BookingStates.New, label: 'ใหม่', count: count(BookingStates.New), color: SERIES[1] },
                { key: BookingStates.Contacted, label: 'ติดต่อแล้ว', count: count(BookingStates.Contacted), color: SERIES[0] },
                { key: BookingStates.Confirmed, label: 'ยืนยันแล้ว', count: confirmed, color: SERIES[2] },
                { key: BookingStates.Cancelled, label: 'ยกเลิก', count: count(BookingStates.Cancelled), color: NEUTRAL },
              ]}
            />
          </div>
        </Card>
      </div>
    </section>
  );
}
