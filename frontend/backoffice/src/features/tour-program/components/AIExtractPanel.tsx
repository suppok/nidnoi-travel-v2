import { FilePdfOutlined, InboxOutlined, RobotOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Progress, Upload } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { ApiError } from '../../../api/client.ts';
import { mediaApi } from '../../../api/media.ts';
import { aiExtractApi } from '../../../api/tourPrograms.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { AIExtractResult } from '../../../types/aiExtract.ts';

const POLL_MS = 3000;
// The backend gives up after ai.requestTimeout (180 s); stop polling a little after that.
const GIVE_UP_MS = 4 * 60 * 1000;
// A rough progress bar: most PDFs take 30-90 s.
const EXPECTED_MS = 75 * 1000;

// The same file uploaded as the tour's PDF (purpose tour-pdf), for the form's "ไฟล์ PDF" field.
export type AIPdfUpload = { url: string } | { error: unknown };

type Props = {
  maxPdfSizeMb: number;
  onResult: (result: AIExtractResult, pdf: AIPdfUpload) => void;
};

type Phase = { kind: 'idle' } | { kind: 'reading'; fileName: string; startedAt: number } | { kind: 'error'; message: string };

/** Upload a tour PDF; the backend reads it with AI in the background and this polls for the draft. */
export default function AIExtractPanel({ maxPdfSizeMb, onResult }: Props) {
  const [phase, setPhase] = useState<Phase>({ kind: 'idle' });
  const [now, setNow] = useState(() => Date.now());
  const cancelled = useRef(false);

  useEffect(() => () => void (cancelled.current = true), []);
  useEffect(() => {
    if (phase.kind !== 'reading') return;
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, [phase.kind]);

  const fail = (err: unknown) => {
    let text = errorMessage(err);
    if (err instanceof ApiError && err.code === ErrorCodes.MediaTooLarge) text = `ไฟล์ใหญ่เกินไป (ไม่เกิน ${maxPdfSizeMb} MB)`;
    if (err instanceof ApiError && err.code === ErrorCodes.MediaTypeNotAllowed) text = 'รองรับเฉพาะไฟล์ PDF';
    if (err instanceof ApiError && err.code === ErrorCodes.NotFound) text = 'ไม่พบงานอ่านไฟล์นี้แล้ว (เซิร์ฟเวอร์อาจรีสตาร์ท) กรุณาอัปโหลดใหม่';
    setPhase({ kind: 'error', message: text });
  };

  const read = async (file: File) => {
    cancelled.current = false;
    const startedAt = Date.now();
    setNow(startedAt);
    setPhase({ kind: 'reading', fileName: file.name, startedAt });
    // In parallel with the AI: an upload that fails never stops the prefill.
    const pdf: Promise<AIPdfUpload> = mediaApi.upload(file, 'tour-pdf').then(
      (res) => ({ url: res.url }),
      (error: unknown) => ({ error }),
    );
    try {
      const { jobId } = await aiExtractApi.start(file);
      while (!cancelled.current) {
        await new Promise((r) => setTimeout(r, POLL_MS));
        if (cancelled.current) return;
        const job = await aiExtractApi.result(jobId);
        if (job.state === 'done' && job.result) {
          const upload = await pdf;
          if (cancelled.current) return;
          setPhase({ kind: 'idle' });
          onResult(job.result, upload);
          return;
        }
        if (job.state === 'failed') {
          fail(new ApiError(job.error?.code ?? ErrorCodes.Unknown, job.error?.message ?? '', 502));
          return;
        }
        if (Date.now() - startedAt > GIVE_UP_MS) {
          setPhase({ kind: 'error', message: 'AI ใช้เวลานานเกินไป กรุณาลองใหม่ หรือกรอกข้อมูลเอง' });
          return;
        }
      }
    } catch (err) {
      if (!cancelled.current) fail(err);
    }
  };

  const cancel = () => {
    cancelled.current = true;
    setPhase({ kind: 'idle' });
  };

  if (phase.kind === 'reading') {
    const elapsed = Math.round((now - phase.startedAt) / 1000);
    return (
      <Card>
        <div className="flex flex-col items-center gap-3 py-6 text-center">
          <RobotOutlined className="text-4xl text-brand" />
          <div className="font-medium">AI กำลังอ่านไฟล์ {phase.fileName}</div>
          <div className="text-ink-muted">ใช้เวลาประมาณ 30–90 วินาที — อยู่หน้านี้ต่อได้เลย ({elapsed} วินาที)</div>
          <Progress percent={Math.min(95, Math.round(((now - phase.startedAt) / EXPECTED_MS) * 100))} showInfo={false} status="active" className="max-w-[360px]" />
          <Button onClick={cancel}>ยกเลิก</Button>
        </div>
      </Card>
    );
  }

  return (
    <Card>
      {phase.kind === 'error' && <Alert className="mb-4" type="error" showIcon message="อ่านไฟล์ไม่สำเร็จ" description={phase.message} />}
      <Upload.Dragger
        accept=".pdf,application/pdf"
        multiple={false}
        showUploadList={false}
        beforeUpload={(file) => {
          if (file.type && file.type !== 'application/pdf' && !file.name.toLowerCase().endsWith('.pdf')) {
            setPhase({ kind: 'error', message: 'รองรับเฉพาะไฟล์ PDF' });
          } else if (file.size > maxPdfSizeMb * 1024 * 1024) {
            setPhase({ kind: 'error', message: `ไฟล์ใหญ่เกินไป (ไม่เกิน ${maxPdfSizeMb} MB)` });
          } else {
            void read(file);
          }
          return Upload.LIST_IGNORE;
        }}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">ลากไฟล์ PDF โปรแกรมทัวร์มาวาง หรือคลิกเพื่อเลือกไฟล์</p>
        <p className="ant-upload-hint">
          <FilePdfOutlined /> PDF ไม่เกิน {maxPdfSizeMb} MB — AI จะกรอกฟอร์มด้านล่างให้ ตรวจสอบและแก้ไขก่อนบันทึกเสมอ
        </p>
      </Upload.Dragger>
    </Card>
  );
}
