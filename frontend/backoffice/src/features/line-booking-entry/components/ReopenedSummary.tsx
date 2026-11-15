import { Alert, Button, Card, Collapse, Descriptions, Image, Tag } from 'antd';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { LineBookingStates } from '../../../constants/lineBookingStates.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import LineBookingHistory from '../../line-booking/components/LineBookingHistory.tsx';
import { financeWarning } from '../lib/deal.ts';
import { tripRange } from '../lib/format.ts';

type Props = {
  booking: LineBooking;
  onCancel: () => void;
  onChangePurpose?: () => void;
};

// The fixed part of a reopened booking: tour, departure, sales owner, the slips and its history.
export default function ReopenedSummary({ booking, onCancel, onChangePurpose }: Props) {
  const depositPaid = booking.state === LineBookingStates.DepositPaid;
  const slips = [
    ...booking.slipImages.map((url, i, all) => ({ url, label: `${booking.isDeposit ? 'สลิปมัดจำ' : 'สลิปโอนเงิน'}${all.length > 1 ? ` ${i + 1}` : ''}` })),
    ...(booking.balanceConfirmedAt
      ? booking.balanceSlipImages.map((url, i, all) => ({ url, label: `สลิปส่วนที่เหลือ${all.length > 1 ? ` ${i + 1}` : ''}` }))
      : []),
  ];

  return (
    <Card
      size="small"
      title={
        <span className="flex flex-wrap items-center gap-2">
          <span className="font-mono">{booking.bookingNo}</span>
          <Tag color={depositPaid ? 'blue' : 'green'} className="m-0">
            {depositPaid ? 'ชำระมัดจำแล้ว' : 'ชำระครบแล้ว'}
          </Tag>
        </span>
      }
      extra={
        <span className="flex gap-1">
          {onChangePurpose && (
            <Button type="link" size="small" onClick={onChangePurpose}>
              เปลี่ยนสิ่งที่ต้องการทำ
            </Button>
          )}
          <Button type="link" size="small" danger onClick={onCancel}>
            ยกเลิก
          </Button>
        </span>
      }
    >
      <div className="flex flex-col gap-3">
        <p className="m-0 font-medium leading-snug">
          <span className="mr-2 font-mono text-[13px] text-ink-muted">{booking.tourCode}</span>
          {booking.tourName}
        </p>
        <Descriptions size="small" column={{ xs: 1, sm: 3 }} colon={false}>
          <Descriptions.Item label="วันเดินทาง">{tripRange(booking.startDate, booking.endDate)}</Descriptions.Item>
          <Descriptions.Item label="เซลล์">{booking.salesUser?.name ?? '-'}</Descriptions.Item>
          <Descriptions.Item label="ยอดเดิม">{formatBaht(booking.totalPrice)}</Descriptions.Item>
        </Descriptions>
        {slips.length > 0 && (
          <Image.PreviewGroup>
            <div className="flex flex-wrap gap-2">
              {slips.map((s) => (
                <Image key={s.url} src={s.url} alt={s.label} width={64} height={84} className="rounded-md object-cover" />
              ))}
            </div>
          </Image.PreviewGroup>
        )}
        {!booking.documentsApprovedAt && <Alert type="warning" showIcon title={financeWarning(booking)} />}
        <Collapse
          size="small"
          items={[
            {
              key: 'history',
              label: 'ประวัติใบจอง',
              children: (
                <LineBookingHistory
                  bookingId={booking.id}
                  version={booking.updateDate}
                  load={() => lineBookingsApi.editableEvents(booking.bookingNo)}
                  readOnly
                />
              ),
            },
          ]}
        />
      </div>
    </Card>
  );
}
