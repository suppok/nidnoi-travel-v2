import { CloseOutlined, DeleteOutlined, PlusOutlined, SaveOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, DatePicker, Input, InputNumber, type InputNumberProps, Switch, Table, type TableProps, Tag, Tooltip } from 'antd';
import dayjs from 'dayjs';
import { useCallback, useEffect, useState } from 'react';
import { periodsApi } from '../../../api/tourPrograms.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatBaht } from '../../../lib/format.ts';
import type { TourProgramPeriod, TourProgramPeriodFields } from '../../../types/tourProgram.ts';

const DATE = 'YYYY-MM-DD';
const NEW_PREFIX = 'new:';

const SAVED_SEATS = 'บันทึกแล้ว — หน้าเว็บอัปเดตภายในประมาณ 30 วินาที';
const SAVED_CONTENT = 'บันทึกแล้ว — หน้าเว็บจะอัปเดตภายในไม่กี่นาที';

type Row = { key: string; period: TourProgramPeriod | null; draft: TourProgramPeriodFields };
type FieldErrors = Partial<Record<keyof TourProgramPeriodFields, string>>;

const PRICE_KEYS = ['price', 'priceBeforeDiscount', 'childWithBedPrice', 'childNoBedPrice', 'singleSupplement', 'deposit'] as const;

function toFields(p: TourProgramPeriod): TourProgramPeriodFields {
  return {
    startDate: p.startDate,
    endDate: p.endDate,
    price: p.price,
    priceBeforeDiscount: p.priceBeforeDiscount,
    childWithBedPrice: p.childWithBedPrice,
    childNoBedPrice: p.childNoBedPrice,
    singleSupplement: p.singleSupplement,
    deposit: p.deposit,
    seatsTotal: p.seatsTotal,
    seatsAvailable: p.seatsAvailable,
    note: p.note,
  };
}

function omit<T>(record: Record<string, T>, key: string): Record<string, T> {
  const next = { ...record };
  delete next[key];
  return next;
}

const sameFields =(a: TourProgramPeriodFields, b: TourProgramPeriodFields) =>
  (Object.keys(a) as (keyof TourProgramPeriodFields)[]).every((k) => a[k] === b[k]);

// Only seats changed → served live by the period island, no site rebuild.
const onlySeatsChanged = (before: TourProgramPeriodFields, after: TourProgramPeriodFields) =>
  sameFields({ ...before, seatsTotal: 0, seatsAvailable: 0, note: '' }, { ...after, seatsTotal: 0, seatsAvailable: 0, note: '' });

function checkDraft(d: TourProgramPeriodFields): FieldErrors {
  const errors: FieldErrors = {};
  if (!d.startDate || !d.endDate) errors.startDate = 'เลือกวันเดินทาง';
  else if (d.endDate < d.startDate) errors.startDate = 'วันกลับต้องไม่ก่อนวันไป';
  if (!Number.isInteger(d.price) || d.price < 0) errors.price = 'กรอกราคา';
  if (d.priceBeforeDiscount !== null && d.priceBeforeDiscount <= d.price) errors.priceBeforeDiscount = 'ต้องมากกว่าราคาขาย';
  if (d.seatsAvailable > d.seatsTotal) errors.seatsAvailable = 'ที่ว่างต้องไม่เกินที่นั่งทั้งหมด';
  return errors;
}

const money: InputNumberProps<number> = {
  min: 0,
  max: 10_000_000,
  precision: 0,
  controls: false,
  formatter: (v) => `${v ?? ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ','),
  parser: (v) => Number((v ?? '').replace(/,/g, '')),
  className: 'w-full',
};

type Props = {
  tourProgramId: string;
  days: number; // fills the return date when only the departure date is picked
  disabled?: boolean; // tour soft-deleted: read only
};

export default function PeriodTable({ tourProgramId, days, disabled = false }: Props) {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const canCreate = can(Permissions.PeriodCreate) && !disabled;
  const canUpdate = can(Permissions.PeriodUpdate) && !disabled;
  const canStatus = can(Permissions.PeriodStatus) && !disabled;

  const [periods, setPeriods] = useState<TourProgramPeriod[] | null>(null);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [drafts, setDrafts] = useState<Record<string, TourProgramPeriodFields>>({});
  const [newKeys, setNewKeys] = useState<string[]>([]);
  const [saving, setSaving] = useState<ReadonlySet<string>>(new Set());
  const [rowErrors, setRowErrors] = useState<Record<string, FieldErrors>>({});
  const [showDeleted, setShowDeleted] = useState(false);

  const load = useCallback(() => {
    periodsApi
      .list({ tourProgramId })
      .then((data) => {
        setPeriods(data.items);
        setLoadError(null);
      })
      .catch((err: unknown) => setLoadError(err));
  }, [tourProgramId]);

  useEffect(load, [load]);

  const setDraft = (key: string, patch: Partial<TourProgramPeriodFields>, base: TourProgramPeriodFields) => {
    setDrafts((prev) => ({ ...prev, [key]: { ...(prev[key] ?? base), ...patch } }));
    setRowErrors((prev) => (prev[key] ? omit(prev, key) : prev));
  };

  const dropDraft = (key: string) => {
    setDrafts((prev) => omit(prev, key));
    setRowErrors((prev) => omit(prev, key));
    if (key.startsWith(NEW_PREFIX)) setNewKeys((keys) => keys.filter((k) => k !== key));
  };

  const markSaving = (key: string, on: boolean) =>
    setSaving((prev) => {
      const next = new Set(prev);
      if (on) next.add(key);
      else next.delete(key);
      return next;
    });

  const addRow = () => {
    const last = [...(periods ?? [])].reverse().find((p) => p.status);
    const key = `${NEW_PREFIX}${crypto.randomUUID()}`;
    // New departures usually repeat the latest one's prices and seats.
    const draft: TourProgramPeriodFields = last
      ? { ...toFields(last), startDate: '', endDate: '', seatsAvailable: last.seatsTotal, note: '' }
      : { startDate: '', endDate: '', price: 0, priceBeforeDiscount: null, childWithBedPrice: null, childNoBedPrice: null, singleSupplement: null, deposit: null, seatsTotal: 0, seatsAvailable: 0, note: '' };
    setDrafts((prev) => ({ ...prev, [key]: draft }));
    setNewKeys((keys) => [...keys, key]);
  };

  const save = async (row: Row) => {
    const errors = checkDraft(row.draft);
    if (Object.keys(errors).length) {
      setRowErrors((prev) => ({ ...prev, [row.key]: errors }));
      message.error(Object.values(errors)[0]);
      return;
    }
    markSaving(row.key, true);
    try {
      if (row.period) {
        const saved = await periodsApi.update({ ...row.draft, id: row.period.id });
        setPeriods((prev) => (prev ?? []).map((p) => (p.id === saved.id ? saved : p)));
        message.success(onlySeatsChanged(toFields(row.period), row.draft) ? SAVED_SEATS : SAVED_CONTENT);
      } else {
        const saved = await periodsApi.create({ ...row.draft, tourProgramId });
        setPeriods((prev) => [...(prev ?? []), saved].sort((a, b) => a.startDate.localeCompare(b.startDate)));
        message.success(SAVED_CONTENT);
      }
      dropDraft(row.key);
    } catch (err) {
      if (err instanceof ApiError && err.code === ErrorCodes.Validation && err.fields) {
        setRowErrors((prev) => ({ ...prev, [row.key]: err.fields as FieldErrors }));
      }
      message.error(errorMessage(err));
    } finally {
      markSaving(row.key, false);
    }
  };

  const confirmStatus = (period: TourProgramPeriod) => {
    const remove = period.status;
    const label = `${dayjs(period.startDate).format('D MMM YY')} – ${dayjs(period.endDate).format('D MMM YY')}`;
    modal.confirm({
      title: remove ? `ลบรอบ ${label}?` : `กู้คืนรอบ ${label}?`,
      content: remove ? 'รอบนี้จะหายจากหน้าเว็บ กู้คืนได้ภายหลัง' : 'รอบนี้จะกลับมาแสดงบนหน้าเว็บ',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await periodsApi.updateStatus({ id: period.id, status: !remove });
          setPeriods((prev) => (prev ?? []).map((p) => (p.id === period.id ? { ...p, status: !remove } : p)));
          dropDraft(period.id);
          message.success(SAVED_CONTENT);
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rows: Row[] = [
    ...(periods ?? []).filter((p) => p.status || showDeleted).map((p) => ({ key: p.id, period: p, draft: drafts[p.id] ?? toFields(p) })),
    ...newKeys.map((key) => ({ key, period: null, draft: drafts[key] })),
  ];

  const editable = (row: Row) => (row.period ? canUpdate && row.period.status : canCreate);
  const status = (row: Row, field: keyof TourProgramPeriodFields) => (rowErrors[row.key]?.[field] ? ('error' as const) : undefined);

  const priceColumn = (field: (typeof PRICE_KEYS)[number], title: string, width = 116): NonNullable<TableProps<Row>['columns']>[number] => ({
    title,
    key: field,
    width,
    render: (_, row) => (
      <Tooltip title={rowErrors[row.key]?.[field]}>
        <InputNumber<number>
          {...money}
          size="small"
          value={row.draft[field]}
          status={status(row, field)}
          disabled={!editable(row)}
          placeholder="—"
          onChange={(v) => setDraft(row.key, { [field]: field === 'price' ? (v ?? 0) : v }, row.draft)}
        />
      </Tooltip>
    ),
  });

  const columns: TableProps<Row>['columns'] = [
    {
      title: 'วันเดินทาง',
      key: 'dates',
      width: 240,
      fixed: 'left',
      render: (_, row) => (
        <Tooltip title={rowErrors[row.key]?.startDate ?? rowErrors[row.key]?.endDate}>
          <DatePicker.RangePicker
            size="small"
            format="DD/MM/YY"
            allowClear={false}
            value={row.draft.startDate ? [dayjs(row.draft.startDate), dayjs(row.draft.endDate || row.draft.startDate)] : null}
            status={status(row, 'startDate') ?? status(row, 'endDate')}
            disabled={!editable(row)}
            onCalendarChange={(dates, _strings, info) => {
              // Picking the departure fills the return date from the tour length.
              const start = dates?.[0];
              if (info.range === 'start' && start && !dates?.[1]) {
                setDraft(row.key, { startDate: start.format(DATE), endDate: start.add(Math.max(days - 1, 0), 'day').format(DATE) }, row.draft);
              }
            }}
            onChange={(dates) => {
              if (dates?.[0] && dates[1]) setDraft(row.key, { startDate: dates[0].format(DATE), endDate: dates[1].format(DATE) }, row.draft);
            }}
          />
        </Tooltip>
      ),
    },
    priceColumn('price', 'ราคา (ผู้ใหญ่)', 120),
    priceColumn('priceBeforeDiscount', 'ราคาก่อนลด'),
    {
      title: 'ที่นั่ง ว่าง / ทั้งหมด',
      key: 'seats',
      width: 190,
      render: (_, row) => (
        <Tooltip title={rowErrors[row.key]?.seatsAvailable}>
          <div className="flex items-center gap-1.5">
            <InputNumber<number>
              size="small"
              min={0}
              max={999}
              precision={0}
              value={row.draft.seatsAvailable}
              status={status(row, 'seatsAvailable')}
              disabled={!editable(row)}
              onChange={(v) => setDraft(row.key, { seatsAvailable: v ?? 0 }, row.draft)}
              className="w-[64px]"
              aria-label="ที่นั่งว่าง"
            />
            <span className="text-ink-muted">/</span>
            <InputNumber<number>
              size="small"
              min={0}
              max={999}
              precision={0}
              value={row.draft.seatsTotal}
              disabled={!editable(row)}
              onChange={(v) => setDraft(row.key, { seatsTotal: v ?? 0 }, row.draft)}
              className="w-[64px]"
              aria-label="ที่นั่งทั้งหมด"
            />
            {row.period && row.draft.seatsAvailable === 0 && (
              <Tag color="red" className="m-0">
                เต็ม
              </Tag>
            )}
          </div>
        </Tooltip>
      ),
    },
    priceColumn('childWithBedPrice', 'เด็ก (มีเตียง)'),
    priceColumn('childNoBedPrice', 'เด็ก (ไม่มีเตียง)'),
    priceColumn('singleSupplement', 'พักเดี่ยวเพิ่ม'),
    priceColumn('deposit', 'มัดจำ'),
    {
      title: 'หมายเหตุ (ภายใน)',
      key: 'note',
      width: 200,
      render: (_, row) => (
        <Input size="small" value={row.draft.note} maxLength={500} disabled={!editable(row)} onChange={(e) => setDraft(row.key, { note: e.target.value }, row.draft)} />
      ),
    },
    {
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_, row) => {
        const dirty = !row.period || !sameFields(row.draft, toFields(row.period));
        const busy = saving.has(row.key);
        if (row.period && !row.period.status) {
          return canStatus ? (
            <Button size="small" icon={<UndoOutlined />} onClick={() => confirmStatus(row.period as TourProgramPeriod)}>
              กู้คืน
            </Button>
          ) : (
            <Tag className="m-0">ลบแล้ว</Tag>
          );
        }
        return (
          <div className="flex items-center gap-1">
            {editable(row) && (
              <Button size="small" type="primary" icon={<SaveOutlined />} disabled={!dirty} loading={busy} onClick={() => void save(row)}>
                บันทึก
              </Button>
            )}
            {dirty && !busy && (
              <Tooltip title={row.period ? 'ยกเลิกการแก้ไข' : 'นำแถวนี้ออก'}>
                <Button size="small" type="text" icon={<CloseOutlined />} onClick={() => dropDraft(row.key)} aria-label="ยกเลิก" />
              </Tooltip>
            )}
            {row.period && !dirty && canStatus && (
              <Tooltip title="ลบรอบนี้">
                <Button size="small" type="text" danger icon={<DeleteOutlined />} onClick={() => confirmStatus(row.period as TourProgramPeriod)} aria-label="ลบรอบ" />
              </Tooltip>
            )}
          </div>
        );
      },
    },
  ];

  const active = (periods ?? []).filter((p) => p.status);
  const startingPrice = active.length ? Math.min(...active.map((p) => p.price)) : null;
  const unsaved = Object.keys(drafts).length;

  return (
    <Card
      title={
        <span className="flex flex-wrap items-baseline gap-x-3">
          <span>รอบเดินทางและราคา</span>
          <span className="text-sm font-normal text-ink-muted">
            {active.length} รอบ{startingPrice !== null && ` · ราคาเริ่มต้น ${formatBaht(startingPrice)}`}
          </span>
        </span>
      }
      extra={
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1.5 text-sm text-ink-muted">
            <Switch size="small" checked={showDeleted} onChange={setShowDeleted} /> แสดงรอบที่ลบแล้ว
          </span>
          {canCreate && (
            <Button icon={<PlusOutlined />} onClick={addRow}>
              เพิ่มรอบ
            </Button>
          )}
        </div>
      }
    >
      {loadError !== null && <Alert type="error" showIcon title={errorMessage(loadError)} className="mb-4" />}
      {unsaved > 0 && <Alert type="warning" showIcon title={`มี ${unsaved} แถวที่ยังไม่บันทึก — กด "บันทึก" ทีละแถว`} className="mb-4" />}
      <Table<Row>
        rowKey="key"
        size="small"
        columns={columns}
        dataSource={rows}
        loading={periods === null && loadError === null}
        pagination={false}
        scroll={{ x: 1600 }}
        rowClassName={(row) => (row.period && !row.period.status ? 'opacity-50' : '')}
        locale={{ emptyText: canCreate ? 'ยังไม่มีรอบเดินทาง — กด "เพิ่มรอบ"' : 'ยังไม่มีรอบเดินทาง' }}
      />
    </Card>
  );
}
