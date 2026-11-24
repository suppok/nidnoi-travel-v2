import { ArrowLeftOutlined, EditOutlined, FilePdfOutlined } from '@ant-design/icons';
import { Alert, App, Button, Segmented, Tag } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { aiExtractApi, periodsApi, tourProgramsApi } from '../../../api/tourPrograms.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import { Permissions } from '../../../constants/permissions.ts';
import AIExtractPanel from '../../../features/tour-program/components/AIExtractPanel.tsx';
import DraftPeriodTable from '../../../features/tour-program/components/DraftPeriodTable.tsx';
import { type DraftPeriodRow, draftPeriodProblem, newDraftRowKey } from '../../../features/tour-program/lib/draftPeriods.ts';
import TourProgramForm, { type AIPrefill, type FormValues } from '../../../features/tour-program/components/TourProgramForm.tsx';
import type { AIConfidence, AIConfidenceLevel, AIExtractConfig, AIExtractResult } from '../../../types/aiExtract.ts';

type Mode = 'manual' | 'pdf';

const CONFIDENCE_SECTIONS: { key: keyof AIConfidence; label: string; fields: (keyof FormValues)[] }[] = [
  { key: 'main', label: 'ข้อมูลโปรแกรม', fields: ['name', 'days', 'nights', 'airlineId', 'wholesaleTourCode'] },
  { key: 'taxonomy', label: 'ปลายทาง', fields: ['countryIds', 'cityIds'] },
  { key: 'periods', label: 'วันเดินทาง', fields: [] },
  { key: 'prices', label: 'ราคา', fields: [] },
  { key: 'extras', label: 'ข้อมูลเพิ่มเติม', fields: ['internalNote'] },
];
const LEVEL: Record<AIConfidenceLevel, { color: string; text: string }> = {
  high: { color: 'green', text: 'มั่นใจ' },
  medium: { color: 'gold', text: 'พอใช้' },
  low: { color: 'red', text: 'ไม่มั่นใจ' },
};
// The wholesale is never in the PDF: staff always pick it, so it isn't flagged.
const REQUIRED: (keyof FormValues)[] = ['name', 'days', 'nights', 'countryIds'];

/** "periods[3].price" → "periods"; "countryIds" → "countryIds". */
const fieldRoot = (field: string) => field.split(/[[.]/)[0];

function toPrefill(result: AIExtractResult, version: number): AIPrefill {
  const d = result.draft;
  const values: Partial<FormValues> = {
    name: d.name,
    wholesaleId: d.wholesaleId ?? '',
    wholesaleTourCode: d.wholesaleTourCode,
    state: d.state,
    days: d.days,
    nights: d.nights,
    airlineId: d.airlineId,
    highlight: d.highlight,
    description: d.description,
    internalNote: d.internalNote,
    countryIds: d.countryIds,
    cityIds: d.cityIds,
  };
  const filled = (v: unknown) => (Array.isArray(v) ? v.length > 0 : v !== null && v !== undefined && v !== '');
  const aiFields = new Set(Object.entries(values).filter(([k, v]) => k !== 'state' && filled(v)).map(([k]) => k));
  const attention = new Set<string>(result.warnings.map((w) => fieldRoot(w.field)).filter(Boolean));
  for (const f of REQUIRED) if (!filled(values[f])) attention.add(f);
  for (const s of CONFIDENCE_SECTIONS) if (result.confidence[s.key] === 'low') s.fields.forEach((f) => attention.add(f));
  return { version, values, aiFields, attentionFields: attention };
}

function toRows(result: AIExtractResult): DraftPeriodRow[] {
  return result.draft.periods.map((value, i) => ({
    key: newDraftRowKey(),
    value,
    warnings: result.warnings.filter((w) => w.field.startsWith(`periods[${i}]`)).map((w) => w.message),
  }));
}

export default function TourProgramCreatePage() {
  const { message, modal } = App.useApp();
  const navigate = useNavigate();
  const { can } = useAuth();
  const canAI = can(Permissions.TourProgramAIExtract);

  const [aiConfig, setAIConfig] = useState<AIExtractConfig | null>(null);
  const [mode, setMode] = useState<Mode>('manual');
  const [result, setResult] = useState<AIExtractResult | null>(null);
  const [prefill, setPrefill] = useState<AIPrefill | null>(null);
  const [periodRows, setPeriodRows] = useState<DraftPeriodRow[]>([]);

  // The AI option shows only with the permission and when an admin has switched AI on.
  useEffect(() => {
    if (!canAI) return;
    aiExtractApi
      .config()
      .then(setAIConfig)
      .catch(() => setAIConfig(null));
  }, [canAI]);
  const aiReady = canAI && !!aiConfig?.ready;

  const apply = (r: AIExtractResult) => {
    setResult(r);
    setPrefill((prev) => toPrefill(r, (prev?.version ?? 0) + 1));
    setPeriodRows(toRows(r));
    message.success('AI กรอกข้อมูลให้แล้ว — ตรวจสอบทุกช่องก่อนบันทึก');
  };

  const onResult = (r: AIExtractResult) => {
    if (!prefill) return apply(r);
    modal.confirm({
      title: 'แทนที่ข้อมูลในฟอร์ม?',
      content: 'ข้อมูลจากไฟล์ใหม่จะแทนที่สิ่งที่กรอกไว้และรอบเดินทางทั้งหมด',
      okText: 'แทนที่',
      cancelText: 'ยกเลิก',
      onOk: () => apply(r),
    });
  };

  const periodAttention = !!result && (result.confidence.periods === 'low' || result.confidence.prices === 'low');
  const documentWarnings = useMemo(() => result?.warnings ?? [], [result]);

  const onSubmit = async (body: Parameters<typeof tourProgramsApi.create>[0]) => {
    const created = await tourProgramsApi.create(body);
    if (periodRows.length === 0) {
      message.success(`สร้างโปรแกรม ${created.code} แล้ว — เพิ่มรอบเดินทางได้ด้านล่าง`);
      navigate(`/tour-programs/${created.id}/edit#periods`, { replace: true });
      return;
    }
    // The tour exists now; periods are saved one by one. Failures are left for the edit page.
    let failed = 0;
    for (const { value: p } of periodRows) {
      try {
        await periodsApi.create({
          tourProgramId: created.id,
          startDate: p.startDate,
          endDate: p.endDate,
          price: p.price ?? 0,
          priceBeforeDiscount: p.priceBeforeDiscount,
          childWithBedPrice: p.childWithBedPrice,
          childNoBedPrice: p.childNoBedPrice,
          infantPrice: p.infantPrice,
          singleSupplement: p.singleSupplement,
          deposit: p.deposit,
          seatsTotal: p.seatsTotal ?? 0,
          seatsAvailable: p.seatsAvailable ?? 0,
          note: p.note.trim(),
        });
      } catch {
        failed++;
      }
    }
    if (failed === 0) message.success(`สร้างโปรแกรม ${created.code} พร้อมรอบเดินทาง ${periodRows.length} รอบแล้ว`);
    else message.warning(`สร้างโปรแกรม ${created.code} แล้ว แต่บันทึกรอบเดินทางไม่สำเร็จ ${failed} จาก ${periodRows.length} รอบ — เพิ่มเองในตารางด้านล่าง`, 8);
    navigate(`/tour-programs/${created.id}/edit#periods`, { replace: true });
  };

  return (
    <>
      <Link to="/tour-programs" className="inline-flex w-fit items-center gap-1.5 text-ink-muted no-underline hover:text-ink">
        <ArrowLeftOutlined /> โปรแกรมทัวร์ทั้งหมด
      </Link>
      <PageHeader title="เพิ่มโปรแกรมทัวร์" subtitle="รหัสทัวร์จะถูกสร้างอัตโนมัติหลังบันทึก จากนั้นเพิ่มรอบเดินทางและราคาได้" />

      {aiReady && (
        <div className="mb-4 flex flex-col gap-4">
          <Segmented<Mode>
            value={mode}
            onChange={setMode}
            className="self-start"
            options={[
              { value: 'manual', label: 'กรอกเอง', icon: <EditOutlined /> },
              { value: 'pdf', label: 'อัปโหลด PDF (AI ช่วยกรอก)', icon: <FilePdfOutlined /> },
            ]}
          />
          {mode === 'pdf' && <AIExtractPanel maxPdfSizeMb={aiConfig?.maxPdfSizeMb ?? 40} onResult={onResult} />}
        </div>
      )}

      {result && (
        <Alert
          className="mb-4"
          type={documentWarnings.length > 0 ? 'warning' : 'info'}
          showIcon
          message={`AI กรอกข้อมูลจาก ${result.fileName} — ข้อมูลยังไม่ถูกบันทึก ตรวจสอบกับไฟล์ต้นฉบับทุกช่องก่อนกดสร้าง`}
          description={
            <div className="flex flex-col gap-2">
              <div className="flex flex-wrap items-center gap-1.5 text-[13px]">
                ความมั่นใจของ AI:
                {CONFIDENCE_SECTIONS.map((s) => (
                  <Tag key={s.key} color={LEVEL[result.confidence[s.key]].color} className="m-0">
                    {s.label}: {LEVEL[result.confidence[s.key]].text}
                  </Tag>
                ))}
              </div>
              {documentWarnings.length > 0 && (
                <ul className="m-0 pl-5">
                  {documentWarnings.map((w, i) => (
                    <li key={i}>{w.message}</li>
                  ))}
                </ul>
              )}
              <div className="text-xs text-ink-muted">
                {result.model} · {Math.round(result.durationMs / 1000)} วินาที · {result.usage.inputTokens.toLocaleString()} / {result.usage.outputTokens.toLocaleString()} tokens
              </div>
            </div>
          }
        />
      )}

      <TourProgramForm
        tour={null}
        prefill={prefill}
        submitText={periodRows.length > 0 ? `สร้างโปรแกรมและรอบเดินทาง ${periodRows.length} รอบ` : 'สร้างโปรแกรม'}
        beforeSubmit={() => draftPeriodProblem(periodRows)}
        extraSections={
          periodRows.length > 0 || result ? (
            <DraftPeriodTable rows={periodRows} onChange={setPeriodRows} days={prefill?.values.days ?? null} attention={periodAttention} />
          ) : null
        }
        extraActions={
          <Link to="/tour-programs">
            <Button>ยกเลิก</Button>
          </Link>
        }
        onSubmit={onSubmit}
      />
    </>
  );
}
