import { Alert } from 'antd';
import type { LineBookingWholesaleEmailResult } from '../../../types/lineBooking.ts';

// What happened to the booking email to the wholesale after a new booking was saved.
export default function WholesaleEmailNotice({ email }: { email: LineBookingWholesaleEmailResult }) {
  const name = email.wholesaleName ? `โฮลเซลล์ ${email.wholesaleName}` : 'โฮลเซลล์';
  if (email.result === 'queued') {
    return <Alert type="success" showIcon className="mt-4 text-left" message={`ส่งอีเมลยืนยันการจองถึง${name}แล้ว`} description={email.to.join(', ')} />;
  }
  if (email.result === 'no_email') {
    return (
      <Alert
        type="warning"
        showIcon
        className="mt-4 text-left"
        message={`ไม่ได้ส่งอีเมลถึง${name}`}
        description="โฮลเซลล์นี้เปิดส่งอีเมลอัตโนมัติไว้ แต่ยังไม่มีอีเมล — เพิ่มอีเมลที่หน้าโฮลเซลล์ แล้วส่งอีเมลเอง"
      />
    );
  }
  return (
    <Alert
      type="info"
      showIcon
      className="mt-4 text-left"
      message={`ไม่ได้ส่งอีเมลถึง${name}`}
      description="โฮลเซลล์นี้ไม่ได้เปิดส่งอีเมลอัตโนมัติหลังบันทึกใบจอง"
    />
  );
}
