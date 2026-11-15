import { AuditOutlined, CheckCircleOutlined, FileDoneOutlined, FileSearchOutlined, MessageOutlined, WalletOutlined } from '@ant-design/icons';
import { Card } from 'antd';
import type { ReactNode } from 'react';
import { Link } from 'react-router';
import { useAuth } from '../../../auth/useAuth.ts';
import { BookingStates } from '../../../constants/bookingStates.ts';
import { FINANCE_LINE_BOOKINGS_PATH, LineBookingStates } from '../../../constants/lineBookingStates.ts';
import { Roles, worksAs } from '../../../constants/roles.ts';
import type { DashboardSummary } from '../../../types/dashboard.ts';
import { formatCount } from '../chartTheme.ts';

type Task = { key: string; title: string; hint: string; count: number; to: string; icon: ReactNode };

// "งานที่รอดำเนินการ": counts that need someone now (any create date), each linking to its list.
export default function TaskStrip({ summary }: { summary: DashboardSummary }) {
  const { user } = useAuth();
  const sales = worksAs(user?.roles, Roles.Sales);
  const finance = worksAs(user?.roles, Roles.Finance);
  const financeOnly = finance && !sales;
  const lb = summary.lineBookings;
  const tasks: Task[] = [];

  if (summary.bookings && sales) {
    tasks.push({
      key: 'web-new',
      title: 'ใบจองเว็บรอติดต่อ',
      hint: 'ลูกค้าส่งคำขอจองจากหน้าเว็บ',
      count: summary.bookings.newCount,
      to: `/bookings?state=${BookingStates.New}`,
      icon: <MessageOutlined />,
    });
  }
  if (lb && sales) {
    tasks.push({
      key: 'line-confirm',
      title: 'ใบจอง LINE รอยืนยัน',
      hint: 'ตรวจสลิปแล้วกดยืนยันการชำระเงิน',
      count: lb.pipeline.waitingConfirm,
      to: `/line-bookings?state=${LineBookingStates.New}`,
      icon: <CheckCircleOutlined />,
    });
  }
  if (lb && (sales || finance)) {
    tasks.push({
      key: 'line-docs',
      title: financeOnly ? 'ใบจองรอออกเอกสาร' : 'รอฝ่ายการเงินออกเอกสาร',
      hint: 'ยืนยันแล้ว รอออก AI / IV / RE',
      count: lb.pipeline.waitingDocuments,
      to: finance ? FINANCE_LINE_BOOKINGS_PATH : '/line-bookings',
      icon: <FileDoneOutlined />,
    });
  }
  if (lb && sales) {
    tasks.push({
      key: 'line-review',
      title: 'เอกสารรอตรวจและส่งลูกค้า',
      hint: 'ฝ่ายการเงินส่งเอกสารให้เซลล์แล้ว',
      count: lb.pipeline.waitingReview,
      to: '/line-bookings',
      icon: <FileSearchOutlined />,
    });
    tasks.push({
      key: 'line-balance',
      title: 'ยอดคงเหลือรอยืนยัน',
      hint: 'ลูกค้าโอนส่วนที่เหลือแล้ว',
      count: lb.pipeline.balancePending,
      to: '/line-bookings',
      icon: <WalletOutlined />,
    });
  }
  if (tasks.length === 0) return null;

  return (
    <section className="flex flex-col gap-3">
      <h2 className="m-0 flex items-center gap-2 text-[17px] font-semibold text-ink">
        <AuditOutlined className="text-ocean" /> งานที่รอดำเนินการ
      </h2>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {tasks.map((t) => (
          <Link key={t.key} to={t.to} className="block">
            <Card hoverable className="h-full" styles={{ body: { padding: 16 } }}>
              <div className="flex items-start gap-3">
                <span
                  className={`grid h-10 w-10 shrink-0 place-items-center rounded-full text-[17px] ${
                    t.count > 0 ? 'bg-brand-soft text-brand-dark' : 'bg-sky text-ocean'
                  }`}
                >
                  {t.icon}
                </span>
                <div className="min-w-0">
                  <div className="font-display text-[24px] leading-none text-ink">{formatCount(t.count)}</div>
                  <div className="mt-1 text-[14px] font-medium text-ink">{t.title}</div>
                  <div className="text-[12px] text-ink-muted">{t.hint}</div>
                </div>
              </div>
            </Card>
          </Link>
        ))}
      </div>
    </section>
  );
}
