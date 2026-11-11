import { useState } from 'react'
import { type ParsedSummary, parseSummary } from '../lib/parseSummary.ts'

export type PasteReport = { filled: string[]; warnings: string[] }

type ApplyFn = (summary: ParsedSummary) => Promise<void>

const NOT_A_SUMMARY = 'ไม่พบข้อมูลการจอง ตรวจสอบว่าวางข้อความสรุปจากเครื่องมือฝ่ายขายครบทั้งข้อความ'

// A paste lands as one big change; typing adds a character or two. Only a paste is applied at once.
const PASTE_MIN_CHARS = 20

// Textarea that fills the form as soon as a "สรุปยอดโอน" text is pasted into it.
function useAutoPaste(onApply: ApplyFn, onApplied: () => void) {
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const [applying, setApplying] = useState(false)

  async function apply(value: string) {
    const parsed = parseSummary(value)
    if (!parsed) {
      setError(NOT_A_SUMMARY)
      return
    }
    setError('')
    setApplying(true)
    try {
      await onApply(parsed)
      setText('')
      onApplied()
    } finally {
      setApplying(false)
    }
  }

  function onChange(value: string) {
    const pasted = value.length - text.length >= PASTE_MIN_CHARS
    setText(value)
    if (pasted && !applying) void apply(value)
  }

  return { text, error, applying, onChange, apply: () => void apply(text) }
}

const textareaClass =
  'w-full rounded-lg border border-gray-300 bg-white px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none disabled:bg-gray-100'

// First screen of the app: paste a summary to prefill the form, or skip to an empty one.
// onDone shows the form, after a paste was applied or on skip.
export function SummaryPasteStart({ onApply, onDone }: { onApply: ApplyFn; onDone: () => void }) {
  const paste = useAutoPaste(onApply, onDone)

  return (
    <div className="space-y-3">
      <label htmlFor="summary-start" className="block text-sm font-medium text-gray-700">
        วางข้อความสรุปยอดโอน
      </label>
      <p className="text-xs text-gray-500">คัดลอกข้อความจากเครื่องมือ "สรุปยอดโอน" ในหลังบ้าน แล้ววางที่นี่ ระบบจะกรอกฟอร์มให้อัตโนมัติ</p>
      <textarea
        id="summary-start"
        value={paste.text}
        onChange={(e) => paste.onChange(e.target.value)}
        disabled={paste.applying}
        rows={10}
        autoFocus
        placeholder="🔴สรุปรายละเอียดการจอง🔴 …"
        className={textareaClass}
      />
      {paste.applying && <p className="text-sm text-blue-700">กำลังกรอกข้อมูล...</p>}
      {paste.error && <p className="text-xs text-red-600">{paste.error}</p>}
      <div className="flex gap-2">
        {paste.text.trim() && !paste.applying && (
          <button
            type="button"
            onClick={paste.apply}
            className="flex-1 rounded-lg bg-blue-600 py-2.5 text-sm font-semibold text-white transition hover:bg-blue-700"
          >
            กรอกข้อมูลจากข้อความ
          </button>
        )}
        <button
          type="button"
          onClick={onDone}
          disabled={paste.applying}
          className="flex-1 rounded-lg border border-gray-300 bg-white py-2.5 text-sm text-gray-600 disabled:opacity-50"
        >
          ข้าม กรอกเอง
        </button>
      </div>
    </div>
  )
}

// Inside the form: paste another summary, and the report of what the last paste filled.
export function SummaryPaste({ onApply, report }: { onApply: ApplyFn; report: PasteReport | null }) {
  const [open, setOpen] = useState(false)
  const paste = useAutoPaste(onApply, () => setOpen(false))

  return (
    <section className="rounded-lg border border-blue-200 bg-blue-50/60 p-3">
      {open ? (
        <div className="space-y-2">
          <label htmlFor="summary-text" className="block text-sm font-medium text-gray-700">
            วางข้อความสรุปการจอง
          </label>
          <textarea
            id="summary-text"
            value={paste.text}
            onChange={(e) => paste.onChange(e.target.value)}
            disabled={paste.applying}
            rows={6}
            autoFocus
            placeholder="🔴สรุปรายละเอียดการจอง🔴 …"
            className={textareaClass}
          />
          {paste.applying && <p className="text-xs text-blue-700">กำลังกรอกข้อมูล...</p>}
          {paste.error && <p className="text-xs text-red-600">{paste.error}</p>}
          <div className="flex gap-2">
            <button
              type="button"
              onClick={paste.apply}
              disabled={paste.applying || !paste.text.trim()}
              className="flex-1 rounded-lg bg-blue-600 py-2.5 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:opacity-50"
            >
              กรอกข้อมูลจากข้อความ
            </button>
            <button
              type="button"
              onClick={() => setOpen(false)}
              className="rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm text-gray-600"
            >
              ยกเลิก
            </button>
          </div>
        </div>
      ) : (
        <button type="button" onClick={() => setOpen(true)} className="w-full text-left text-sm font-medium text-blue-700">
          📋 วางข้อความสรุปยอดโอนใหม่
        </button>
      )}

      {report && !open && (
        <div className="mt-2 space-y-1 text-xs">
          {report.filled.length > 0 && <p className="text-green-700">กรอกแล้ว: {report.filled.join(', ')}</p>}
          {report.warnings.map((w) => (
            <p key={w} className="text-amber-700">
              ⚠ {w}
            </p>
          ))}
          <p className="text-gray-500">ตรวจสอบข้อมูลอีกครั้ง แล้วอัปโหลดสลิปก่อนบันทึก</p>
        </div>
      )}
    </section>
  )
}
