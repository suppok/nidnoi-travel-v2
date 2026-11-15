import { Alert, Button, Input } from 'antd';
import { useState } from 'react';
import { type ParsedSummary, parseSummary } from '../lib/parseSummary.ts';

const NOT_A_SUMMARY = 'ไม่พบข้อมูลการจอง ตรวจสอบว่าวางข้อความสรุปจากเครื่องมือ "สรุปยอดโอน" ครบทั้งข้อความ';

type Props = {
  onApply: (summary: ParsedSummary) => void | Promise<void>;
  rows?: number;
  autoFocus?: boolean;
  // Extra buttons next to "กรอกข้อมูลจากข้อความ" (e.g. skip).
  extra?: React.ReactNode;
};

// A text box that fills the form from a "สรุปยอดโอน" text: a paste is applied at once; typed text
// only with the button.
export default function SummaryPasteBox({ onApply, rows = 8, autoFocus, extra }: Props) {
  const [text, setText] = useState('');
  const [error, setError] = useState('');
  const [applying, setApplying] = useState(false);

  const apply = async (value: string) => {
    const parsed = parseSummary(value);
    if (!parsed) {
      setError(NOT_A_SUMMARY);
      return;
    }
    setError('');
    setApplying(true);
    try {
      await onApply(parsed);
      setText('');
    } finally {
      setApplying(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <Input.TextArea
        value={text}
        onChange={(e) => setText(e.target.value)}
        onPaste={(e) => {
          const pasted = e.clipboardData.getData('text');
          if (pasted.trim()) {
            e.preventDefault();
            setText(pasted);
            void apply(pasted);
          }
        }}
        rows={rows}
        autoFocus={autoFocus}
        disabled={applying}
        placeholder="🔴สรุปรายละเอียดการจอง🔴 …"
      />
      {error && <Alert type="error" showIcon title={error} />}
      <div className="flex flex-wrap gap-2">
        <Button type="primary" loading={applying} disabled={!text.trim()} onClick={() => void apply(text)}>
          กรอกข้อมูลจากข้อความ
        </Button>
        {extra}
      </div>
    </div>
  );
}
