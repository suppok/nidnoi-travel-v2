import { FilePdfOutlined, LoadingOutlined, SendOutlined, UploadOutlined } from '@ant-design/icons';
import { Alert, App, Button, Descriptions, Typography, Upload } from 'antd';
import { useState } from 'react';
import { ApiError } from '../../../api/client.ts';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { LineBooking, LineBookingDocumentKey } from '../../../types/lineBooking.ts';
import { fileNameFromUrl } from '../../tour-program/components/mediaUpload.ts';
import { requiredDocuments } from '../lib/lineBookingFormat.ts';

const MAX_UPLOAD_BYTES = 10 * 1024 * 1024;

type Props = {
  booking: LineBooking;
  // The caller is the booking's finance user (or an admin) and may upload/send.
  canSend: boolean;
  onSent: (updated: LineBooking) => void;
  // Set for the sales owner while the documents wait for their review: reopens the review popup.
  onReview?: () => void;
};

type Files = Record<LineBookingDocumentKey, string>;

// "เอกสารการเงิน" of a confirmed booking: deposit → AI + IV, full payment → RE (PDF only). Sending
// the first time alerts the sales owner.
export default function FinanceDocumentsSection({ booking, canSend, onSent, onReview }: Props) {
  const { message } = App.useApp();
  const docs = requiredDocuments(booking);
  const [files, setFiles] = useState<Files>({ aiFile: booking.aiFile, ivFile: booking.ivFile, reFile: booking.reFile });
  const [uploading, setUploading] = useState<LineBookingDocumentKey | null>(null);
  const [sending, setSending] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const approved = booking.documentsApprovedAt !== null;
  // Finance may change files until the sales owner approves them.
  const editable = canSend && !approved;
  const changed = docs.some((d) => files[d.key] !== booking[d.key]);
  const complete = docs.every((d) => files[d.key]);

  const send = async () => {
    setSending(true);
    setFieldErrors({});
    try {
      const updated = await lineBookingsApi.documents({ id: booking.id, aiFile: files.aiFile, ivFile: files.ivFile, reFile: files.reFile });
      message.success(booking.documentsAt ? 'บันทึกเอกสารแล้ว' : `ส่งเอกสารแล้ว — แจ้ง ${updated.salesUser?.name ?? 'เซลล์'} แล้ว`);
      onSent(updated);
    } catch (err) {
      // VALIDATION on aiFile/ivFile/reFile shows under that row; anything else as a message.
      if (err instanceof ApiError && err.code === ErrorCodes.Validation && err.fields && docs.some((d) => err.fields?.[d.key])) {
        setFieldErrors(err.fields);
      } else {
        message.error(errorMessage(err));
      }
    } finally {
      setSending(false);
    }
  };

  return (
    <section>
      <Typography.Title level={5}>เอกสารการเงิน</Typography.Title>
      {!booking.documentsAt && booking.fixRequestedAt && booking.fixNote && (
        <Alert
          type="error"
          showIcon
          className="mb-3"
          title={`${booking.salesUser?.name ?? 'เซลล์'} ขอให้แก้ไข (${formatDateTime(booking.fixRequestedAt)})`}
          description={<span className="whitespace-pre-wrap">{booking.fixNote}</span>}
        />
      )}
      {booking.documentsAt && !approved && (
        <Alert
          type="warning"
          showIcon
          className="mb-3"
          title={`รอ ${booking.salesUser?.name ?? 'เซลล์'} ตรวจเอกสารและส่งให้ลูกค้า`}
          action={
            onReview && (
              <Button size="small" type="primary" onClick={onReview}>
                ตรวจเอกสาร
              </Button>
            )
          }
        />
      )}
      {approved && (
        <Alert
          type="success"
          showIcon
          className="mb-3"
          title={
            booking.customerSentAt
              ? `ส่งให้ลูกค้าทาง LINE แล้ว (${formatDateTime(booking.customerSentAt)})`
              : `เซลล์ยืนยันเอกสารแล้ว (${formatDateTime(booking.documentsApprovedAt ?? '')}) — ไม่ได้ส่งทาง LINE`
          }
        />
      )}
      {!booking.documentsAt && (
        <Alert
          type="info"
          showIcon
          className="mb-3"
          title={canSend ? 'รอคุณอัปโหลดเอกสาร' : `รอฝ่ายการเงิน${booking.financeUser ? ` (${booking.financeUser.name})` : ''} ส่งเอกสาร`}
          description={docs.map((d) => `${d.code} ${d.label}`).join(' และ ') + ' (PDF)'}
        />
      )}
      <Descriptions column={1} size="small" bordered styles={{ label: { width: 150, whiteSpace: 'nowrap' } }}>
        {docs.map((d) => {
          const url = files[d.key];
          return (
            <Descriptions.Item key={d.key} label={`${d.code} · ${d.label}`}>
              <div className="flex flex-wrap items-center gap-2">
                {url ? (
                  <a href={url} target="_blank" rel="noreferrer" className="inline-flex max-w-[300px] items-center gap-2">
                    <FilePdfOutlined className="text-red-600" />
                    <span className="truncate">{fileNameFromUrl(url)}</span>
                  </a>
                ) : (
                  <span className="text-ink-muted">ยังไม่มี</span>
                )}
                {editable && (
                  <Upload
                    accept="application/pdf,.pdf"
                    showUploadList={false}
                    disabled={uploading !== null || sending}
                    beforeUpload={(file) => {
                      const pdf = file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf');
                      const problem = !pdf ? 'อัปโหลดได้เฉพาะไฟล์ PDF' : file.size > MAX_UPLOAD_BYTES ? 'ไฟล์ใหญ่เกินไป (ไม่เกิน 10 MB)' : null;
                      if (problem) message.error(problem);
                      return problem ? Upload.LIST_IGNORE : true;
                    }}
                    customRequest={({ file, onSuccess, onError }) => {
                      setUploading(d.key);
                      lineBookingsApi
                        .uploadDocument(file as File)
                        .then((res) => {
                          setFiles((f) => ({ ...f, [d.key]: res.url }));
                          setFieldErrors((e) => ({ ...e, [d.key]: '' }));
                          onSuccess?.(res);
                        })
                        .catch((err: unknown) => {
                          message.error(errorMessage(err));
                          onError?.(err instanceof Error ? err : new Error(String(err)));
                        })
                        .finally(() => setUploading(null));
                    }}
                  >
                    <Button size="small" icon={uploading === d.key ? <LoadingOutlined /> : <UploadOutlined />} disabled={uploading !== null || sending}>
                      {uploading === d.key ? 'กำลังอัปโหลด…' : url ? 'เปลี่ยนไฟล์' : 'อัปโหลด PDF'}
                    </Button>
                  </Upload>
                )}
              </div>
              {fieldErrors[d.key] && <div className="mt-1 text-sm text-red-600">{fieldErrors[d.key]}</div>}
            </Descriptions.Item>
          );
        })}
        {booking.documentsAt && <Descriptions.Item label="ส่งเอกสารเมื่อ">{formatDateTime(booking.documentsAt)}</Descriptions.Item>}
      </Descriptions>
      {editable && (changed || !booking.documentsAt) && (
        <div className="mt-3 flex justify-end">
          <Button type="primary" icon={<SendOutlined />} loading={sending} disabled={!complete || uploading !== null} onClick={() => void send()}>
            {booking.documentsAt ? 'บันทึกเอกสาร' : 'ส่งเอกสารให้เซลล์'}
          </Button>
        </div>
      )}
    </section>
  );
}
