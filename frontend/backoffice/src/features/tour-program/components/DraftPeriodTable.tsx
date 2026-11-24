import { DeleteOutlined, PlusOutlined, WarningOutlined } from '@ant-design/icons';
import { Button, Card, DatePicker, Input, InputNumber, type InputNumberProps, Space, Table, type TableProps, Tag, Tooltip } from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';
import type { AIDraftPeriod } from '../../../types/aiExtract.ts';
import { type DraftPeriodRow, newDraftRowKey } from '../lib/draftPeriods.ts';
import { seatsForTotal } from '../lib/seats.ts';

const DATE = 'YYYY-MM-DD';

const money: InputNumberProps<number> = {
  min: 0,
  max: 10_000_000,
  precision: 0,
  controls: false,
  formatter: (v) => `${v ?? ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ','),
  parser: (v) => Number((v ?? '').replace(/,/g, '')),
};

type PriceKey = 'price' | 'childWithBedPrice' | 'childNoBedPrice' | 'infantPrice' | 'singleSupplement' | 'deposit';

type Props = {
  rows: DraftPeriodRow[];
  onChange: (rows: DraftPeriodRow[]) => void;
  days: number | null;
  attention: boolean; // low AI confidence in dates or prices
  disabled?: boolean;
};

/**
 * Periods read from the PDF, edited before the tour exists. They are created with
 * POST /tour-program-periods right after the tour is saved. Seats never come from the PDF.
 */
export default function DraftPeriodTable({ rows, onChange, days, attention, disabled = false }: Props) {
  const [bulkSeats, setBulkSeats] = useState<number | null>(null);

  const set = (key: string, patch: Partial<AIDraftPeriod>) =>
    onChange(rows.map((r) => (r.key === key ? { ...r, value: { ...r.value, ...patch } } : r)));

  const add = () => {
    const last = rows.at(-1)?.value;
    onChange([
      ...rows,
      {
        key: newDraftRowKey(),
        warnings: [],
        value: {
          startDate: '',
          endDate: '',
          price: last?.price ?? null,
          priceBeforeDiscount: null,
          childWithBedPrice: last?.childWithBedPrice ?? null,
          childNoBedPrice: last?.childNoBedPrice ?? null,
          infantPrice: last?.infantPrice ?? null,
          singleSupplement: last?.singleSupplement ?? null,
          deposit: last?.deposit ?? null,
          seatsTotal: last?.seatsTotal ?? 0,
          seatsAvailable: last?.seatsAvailable ?? 0,
          note: '',
        },
      },
    ]);
  };

  const fillSeats = () => {
    if (bulkSeats === null) return;
    onChange(rows.map((r) => ({ ...r, value: { ...r.value, seatsTotal: bulkSeats, seatsAvailable: bulkSeats } })));
  };

  const price = (field: PriceKey, title: string, required = false, width = 116): NonNullable<TableProps<DraftPeriodRow>['columns']>[number] => ({
    title,
    key: field,
    width,
    render: (_, row) => (
      <InputNumber<number>
        {...money}
        size="small"
        value={row.value[field]}
        status={required && row.value[field] === null ? 'error' : undefined}
        disabled={disabled}
        placeholder="—"
        onChange={(v) => set(row.key, { [field]: v ?? null })}
      />
    ),
  });

  const columns: TableProps<DraftPeriodRow>['columns'] = [
    {
      title: 'วันเดินทาง',
      key: 'dates',
      width: 260,
      fixed: 'left',
      render: (_, row) => (
        <div className="flex items-center gap-1.5">
          <DatePicker.RangePicker
            size="small"
            format="DD/MM/YY"
            allowClear={false}
            disabled={disabled}
            status={!row.value.startDate ? 'error' : undefined}
            value={row.value.startDate ? [dayjs(row.value.startDate), dayjs(row.value.endDate || row.value.startDate)] : null}
            onCalendarChange={(dates, _s, info) => {
              const start = dates?.[0];
              if (info.range === 'start' && start && !dates?.[1]) {
                set(row.key, { startDate: start.format(DATE), endDate: start.add(Math.max((days ?? 1) - 1, 0), 'day').format(DATE) });
              }
            }}
            onChange={(dates) => {
              if (dates?.[0] && dates[1]) set(row.key, { startDate: dates[0].format(DATE), endDate: dates[1].format(DATE) });
            }}
          />
          {row.warnings.length > 0 && (
            <Tooltip title={row.warnings.join('\n')}>
              <WarningOutlined className="text-amber-500" aria-label="AI แจ้งเตือน" />
            </Tooltip>
          )}
        </div>
      ),
    },
    price('price', 'ราคา (ผู้ใหญ่)', true, 120),
    {
      title: 'ที่นั่ง ว่าง / ทั้งหมด',
      key: 'seats',
      width: 170,
      render: (_, row) => (
        <div className="flex items-center gap-1.5">
          <InputNumber<number>
            size="small"
            min={0}
            max={999}
            precision={0}
            value={row.value.seatsAvailable}
            status={row.value.seatsAvailable === null ? 'error' : undefined}
            disabled={disabled}
            onChange={(v) => set(row.key, { seatsAvailable: v ?? null })}
            className="w-[60px]"
            aria-label="ที่นั่งว่าง"
          />
          <span className="text-ink-muted">/</span>
          <InputNumber<number>
            size="small"
            min={0}
            max={999}
            precision={0}
            value={row.value.seatsTotal}
            status={row.value.seatsTotal === null ? 'error' : undefined}
            disabled={disabled}
            onChange={(v) => set(row.key, v === null ? { seatsTotal: null } : seatsForTotal(row.value.seatsTotal, row.value.seatsAvailable, v))}
            className="w-[60px]"
            aria-label="ที่นั่งทั้งหมด"
          />
        </div>
      ),
    },
    price('childWithBedPrice', 'เด็ก (มีเตียง)'),
    price('childNoBedPrice', 'เด็ก (ไม่มีเตียง)'),
    price('infantPrice', 'ทารก (Infant)'),
    price('singleSupplement', 'พักเดี่ยวเพิ่ม'),
    price('deposit', 'มัดจำ'),
    {
      title: 'หมายเหตุ (ภายใน)',
      key: 'note',
      width: 200,
      render: (_, row) => <Input size="small" value={row.value.note} maxLength={500} disabled={disabled} onChange={(e) => set(row.key, { note: e.target.value })} />,
    },
    {
      key: 'remove',
      width: 56,
      fixed: 'right',
      render: (_, row) => (
        <Button size="small" type="text" danger icon={<DeleteOutlined />} disabled={disabled} aria-label="ลบรอบ" onClick={() => onChange(rows.filter((r) => r.key !== row.key))} />
      ),
    },
  ];

  return (
    <Card
      title={
        <span className="flex items-center gap-2">
          รอบเดินทางจาก PDF ({rows.length} รอบ)
          <Tag color="purple" className="m-0">
            AI
          </Tag>
          {attention && (
            <Tag color="orange" className="m-0">
              ตรวจสอบวันที่และราคา
            </Tag>
          )}
        </span>
      }
      extra={
        <Space wrap>
          <InputNumber<number> size="small" min={0} max={999} precision={0} value={bulkSeats} onChange={setBulkSeats} placeholder="ที่นั่ง" className="w-[80px]" disabled={disabled} />
          <Button size="small" onClick={fillSeats} disabled={disabled || bulkSeats === null}>
            ใส่ที่นั่งทุกรอบ
          </Button>
        </Space>
      }
    >
      <p className="mt-0 text-[13px] text-ink-muted">
        บันทึกพร้อมโปรแกรม — ตรวจทุกรอบกับไฟล์ต้นฉบับ ที่นั่งเริ่มที่ 0 (ไม่ระบุ — หน้าเว็บไม่แสดงว่าเต็ม) แก้ได้ภายหลัง ลบรอบที่ไม่ต้องการได้
      </p>
      <Table<DraftPeriodRow>
        size="small"
        rowKey="key"
        columns={columns}
        dataSource={rows}
        pagination={false}
        scroll={{ x: 1370 }}
        rowClassName={(row) => (row.warnings.length > 0 ? 'bg-amber-50' : '')}
      />
      <Button className="mt-3" icon={<PlusOutlined />} onClick={add} disabled={disabled}>
        เพิ่มรอบ
      </Button>
    </Card>
  );
}
