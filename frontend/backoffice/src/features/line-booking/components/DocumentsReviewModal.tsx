import { FilePdfOutlined, SendOutlined, WarningOutlined } from '@ant-design/icons';
import { Alert, App, Avatar, Button, Input, Modal } from 'antd';
import { useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { LineBooking } from '../../../types/lineBooking.ts';
import { fileNameFromUrl } from '../../tour-program/components/mediaUpload.ts';
import { requiredDocuments } from '../lib/lineBookingFormat.ts';

type Props = {
  booking: LineBooking | null; // open when set
  onClose: () => void;
  onDone: (updated: LineBooking) => void;
};

// The sales owner's review of the finance documents: recheck the PDFs, then send them to the
// customer's LINE chat (or just approve when no chat is linked), or send them back with a note.
export default function DocumentsReviewModal({ booking, onClose, onDone }: Props) {
  const { message } = App.useApp();
  const [fixing, setFixing] = useState(false);
  const [note, setNote] = useState('');
  const [busy, setBusy] = useState<'approve' | 'reject' | null>(null);
  if (!booking) return null;

  // Start fresh next time it opens.
  const close = () => {
    setFixing(false);
    setNote('');
    onClose();
  };

  const docs = requiredDocuments(booking);
  const chat = booking.lineContact;
  const canPush = chat !== null && chat.following;

  const approve = async () => {
    setBusy('approve');
    try {
      const updated = await lineBookingsApi.approveDocuments(booking.id);
      message.success(updated.customerSentAt ? `ส่งเอกสารให้ ${chat?.displayName ?? 'ลูกค้า'} ทาง LINE แล้ว` : 'ยืนยันเอกสารแล้ว (ไม่ได้ส่ง LINE)');
      setFixing(false);
      setNote('');
      onDone(updated);
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setBusy(null);
    }
  };

  const reject = async () => {
    setBusy('reject');
    try {
      const updated = await lineBookingsApi.rejectDocuments(booking.id, note.trim());
      message.success(`ส่งกลับให้ ${updated.financeUser?.name ?? 'ฝ่ายการเงิน'} แก้ไขแล้ว`);
      setFixing(false);
      setNote('');
      onDone(updated);
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setBusy(null);
    }
  };

  return (
    <Modal
      open
      title={`เอกสารการเงินพร้อมแล้ว · ${booking.bookingNo}`}
      onCancel={close}
      destroyOnHidden
      footer={
        fixing ? (
          <div className="flex justify-end gap-2">
            <Button onClick={() => setFixing(false)} disabled={busy !== null}>
              กลับ
            </Button>
            <Button danger type="primary" loading={busy === 'reject'} disabled={!note.trim()} onClick={() => void reject()}>
              ส่งกลับให้ฝ่ายการเงิน
            </Button>
          </div>
        ) : (
          <div className="flex flex-wrap justify-end gap-2">
            <Button onClick={close} disabled={busy !== null}>
              ไว้ทีหลัง
            </Button>
            <Button icon={<WarningOutlined />} onClick={() => setFixing(true)} disabled={busy !== null}>
              แจ้งแก้ไข
            </Button>
            <Button type="primary" icon={<SendOutlined />} loading={busy === 'approve'} onClick={() => void approve()}>
              {canPush ? 'ส่งให้ลูกค้า' : 'ยืนยันเอกสาร'}
            </Button>
          </div>
        )
      }
    >
      <p className="mb-2 text-ink-muted">
        {booking.financeUser?.name ?? 'ฝ่ายการเงิน'} ส่งเอกสารมาแล้ว กดเพื่อเปิดตรวจสอบก่อนส่งให้ลูกค้า
      </p>
      <ul className="mb-4 flex flex-col gap-2">
        {docs.map((d) => {
          const url = booking[d.key];
          return (
            <li key={d.key}>
              <a href={url} target="_blank" rel="noreferrer" className="inline-flex items-center gap-2">
                <FilePdfOutlined className="text-red-600" />
                <strong>
                  {d.code} {d.label}
                </strong>
                <span className="truncate text-ink-muted">{fileNameFromUrl(url)}</span>
              </a>
            </li>
          );
        })}
      </ul>

      {fixing ? (
        <>
          <p className="mb-2">สิ่งที่ต้องแก้ไข</p>
          <Input.TextArea
            autoFocus
            rows={4}
            maxLength={1000}
            showCount
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="เช่น ชื่อลูกค้าในใบรับเงินมัดจำสะกดผิด / ยอดในใบแจ้งหนี้ไม่ตรง"
          />
        </>
      ) : canPush && chat ? (
        <div className="flex items-center gap-3 rounded-lg border border-line px-3 py-2">
          <Avatar src={chat.pictureUrl ?? undefined}>{chat.displayName.slice(0, 1)}</Avatar>
          <div className="flex flex-col">
            <span className="text-xs text-ink-muted">จะส่งทางแชท LINE ของลูกค้า</span>
            <span className="font-medium">{chat.displayName}</span>
          </div>
        </div>
      ) : (
        <Alert
          type="info"
          showIcon
          title={chat ? `${chat.displayName} บล็อก LINE OA แล้ว — ส่งทาง LINE ไม่ได้` : 'ใบจองนี้ยังไม่ได้เชื่อมแชท LINE ของลูกค้า'}
          description="กดยืนยันเอกสารเพื่อปิดงานโดยไม่ส่ง LINE (ส่งเอกสารให้ลูกค้าเองได้)"
        />
      )}
    </Modal>
  );
}
