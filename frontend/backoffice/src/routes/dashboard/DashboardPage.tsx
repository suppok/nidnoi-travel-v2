import { ReloadOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Empty, Segmented, Skeleton, Tooltip } from 'antd';
import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router';
import { dashboardApi } from '../../api/dashboard.ts';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { Roles, worksAs } from '../../constants/roles.ts';
import LineBookingSection from '../../features/dashboard/components/LineBookingSection.tsx';
import TaskStrip from '../../features/dashboard/components/TaskStrip.tsx';
import TourSection from '../../features/dashboard/components/TourSection.tsx';
import WebBookingSection from '../../features/dashboard/components/WebBookingSection.tsx';
import { formatShortDate } from '../../features/dashboard/lib/dates.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatLongDate } from '../../lib/format.ts';
import { DASHBOARD_DAYS, type DashboardDays, type DashboardSummary } from '../../types/dashboard.ts';

const DAY_LABELS: Record<DashboardDays, string> = { 7: '7 วัน', 30: '30 วัน', 90: '90 วัน', 365: '12 เดือน' };
const time = new Intl.DateTimeFormat('th-TH', { timeStyle: 'short' });

type Result = { key: string; data: DashboardSummary | null; error: unknown };

function parseDays(value: string | null): DashboardDays {
  const n = Number(value);
  return (DASHBOARD_DAYS as readonly number[]).includes(n) ? (n as DashboardDays) : 30;
}

export default function DashboardPage() {
  const { user } = useAuth();
  const [params, setParams] = useSearchParams();
  const days = parseDays(params.get('days'));
  const [today] = useState(() => new Date());
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result | null>(null);
  const key = `${days}#${version}`;

  useEffect(() => {
    let cancelled = false;
    dashboardApi
      .summary(days)
      .then((data) => !cancelled && setResult({ key, data, error: null }))
      .catch((error: unknown) => !cancelled && setResult((prev) => ({ key, data: prev?.data ?? null, error })));
    return () => {
      cancelled = true;
    };
  }, [days, key]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);
  const loading = result?.key !== key;
  const data = result?.data ?? null;
  const error = result?.key === key ? result.error : null;

  const setDays = (d: DashboardDays) =>
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (d === 30) next.delete('days');
        else next.set('days', String(d));
        return next;
      },
      { replace: true },
    );

  // Customer booking requests are a sales area (menus follow worksAs too).
  const showWebBookings = data?.bookings && worksAs(user?.roles, Roles.Sales);
  const nothing = data && !data.tours && !data.bookings && !data.lineBookings;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={`สวัสดี ${user?.name ?? ''}`}
        subtitle={
          <>
            {formatLongDate(today)}
            {data && (
              <span className="ml-2 text-ink-muted">
                · ข้อมูล {formatShortDate(data.range.from)} – {formatShortDate(data.range.to)}
              </span>
            )}
          </>
        }
        actions={
          <>
            <Segmented<DashboardDays> value={days} onChange={setDays} options={DASHBOARD_DAYS.map((d) => ({ label: DAY_LABELS[d], value: d }))} />
            <Tooltip title={data ? `อัปเดตล่าสุด ${time.format(new Date(data.generatedAt))}` : undefined}>
              <Button icon={<ReloadOutlined spin={loading && !!data} />} onClick={reload} disabled={loading}>
                รีเฟรช
              </Button>
            </Tooltip>
          </>
        }
      />

      {error != null && (
        <Alert
          type="error"
          showIcon
          title="โหลดข้อมูลแดชบอร์ดไม่สำเร็จ"
          description={errorMessage(error)}
          action={
            <Button size="small" onClick={reload}>
              ลองอีกครั้ง
            </Button>
          }
        />
      )}

      {!data && loading && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 8 }, (_, i) => (
            <Card key={i}>
              <Skeleton active paragraph={{ rows: 2 }} />
            </Card>
          ))}
        </div>
      )}

      {nothing && (
        <Card>
          <Empty description="ยังไม่มีข้อมูลที่คุณมีสิทธิ์ดูบนแดชบอร์ด" />
        </Card>
      )}

      {data && (
        <div className={`flex flex-col gap-8 transition-opacity ${loading ? 'opacity-60' : ''}`}>
          <TaskStrip summary={data} />
          {data.lineBookings && <LineBookingSection data={data.lineBookings} range={data.range} />}
          {showWebBookings && data.bookings && <WebBookingSection data={data.bookings} range={data.range} />}
          {data.tours && <TourSection data={data.tours} />}
        </div>
      )}
    </div>
  );
}
