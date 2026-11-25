import { CheckCircleFilled, EditOutlined, WalletOutlined } from '@ant-design/icons';
import { App, Button, Card, Result, Skeleton, Tag, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { bookingSourcesApi } from '../../../api/bookingSources.ts';
import { usersApi } from '../../../api/users.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import logoUrl from '../../../assets/logo.png';
import { env } from '../../../config/env.ts';
import BalancePanel from '../../../features/line-booking-entry/components/BalancePanel.tsx';
import EntryForm, { type EntryLookups, type SavedResult } from '../../../features/line-booking-entry/components/EntryForm.tsx';
import ReopenedSummary from '../../../features/line-booking-entry/components/ReopenedSummary.tsx';
import ReopenPanel from '../../../features/line-booking-entry/components/ReopenPanel.tsx';
import SummaryPasteBox from '../../../features/line-booking-entry/components/SummaryPasteBox.tsx';
import WholesaleEmailNotice from '../../../features/line-booking-entry/components/WholesaleEmailNotice.tsx';
import { canRecordBalance, dealFromBooking, dealTotals } from '../../../features/line-booking-entry/lib/deal.ts';
import type { ParsedSummary } from '../../../features/line-booking-entry/lib/parseSummary.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';

type Step =
  | { kind: 'start' }
  | { kind: 'create'; summary: ParsedSummary | null; seq: number }
  | { kind: 'choose'; booking: LineBooking }
  | { kind: 'edit'; booking: LineBooking }
  | { kind: 'balance'; booking: LineBooking }
  | { kind: 'done'; result: SavedResult };

type PartialLookups = { [K in keyof EntryLookups]: EntryLookups[K] | null };

// Sales owner options and booking sources, loaded once for the whole page; null until both arrive.
function useLookups(): EntryLookups | null {
  const { message } = App.useApp();
  const [lookups, setLookups] = useState<PartialLookups>({ sales: null, sources: null });
  useEffect(() => {
    let cancelled = false;
    usersApi
      .sales()
      .then((sales) => !cancelled && setLookups((l) => ({ ...l, sales })))
      .catch((err: unknown) => {
        if (cancelled) return;
        message.error(errorMessage(err));
        setLookups((l) => ({ ...l, sales: { items: [], defaultId: '' } }));
      });
    bookingSourcesApi
      .list({ status: true, page: 1, pageSize: 500 })
      .then((res) => !cancelled && setLookups((l) => ({ ...l, sources: res.items })))
      .catch(() => !cancelled && setLookups((l) => ({ ...l, sources: [] })));
    return () => {
      cancelled = true;
    };
  }, [message]);
  const { sales, sources } = lookups;
  return sales && sources ? { sales, sources } : null;
}

// "บันทึกใบจอง LINE": sales enter a booking for a customer they serve in LINE OA chat (with its
// payment slips), or reopen a paid one to edit it or record the balance. Opens in its own tab from
// the sidebar, without the sidebar, and works on a phone.
export default function LineBookingEntryPage() {
  const { user } = useAuth();
  const lookups = useLookups();
  const [step, setStep] = useState<Step>({ kind: 'start' });

  useEffect(() => {
    document.title = 'บันทึกใบจอง LINE · nidnoi';
  }, []);

  // seq remounts the form, so a new booking always starts clean.
  const startNew = (summary: ParsedSummary | null) => setStep({ kind: 'create', summary, seq: Date.now() });
  const open = (booking: LineBooking) => setStep(canRecordBalance(booking) ? { kind: 'choose', booking } : { kind: 'edit', booking });
  const reset = () => setStep({ kind: 'start' });
  const saved = (result: SavedResult) => {
    setStep({ kind: 'done', result });
    window.scrollTo({ top: 0 });
  };

  const title =
    step.kind === 'edit'
      ? `แก้ไขใบจอง ${step.booking.bookingNo}`
      : step.kind === 'balance'
        ? `ชำระส่วนที่เหลือ ${step.booking.bookingNo}`
        : 'บันทึกใบจอง LINE';

  return (
    <div className="min-h-full bg-page">
      <header className="sticky top-0 z-10 border-b border-line bg-white/95 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-[1180px] items-center gap-3 px-4">
          <Link to="/line-bookings" className="flex shrink-0" aria-label="ไปหน้าใบจอง LINE">
            <img src={logoUrl} alt="nidnoi travel" className="h-9 w-9 rounded-lg" />
          </Link>
          <div className="min-w-0 flex-1 leading-tight">
            <h1 className="m-0 truncate font-display text-base font-normal">{title}</h1>
            <p className="m-0 truncate text-xs text-ink-muted">โดย {user?.name}</p>
          </div>
          {env.APP_ENV !== 'prd' && <Tag color="orange">{env.APP_ENV.toUpperCase()}</Tag>}
        </div>
      </header>

      <main className="mx-auto max-w-[1180px] px-4 py-5">
        {step.kind === 'start' && (
          <div className="grid items-start gap-5 lg:grid-cols-2">
            <Card title="ใบจองใหม่ — วางข้อความสรุปยอดโอน" size="small">
              <Typography.Paragraph type="secondary" className="text-[13px]">
                คัดลอกข้อความจากเครื่องมือ "สรุปยอดโอน" แล้ววางที่นี่ ระบบจะกรอกฟอร์มให้อัตโนมัติ
              </Typography.Paragraph>
              <SummaryPasteBox
                autoFocus
                onApply={(s) => startNew(s)}
                extra={<Button onClick={() => startNew(null)}>ข้าม กรอกเอง</Button>}
              />
            </Card>
            <Card title="แก้ไขใบจองเดิม / ลูกค้าชำระส่วนที่เหลือ" size="small">
              <Typography.Paragraph type="secondary" className="text-[13px]">
                เปิดใบจองที่ยืนยันการชำระเงินแล้ว (มัดจำหรือชำระเต็มจำนวน)
              </Typography.Paragraph>
              <ReopenPanel onOpen={open} />
            </Card>
          </div>
        )}

        {(step.kind === 'create' || step.kind === 'edit') && !lookups && (
          <Card>
            <Skeleton active paragraph={{ rows: 10 }} />
          </Card>
        )}

        {step.kind === 'create' && lookups && (
          <EntryForm key={step.seq} booking={null} initialSummary={step.summary} lookups={lookups} onSaved={saved} onCancel={reset} />
        )}

        {step.kind === 'choose' && <PurposeChoice booking={step.booking} onCancel={reset} onChoose={(kind) => setStep({ kind, booking: step.booking })} />}

        {step.kind === 'edit' && lookups && (
          <EntryForm
            key={step.booking.id}
            booking={step.booking}
            initialSummary={null}
            lookups={lookups}
            onSaved={saved}
            onCancel={reset}
            onChangePurpose={canRecordBalance(step.booking) ? () => setStep({ kind: 'choose', booking: step.booking }) : undefined}
          />
        )}

        {step.kind === 'balance' && (
          <BalancePanel booking={step.booking} onSaved={saved} onCancel={reset} onChangePurpose={() => setStep({ kind: 'choose', booking: step.booking })} />
        )}

        {step.kind === 'done' && (
          <Card className="mx-auto max-w-[560px]">
            <Result
              icon={<CheckCircleFilled className="text-green-600" />}
              title="บันทึกสำเร็จ"
              subTitle={step.result.message ?? 'ข้อมูลการจองถูกบันทึกเข้าระบบแล้ว รอเซลล์เจ้าของการจองยืนยันการชำระเงิน'}
              extra={
                <Button type="primary" size="large" onClick={reset}>
                  บันทึกรายการใหม่
                </Button>
              }
            >
              <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
                <dt className="text-ink-muted">เลขที่จอง</dt>
                <dd className="m-0 text-right">
                  <Typography.Text strong copyable className="font-mono">
                    {step.result.bookingNo}
                  </Typography.Text>
                </dd>
                <dt className="text-ink-muted">ยอดรวม</dt>
                <dd className="m-0 text-right font-semibold">{formatBaht(step.result.totalPrice)}</dd>
              </dl>
              {step.result.wholesaleEmail && <WholesaleEmailNotice email={step.result.wholesaleEmail} />}
            </Result>
          </Card>
        )}
      </main>
    </div>
  );
}

// A reopened booking that can take a balance: record it, or edit the booking.
function PurposeChoice({ booking, onCancel, onChoose }: { booking: LineBooking; onCancel: () => void; onChoose: (kind: 'balance' | 'edit') => void }) {
  const remaining = dealTotals(dealFromBooking(booking)).remaining;
  return (
    <div className="mx-auto flex w-full max-w-[720px] flex-col gap-4">
      <ReopenedSummary booking={booking} onCancel={onCancel} />
      <Typography.Text strong>ต้องการทำอะไรกับใบจองนี้?</Typography.Text>
      <div className="grid gap-3 sm:grid-cols-2">
        <Card hoverable size="small" className="border-green-300" onClick={() => onChoose('balance')}>
          <p className="m-0 flex items-center gap-2 font-semibold text-green-800">
            <WalletOutlined /> ลูกค้าชำระส่วนที่เหลือ
          </p>
          <p className="mt-1 mb-0 text-[13px] text-ink-muted">แนบสลิปยอดคงเหลือ {formatBaht(remaining)} ให้เซลล์ยืนยัน แล้วฝ่ายการเงินออกใบเสร็จ</p>
        </Card>
        <Card hoverable size="small" onClick={() => onChoose('edit')}>
          <p className="m-0 flex items-center gap-2 font-semibold">
            <EditOutlined /> แก้ไขข้อมูลใบจอง
          </p>
          <p className="mt-1 mb-0 text-[13px] text-ink-muted">เช่น เพิ่ม/ลดผู้เดินทาง เปลี่ยนราคา ข้อมูลลูกค้า หรือกำหนดชำระ</p>
        </Card>
      </div>
    </div>
  );
}
