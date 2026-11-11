import { CopyOutlined, DeleteOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, DatePicker, Form, Input, InputNumber, Radio, Select, Switch, TimePicker } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react';
import { bookingSourcesApi } from '../../../api/bookingSources.ts';
import { periodsApi, tourProgramsApi } from '../../../api/tourPrograms.ts';
import { usersApi } from '../../../api/users.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatBaht, formatPhone } from '../../../lib/format.ts';
import type { BookingSource } from '../../../types/bookingSource.ts';
import type { TourProgram, TourProgramListItem, TourProgramPeriod } from '../../../types/tourProgram.ts';
import type { SalesUser } from '../../../types/user.ts';
import {
  buildSummaryText,
  computeTotals,
  defaultIsDeposit,
  defaultRemainingDueDate,
  type PaymentMethod,
  refNameLabel,
  type SummaryInput,
  thaiDate,
} from './summary.ts';
import TourPicker from './TourPicker.tsx';

const DEFAULT_CREDIT_FEE_PERCENT = 2.6;
const DATE_FORMAT = 'DD/MM/YYYY';

type Num = number | null;

type FormValues = {
  sourceId?: string;
  refName: string;
  refBookingNo: string;
  periodId?: string;
  adultCount: Num;
  adultPrice: Num;
  childCount: Num;
  childPrice: Num;
  infantCount: Num;
  infantPrice: Num;
  visaCount: Num;
  visaPrice: Num;
  foreignCount: Num;
  foreignPrice: Num;
  singleRoomCount: Num;
  singleRoomPrice: Num;
  joinLandCount: Num;
  joinLandPrice: Num;
  fuelTaxCount: Num; // empty = number of travellers
  fuelTaxPrice: Num;
  airportTaxCount: Num; // empty = number of travellers
  airportTaxPrice: Num;
  options: { name?: string; qty?: Num; price?: Num }[];
  discountPerPerson: Num;
  paymentMethod: PaymentMethod;
  creditFeePercent: Num;
  isDeposit: boolean;
  depositPerPerson: Num;
  depositDueDate: Dayjs | null;
  depositDueTime: Dayjs | null;
  remainingDueDate: Dayjs | null;
  fullDueDate: Dayjs | null;
  fullDueTime: Dayjs | null;
  customerName: string;
  customerPhone: string;
  customerAddress: string;
  salesUserId?: string;
};

const INITIAL: FormValues = {
  refName: '',
  refBookingNo: '',
  adultCount: 1,
  adultPrice: null,
  childCount: null,
  childPrice: null,
  infantCount: null,
  infantPrice: null,
  visaCount: null,
  visaPrice: null,
  foreignCount: null,
  foreignPrice: null,
  singleRoomCount: null,
  singleRoomPrice: null,
  joinLandCount: null,
  joinLandPrice: null,
  fuelTaxCount: null,
  fuelTaxPrice: null,
  airportTaxCount: null,
  airportTaxPrice: null,
  options: [],
  discountPerPerson: null,
  paymentMethod: 'transfer',
  creditFeePercent: DEFAULT_CREDIT_FEE_PERCENT,
  isDeposit: false,
  depositPerPerson: null,
  depositDueDate: null,
  depositDueTime: null,
  remainingDueDate: null,
  fullDueDate: null,
  fullDueTime: null,
  customerName: '',
  customerPhone: '',
  customerAddress: '',
};

const num = (v: Num | undefined) => v ?? 0;
const date = (d: Dayjs | null | undefined) => (d ? d.format('YYYY-MM-DD') : null);
const time = (t: Dayjs | null | undefined) => (t ? t.format('HH:mm') : null);

function periodLabel(p: TourProgramPeriod) {
  const range = `${dayjs(p.startDate).format('DD/MM/YY')} – ${dayjs(p.endDate).format('DD/MM/YY')}`;
  const seats = p.seatsAvailable > 0 ? `ว่าง ${p.seatsAvailable}` : 'เต็ม';
  return `${range} · ${formatBaht(p.price)} · ${seats}`;
}

// Active sources and sales users for the dropdowns, loaded once. onSales gets the caller's id
// (or '') so the form can default the sales person.
function useLookups(onSales: (defaultId: string) => void) {
  const { message } = App.useApp();
  const [sources, setSources] = useState<BookingSource[]>([]);
  const [sales, setSales] = useState<{ items: SalesUser[]; defaultId: string } | null>(null);
  const onSalesRef = useRef(onSales);
  useEffect(() => {
    onSalesRef.current = onSales;
  });

  useEffect(() => {
    let cancelled = false;
    bookingSourcesApi
      .list({ status: true, page: 1, pageSize: 500 })
      .then((res) => !cancelled && setSources(res.items))
      .catch((err: unknown) => !cancelled && message.error(errorMessage(err)));
    usersApi
      .sales()
      .then((res) => {
        if (cancelled) return;
        setSales(res);
        onSalesRef.current(res.defaultId);
      })
      .catch((err: unknown) => !cancelled && message.error(errorMessage(err)));
    return () => {
      cancelled = true;
    };
  }, [message]);

  return { sources, sales };
}

export default function BookingSummaryTool() {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [values, setValues] = useState<FormValues>(INITIAL);
  const { sources, sales } = useLookups((defaultId) => {
    if (defaultId && !form.getFieldValue('salesUserId')) {
      form.setFieldsValue({ salesUserId: defaultId });
      setValues(form.getFieldsValue(true) as FormValues);
    }
  });

  const [tourItem, setTourItem] = useState<TourProgramListItem | null>(null);
  const [tour, setTour] = useState<TourProgram | null>(null);
  const [periods, setPeriods] = useState<TourProgramPeriod[]>([]);
  const [tourLoading, setTourLoading] = useState(false);

  // setFieldsValue doesn't fire onValuesChange, so sync the preview by hand after it.
  const patch = (next: Partial<FormValues>) => {
    form.setFieldsValue(next);
    setValues(form.getFieldsValue(true) as FormValues);
  };

  const selectTour = async (item: TourProgramListItem | null) => {
    setTourItem(item);
    setTour(null);
    setPeriods([]);
    patch({ periodId: undefined });
    if (!item) return;
    setTourLoading(true);
    try {
      const [full, res] = await Promise.all([tourProgramsApi.get(item.id), periodsApi.list({ tourProgramId: item.id, status: true })]);
      const today = dayjs().format('YYYY-MM-DD');
      setTour(full);
      setPeriods(res.items.filter((p) => p.endDate >= today).sort((a, b) => a.startDate.localeCompare(b.startDate)));
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setTourLoading(false);
    }
  };

  // Picking a departure fills its prices, deposit and due dates; staff can still edit them.
  const selectPeriod = (periodId: string | undefined) => {
    const p = periods.find((x) => x.id === periodId);
    if (!p) return;
    const deposit = defaultIsDeposit(p.startDate);
    patch({
      periodId,
      adultPrice: p.price,
      childPrice: p.childWithBedPrice ?? p.price,
      singleRoomPrice: p.singleSupplement,
      depositPerPerson: p.deposit,
      isDeposit: deposit && p.deposit !== null,
      remainingDueDate: dayjs(defaultRemainingDueDate(p.startDate)),
    });
  };

  const reset = () => {
    form.resetFields();
    form.setFieldsValue({ salesUserId: sales?.defaultId || undefined });
    setValues(form.getFieldsValue(true) as FormValues);
    setTourItem(null);
    setTour(null);
    setPeriods([]);
  };

  const period = periods.find((p) => p.id === values.periodId) ?? null;
  const source = sources.find((s) => s.id === values.sourceId) ?? null;
  const salesUser = sales?.items.find((u) => u.id === values.salesUserId) ?? null;

  const input: SummaryInput = useMemo(
    () => ({
      source: source && { name: source.name, shortName: source.shortName },
      refName: values.refName ?? '',
      refBookingNo: values.refBookingNo ?? '',
      tour: tour && {
        name: tour.name,
        days: tour.days,
        nights: tour.nights,
        pdfFile: tour.pdfFile,
        airlineName: tour.airline?.name ?? null,
        wholesaleName: tour.wholesale.name,
      },
      period: period && { startDate: period.startDate, endDate: period.endDate },
      adult: { count: num(values.adultCount), price: num(values.adultPrice) },
      adultPriceBeforeDiscount: period?.priceBeforeDiscount ?? null,
      child: { count: num(values.childCount), price: num(values.childPrice) },
      infant: { count: num(values.infantCount), price: num(values.infantPrice) },
      visa: { count: num(values.visaCount), price: num(values.visaPrice) },
      foreignPassport: { count: num(values.foreignCount), price: num(values.foreignPrice) },
      singleRoom: { count: num(values.singleRoomCount), price: num(values.singleRoomPrice) },
      joinLand: { count: num(values.joinLandCount), price: num(values.joinLandPrice) },
      fuelTax: { count: values.fuelTaxCount ?? null, price: num(values.fuelTaxPrice) },
      airportTax: { count: values.airportTaxCount ?? null, price: num(values.airportTaxPrice) },
      options: (values.options ?? []).map((o) => ({ name: o?.name ?? '', qty: num(o?.qty), price: num(o?.price) })),
      discountPerPerson: num(values.discountPerPerson),
      paymentMethod: values.paymentMethod,
      creditFeePercent: num(values.creditFeePercent),
      isDeposit: values.isDeposit,
      depositPerPerson: num(values.depositPerPerson),
      depositDueDate: date(values.depositDueDate),
      depositDueTime: time(values.depositDueTime),
      remainingDueDate: date(values.remainingDueDate),
      fullDueDate: date(values.fullDueDate),
      fullDueTime: time(values.fullDueTime),
      customerName: values.customerName ?? '',
      customerPhone: values.customerPhone ?? '',
      customerAddress: values.customerAddress ?? '',
      sales: salesUser && { name: salesUser.name, phone: salesUser.phone },
    }),
    [values, source, tour, period, salesUser],
  );
  const totals = useMemo(() => computeTotals(input), [input]);
  const text = useMemo(() => buildSummaryText(input, totals), [input, totals]);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('คัดลอกข้อความแล้ว');
    } catch {
      message.error('คัดลอกไม่สำเร็จ ลองเลือกข้อความแล้วกด Ctrl+C');
    }
  };

  const seatsShort = period && totals.travellers > period.seatsAvailable;

  return (
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(340px,440px)]">
      <Form<FormValues>
        form={form}
        layout="vertical"
        requiredMark={false}
        initialValues={INITIAL}
        onValuesChange={(_, all) => setValues(all)}
        className="flex min-w-0 flex-col gap-4"
      >
        <Card title="ข้อมูลการจอง" size="small">
          <div className="grid gap-x-4 sm:grid-cols-3">
            <Form.Item label="จองผ่าน" name="sourceId">
              <Select
                allowClear
                placeholder="เลือกช่องทาง"
                options={sources.map((s) => ({ value: s.id, label: s.name }))}
                notFoundContent="ยังไม่มีช่องทาง — เพิ่มที่เมนู ช่องทางการจอง"
              />
            </Form.Item>
            <Form.Item label={refNameLabel(source)} name="refName">
              <Input autoComplete="off" />
            </Form.Item>
            <Form.Item label="เลข booking" name="refBookingNo">
              <Input autoComplete="off" />
            </Form.Item>
          </div>
        </Card>

        <Card title="โปรแกรมทัวร์" size="small">
          <Form.Item label="โปรแกรม">
            <TourPicker value={tourItem} onChange={(t) => void selectTour(t)} />
          </Form.Item>
          <Form.Item label="วันเดินทาง" name="periodId" className="mb-2">
            <Select
              placeholder={tourItem ? 'เลือกรอบเดินทาง' : 'เลือกโปรแกรมก่อน'}
              disabled={!tourItem}
              loading={tourLoading}
              onChange={selectPeriod}
              options={periods.map((p) => ({ value: p.id, label: periodLabel(p) }))}
              notFoundContent="ไม่มีรอบเดินทางที่เปิดอยู่"
            />
          </Form.Item>
          {tour && (
            <p className="m-0 text-[13px] text-ink-muted">
              {tour.days} วัน {tour.nights} คืน · {tour.wholesale.name}
              {tour.airline && ` · ${tour.airline.name}`}
              {!tour.pdfFile && ' · ไม่มีไฟล์ PDF โปรแกรม'}
            </p>
          )}
        </Card>

        <Card title="ผู้เดินทางและราคา" size="small">
          <div className="grid grid-cols-[minmax(0,1fr)_96px_140px] items-center gap-x-3 gap-y-2 max-sm:grid-cols-[minmax(0,1fr)_80px_112px]">
            <span className="text-[13px] text-ink-muted" />
            <span className="text-[13px] text-ink-muted">จำนวน</span>
            <span className="text-[13px] text-ink-muted">ราคา/หน่วย (บาท)</span>
            <PriceRow label="ผู้ใหญ่" count="adultCount" price="adultPrice" />
            <PriceRow label="เด็กอายุต่ำกว่า 20 ปี" count="childCount" price="childPrice" />
            <PriceRow label="ทารก" count="infantCount" price="infantPrice" />
            <PriceRow label="วีซ่า" count="visaCount" price="visaPrice" />
            <PriceRow label="พาสปอร์ตต่างชาติ" count="foreignCount" price="foreignPrice" />
            <PriceRow label="ห้องพักเดี่ยว (ห้อง)" count="singleRoomCount" price="singleRoomPrice" />
            <PriceRow label="Join land" count="joinLandCount" price="joinLandPrice" />
            <PriceRow label="ภาษีน้ำมัน" count="fuelTaxCount" price="fuelTaxPrice" countPlaceholder={String(totals.travellers)} />
            <PriceRow label="ภาษีสนามบิน" count="airportTaxCount" price="airportTaxPrice" countPlaceholder={String(totals.travellers)} />
          </div>

          <Form.List name="options">
            {(fields, { add, remove }) => (
              <div className="mt-3 flex flex-col gap-2">
                {fields.map((field) => (
                  <div key={field.key} className="grid grid-cols-[minmax(0,1fr)_96px_140px_32px] items-center gap-x-3 max-sm:grid-cols-[minmax(0,1fr)_80px_112px_32px]">
                    <Form.Item name={[field.name, 'name']} noStyle>
                      <Input placeholder="ชื่อรายการเพิ่มเติม" autoComplete="off" />
                    </Form.Item>
                    <Form.Item name={[field.name, 'qty']} noStyle>
                      <InputNumber min={0} precision={0} className="w-full" placeholder="0" />
                    </Form.Item>
                    <Form.Item name={[field.name, 'price']} noStyle>
                      <InputNumber min={0} precision={0} className="w-full" placeholder="0" />
                    </Form.Item>
                    <Button type="text" icon={<DeleteOutlined />} aria-label="ลบรายการ" onClick={() => remove(field.name)} />
                  </div>
                ))}
                <Button type="dashed" icon={<PlusOutlined />} onClick={() => add({ name: '', qty: totals.travellers || 1, price: null })} className="self-start">
                  เพิ่มรายการ
                </Button>
              </div>
            )}
          </Form.List>

          <div className="mt-4 flex flex-wrap items-end gap-4">
            <Form.Item label="ส่วนลด/ท่าน (บาท)" name="discountPerPerson" className="mb-0 w-[180px]" extra="หักจากราคาผู้ใหญ่ เด็ก และทารก">
              <InputNumber min={0} precision={0} className="w-full" placeholder="0" />
            </Form.Item>
          </div>

          <p className="mt-4 mb-0 text-sm">
            ผู้เดินทางทั้งหมด <span className="font-semibold">{totals.travellers}</span> ท่าน
            {period && <span className="text-ink-muted"> · ที่นั่งว่าง {period.seatsAvailable}</span>}
          </p>
          {seatsShort && <Alert type="warning" showIcon className="mt-2" title="จำนวนผู้เดินทางมากกว่าที่นั่งว่างของรอบนี้" />}
        </Card>

        <Card title="การชำระเงิน" size="small">
          <div className="flex flex-wrap items-end gap-4">
            <Form.Item label="ประเภทการจ่าย" name="paymentMethod">
              <Radio.Group
                optionType="button"
                options={[
                  { value: 'transfer', label: 'โอน' },
                  { value: 'credit', label: 'รูดบัตร' },
                ]}
              />
            </Form.Item>
            {values.paymentMethod === 'credit' && (
              <Form.Item label="ค่าธรรมเนียม (%)" name="creditFeePercent" className="w-[140px]">
                <InputNumber min={0} max={20} step={0.1} precision={2} className="w-full" />
              </Form.Item>
            )}
            <Form.Item label="มัดจำ" name="isDeposit" valuePropName="checked">
              <Switch />
            </Form.Item>
          </div>

          {values.isDeposit ? (
            <div className="grid gap-x-4 sm:grid-cols-2">
              <Form.Item label="มัดจำ/ท่าน (บาท)" name="depositPerPerson">
                <InputNumber min={0} precision={0} className="w-full" />
              </Form.Item>
              <div className="grid grid-cols-[minmax(0,1fr)_110px] gap-x-2">
                <Form.Item label="ชำระมัดจำภายในวันที่" name="depositDueDate">
                  <DatePicker format={DATE_FORMAT} className="w-full" />
                </Form.Item>
                <Form.Item label="ก่อนเวลา" name="depositDueTime">
                  <TimePicker format="HH:mm" minuteStep={5} needConfirm={false} className="w-full" />
                </Form.Item>
              </div>
              <Form.Item label="ชำระส่วนที่เหลือภายในวันที่" name="remainingDueDate" extra={period ? 'ค่าเริ่มต้น 30 วันก่อนเดินทาง' : undefined}>
                <DatePicker format={DATE_FORMAT} className="w-full" />
              </Form.Item>
            </div>
          ) : (
            <div className="grid grid-cols-[minmax(0,1fr)_110px] gap-x-2 sm:max-w-[50%]">
              <Form.Item label="ชำระเต็มจำนวนภายในวันที่" name="fullDueDate">
                <DatePicker format={DATE_FORMAT} className="w-full" />
              </Form.Item>
              <Form.Item label="ก่อนเวลา" name="fullDueTime">
                <TimePicker format="HH:mm" minuteStep={5} needConfirm={false} className="w-full" />
              </Form.Item>
            </div>
          )}
        </Card>

        <Card title="ข้อมูลลูกค้า" size="small">
          <div className="grid gap-x-4 sm:grid-cols-2">
            <Form.Item label="ชื่อลูกค้า" name="customerName" extra='ข้อความจะขึ้นต้นด้วย "คุณ" ให้อัตโนมัติ'>
              <Input autoComplete="off" />
            </Form.Item>
            <Form.Item label="เบอร์ลูกค้า" name="customerPhone">
              <Input type="tel" autoComplete="off" />
            </Form.Item>
          </div>
          <Form.Item label="ที่อยู่ลูกค้า" name="customerAddress">
            <Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} />
          </Form.Item>
          <Form.Item label="ผู้ดูแล" name="salesUserId" className="mb-0">
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder="เลือกพนักงานขาย"
              loading={sales === null}
              options={(sales?.items ?? []).map((u) => ({ value: u.id, label: u.phone ? `${u.name} ${formatPhone(u.phone)}` : u.name }))}
            />
          </Form.Item>
          {salesUser && !salesUser.phone && <p className="mt-1 mb-0 text-[13px] text-ink-muted">ผู้ดูแลยังไม่มีเบอร์โทร — เพิ่มได้ที่หน้าผู้ใช้</p>}
        </Card>
      </Form>

      <div className="flex flex-col gap-3 lg:sticky lg:top-6">
        <Card
          title="ข้อความสำหรับส่งลูกค้า"
          size="small"
          extra={
            <Button type="text" size="small" icon={<ReloadOutlined />} onClick={reset}>
              ล้างฟอร์ม
            </Button>
          }
        >
          <dl className="m-0 mb-3 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
            <Stat label="รวมเป็นเงิน" value={formatBaht(totals.grandTotal)} strong />
            {totals.fee > 0 && <Stat label="ค่าธรรมเนียมบัตร" value={formatBaht(totals.fee)} />}
            {values.isDeposit && <Stat label="มัดจำ" value={`${formatBaht(totals.depositToPay)} ภายใน ${thaiDate(input.depositDueDate)}`} />}
            {values.isDeposit && <Stat label="ส่วนที่เหลือ" value={`${formatBaht(totals.remaining)} ภายใน ${thaiDate(input.remainingDueDate)}`} />}
          </dl>
          <pre className="m-0 max-h-[60vh] overflow-auto rounded-lg border border-line bg-sky-soft/40 p-3 font-sans text-[13px] leading-relaxed whitespace-pre-wrap">
            {text}
          </pre>
          <Button type="primary" block icon={<CopyOutlined />} onClick={() => void copy()} className="mt-3">
            คัดลอกข้อความ
          </Button>
        </Card>
      </div>
    </div>
  );
}

function Stat({ label, value, strong }: { label: string; value: ReactNode; strong?: boolean }) {
  return (
    <>
      <dt className="text-ink-muted">{label}</dt>
      <dd className={`m-0 text-right ${strong ? 'text-base font-semibold' : ''}`}>{value}</dd>
    </>
  );
}

type NumField = { [K in keyof FormValues]: FormValues[K] extends Num ? K : never }[keyof FormValues];

function PriceRow({ label, count, price, countPlaceholder = '0' }: { label: string; count: NumField; price: NumField; countPlaceholder?: string }) {
  return (
    <>
      <span className="text-sm">{label}</span>
      <Form.Item name={count} noStyle>
        <InputNumber min={0} precision={0} className="w-full" placeholder={countPlaceholder} aria-label={`จำนวน${label}`} />
      </Form.Item>
      <Form.Item name={price} noStyle>
        <InputNumber min={0} precision={0} className="w-full" placeholder="0" aria-label={`ราคา${label}`} />
      </Form.Item>
    </>
  );
}
