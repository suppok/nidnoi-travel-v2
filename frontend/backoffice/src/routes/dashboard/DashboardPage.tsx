import { FileDoneOutlined } from '@ant-design/icons';
import { Button, Card, Empty, Statistic } from 'antd';
import { useState } from 'react';
import { Link } from 'react-router';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { BookingStates } from '../../constants/bookingStates.ts';
import { useBookingAlerts } from '../../features/booking/hooks/useBookingAlerts.ts';
import { formatLongDate } from '../../lib/format.ts';

export default function DashboardPage() {
  const { user } = useAuth();
  const { newCount } = useBookingAlerts();
  const [today] = useState(() => new Date());

  return (
    <>
      <PageHeader title={`สวัสดี ${user?.name ?? ''}`} subtitle={formatLongDate(today)} />
      {newCount !== null && (
        <Card className="max-w-[420px]">
          <div className="flex items-end justify-between gap-4">
            <Statistic title="ใบจองใหม่รอติดต่อ" value={newCount} suffix="รายการ" prefix={<FileDoneOutlined />} />
            <Link to={`/bookings?state=${BookingStates.New}`}>
              <Button type={newCount > 0 ? 'primary' : 'default'}>ดูใบจอง</Button>
            </Link>
          </div>
        </Card>
      )}
      <Card>
        <Empty description="ภาพรวมทัวร์และการจองจะแสดงที่นี่" />
      </Card>
    </>
  );
}
