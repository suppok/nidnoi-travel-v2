import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, DatePicker, Form, Input, InputNumber, Radio, Select, Switch, Tag, TimePicker } from 'antd';
import dayjs from 'dayjs';
import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { periodsApi, tourProgramsApi } from '../../../api/tourPrograms.ts';
import { TourProgramStates } from '../../../constants/tourProgramStates.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import { formatBaht, formatPhone } from '../../../lib/format.ts';
import type { BookingSource } from '../../../types/bookingSource.ts';
import type { LineBooking, LineBookingContact, LineBookingWholesaleEmailResult } from '../../../types/lineBooking.ts';
import type { TourProgramListItem, TourProgramPeriod } from '../../../types/tourProgram.ts';
import type { SalesUserListResponse } from '../../../types/user.ts';
import LineContactPicker from '../../line-booking/components/LineContactPicker.tsx';
import OpenChatLink from '../../line-booking/components/OpenChatLink.tsx';
import { refNameLabel } from '../../sales-tool/booking-summary/summary.ts';
import TourPicker from '../../sales-tool/booking-summary/TourPicker.tsx';
import { type DealValues, dealError, dealFromBooking, dealFromSummary, dealRequest, dealTotals, EMPTY_DEAL, financeWarning, periodPatch } from '../lib/deal.ts';
import { shortThaiDate } from '../lib/format.ts';
import type { ParsedSummary } from '../lib/parseSummary.ts';
import ReopenedSummary from './ReopenedSummary.tsx';
import SlipUploadField from './SlipUploadField.tsx';
import SummaryPasteBox from './SummaryPasteBox.tsx';

const DATE_FORMAT = 'DD/MM/YYYY';

type FormValues = DealValues & {
  salesUserId?: string;
  bookingSourceId?: string;
  refName: string;
  refBookingNo: string;
  periodId?: string;
  customerName: string;
  customerPhone: string;
  customerAddress: string;
  chatUrl: string;
  slipImages: string[];
  note: string;
};

// Loaded by the page before the form renders.
export type EntryLookups = { sales: SalesUserListResponse; sources: BookingSource[] };

export type SavedResult = { bookingNo: string; totalPrice: number; message?: string; wholesaleEmail?: LineBookingWholesaleEmailResult };

type Props = {
  // A reopened booking to edit; null = a new booking.
  booking: LineBooking | null;
  // A summary pasted on the first screen, applied when the form opens.
  initialSummary: ParsedSummary | null;
  lookups: EntryLookups;
  onSaved: (result: SavedResult) => void;
  onCancel: () => void;
  onChangePurpose?: () => void;
};

type PasteReport = { filled: string[]; warnings: string[] };

function initialValues(booking: LineBooking | null, defaultSalesId: string | undefined): FormValues {
  if (!booking) {
    return {
      ...EMPTY_DEAL,
      salesUserId: defaultSalesId || undefined,
      refName: '',
      refBookingNo: '',
      customerName: '',
      customerPhone: '',
      customerAddress: '',
      chatUrl: '',
      slipImages: [],
      note: '',
    };
  }
  return {
    ...dealFromBooking(booking),
    bookingSourceId: booking.bookingSource?.id,
    refName: booking.refName,
    refBookingNo: booking.refBookingNo,
    customerName: booking.customerName,
    customerPhone: booking.customerPhone,
    customerAddress: booking.customerAddress,
    chatUrl: booking.lineContact?.chatUrl ?? '',
    slipImages: [],
    note: booking.note,
  };
}

function periodLabel(p: TourProgramPeriod) {
  const range = `${dayjs(p.startDate).format('DD/MM/YY')} – ${dayjs(p.endDate).format('DD/MM/YY')}`;
  return `${range} · ${formatBaht(p.price)} · ว่าง ${p.seatsAvailable}`;
}

// Departures that can still be booked: active (asked for), after today and with seats left.
async function bookablePeriods(tourId: string): Promise<TourProgramPeriod[]> {
  const res = await periodsApi.list({ tourProgramId: tourId, status: true });
  const today = dayjs().format('YYYY-MM-DD');
  return res.items.filter((p) => p.startDate > today && p.seatsAvailable > 0).sort((a, b) => a.startDate.localeCompare(b.startDate));
}

// The booking form: a new booking (with its slips), or the deal of a reopened one (tour, period,
// sales owner, slips and payment plan stay as they are).
export default function EntryForm({ booking, initialSummary, lookups, onSaved, onCancel, onChangePurpose }: Props) {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [initial] = useState(() => initialValues(booking, lookups.sales.defaultId));
  const [values, setValues] = useState<FormValues>(initial);
  const [tourItem, setTourItem] = useState<TourProgramListItem | null>(null);
  const [periods, setPeriods] = useState<TourProgramPeriod[]>([]);
  const [tourLoading, setTourLoading] = useState(false);
  const [lineContact, setLineContact] = useState<LineBookingContact | null>(booking?.lineContact ?? null);
  const [beforeDiscount, setBeforeDiscount] = useState<number | null>(booking?.adultPriceBeforeDiscount ?? null);
  const [pasteReport, setPasteReport] = useState<PasteReport | null>(null);
  const [pastedTotal, setPastedTotal] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const editing = booking !== null;
  const totals = useMemo(() => dealTotals(values), [values]);
  const period = periods.find((p) => p.id === values.periodId) ?? null;
  const sources = lookups.sources;
  const sourceOptions = booking?.bookingSource && !sources.some((s) => s.id === booking.bookingSource?.id) ? [...sources, booking.bookingSource] : sources;
  const source = sourceOptions.find((s) => s.id === values.bookingSourceId) ?? null;

  // A picked chat brings the OA Manager link saved on it (if any).
  const pickContact = (c: LineBookingContact | null) => {
    setLineContact(c);
    patch({ chatUrl: c?.chatUrl ?? '' });
  };

  // setFieldsValue doesn't fire onValuesChange, so sync the totals by hand after it.
  const patch = (next: Partial<FormValues>) => {
    form.setFieldsValue(next);
    setValues(form.getFieldsValue(true) as FormValues);
  };

  const selectTour = async (item: TourProgramListItem | null): Promise<TourProgramPeriod[]> => {
    setTourItem(item);
    setPeriods([]);
    patch({ periodId: undefined });
    if (!item) return [];
    setTourLoading(true);
    try {
      const open = await bookablePeriods(item.id);
      setPeriods(open);
      return open;
    } catch (err) {
      message.error(errorMessage(err));
      return [];
    } finally {
      setTourLoading(false);
    }
  };

  // Picking a departure fills its prices and payment defaults; everything stays editable.
  const selectPeriod = (periodId: string | undefined) => {
    const p = periods.find((x) => x.id === periodId);
    if (!p) return;
    setBeforeDiscount(p.priceBeforeDiscount);
    patch({ periodId, ...periodPatch(p) });
  };

  const applySummary = async (s: ParsedSummary) => {
    const filled: string[] = [];
    const warnings: string[] = [];
    const next: Partial<FormValues> = {};

    if (s.salesName && !editing) {
      const name = s.salesName;
      // The summary writes "<name> <phone>"; prefer the longest matching name.
      const match = [...lookups.sales.items]
        .sort((a, b) => b.name.length - a.name.length)
        .find((u) => name === u.name || name.startsWith(`${u.name} `));
      if (match) {
        next.salesUserId = match.id;
        filled.push('เซลล์');
      } else warnings.push(`ไม่พบเซลล์ "${name}" กรุณาเลือกเอง`);
    }
    if (s.source) {
      const match = sources.find((x) => x.name.toLowerCase() === s.source?.toLowerCase());
      if (match) {
        next.bookingSourceId = match.id;
        filled.push('ช่องทางการจอง');
      } else warnings.push(`ไม่พบช่องทาง "${s.source}" กรุณาเลือกเอง`);
    }
    if (s.refName) next.refName = s.refName;
    if (s.refBookingNo) next.refBookingNo = s.refBookingNo;
    if (s.customerName) {
      next.customerName = s.customerName;
      filled.push('ชื่อลูกค้า');
    }
    if (s.customerPhone) {
      next.customerPhone = s.customerPhone;
      filled.push('เบอร์โทร');
    }
    if (s.customerAddress) {
      next.customerAddress = s.customerAddress;
      filled.push('ที่อยู่');
    }
    if (s.travellers) {
      Object.assign(next, dealFromSummary(s));
      setBeforeDiscount(s.adultPriceBeforeDiscount);
      filled.push('ผู้เดินทาง ราคา และการชำระเงิน');
    }
    setPastedTotal(s.total);

    // The tour and its departure: picked without replacing the pasted prices.
    const keyword = s.tourCode ?? s.tourName;
    if (keyword && !editing) {
      try {
        const res = await tourProgramsApi.list({ keyword, status: true, state: TourProgramStates.Selling, page: 1, pageSize: 20 });
        const tour = s.tourCode ? res.items.find((t) => t.code.toLowerCase() === s.tourCode?.toLowerCase()) : res.items.find((t) => t.name === s.tourName);
        if (tour) {
          filled.push('โปรแกรมทัวร์');
          const open = await selectTour(tour);
          const match = open.find((p) => p.startDate === s.startDate && p.endDate === s.endDate);
          if (match) {
            next.periodId = match.id;
            filled.push('ช่วงเดินทาง');
          } else if (s.startDate && s.endDate) {
            warnings.push(`ไม่พบรอบ ${shortThaiDate(s.startDate)} – ${shortThaiDate(s.endDate)} ที่ยังว่าง กรุณาเลือกเอง`);
          } else warnings.push('ไม่พบวันเดินทางในข้อความ กรุณาเลือกรอบเดินทางเอง');
        } else warnings.push(`ไม่พบโปรแกรม "${keyword}" ที่เปิดขาย กรุณาเลือกเอง`);
      } catch {
        warnings.push('ค้นหาโปรแกรมทัวร์ไม่สำเร็จ กรุณาเลือกเอง');
      }
    }

    patch(next);
    setPasteReport({ filled, warnings });
  };

  // A summary pasted on the first screen: applied once, when the form opens.
  const applied = useRef(false);
  const applyRef = useRef(applySummary);
  useEffect(() => {
    applyRef.current = applySummary;
  });
  useEffect(() => {
    if (applied.current || !initialSummary) return;
    applied.current = true;
    void applyRef.current(initialSummary);
  }, [initialSummary]);

  const save = async (v: FormValues) => {
    const common = {
      lineContactId: lineContact?.id ?? null,
      chatUrl: lineContact ? (v.chatUrl ?? '').trim() : '',
      bookingSourceId: v.bookingSourceId ?? null,
      refName: v.refName.trim(),
      refBookingNo: v.refBookingNo.trim(),
      ...dealRequest(v, beforeDiscount),
      customerName: v.customerName.trim(),
      customerPhone: v.customerPhone.trim(),
      customerAddress: (v.customerAddress ?? '').trim(),
      note: (v.note ?? '').trim(),
    };
    setSubmitting(true);
    try {
      if (booking) {
        const res = await lineBookingsApi.updateDeal({ ...common, bookingNo: booking.bookingNo, balanceSlipImages: [] });
        onSaved({ bookingNo: res.bookingNo, totalPrice: res.totalPrice, message: 'บันทึกการแก้ไขใบจองแล้ว' });
      } else {
        if (!tourItem || !v.periodId || !v.salesUserId) return;
        const res = await lineBookingsApi.create({ ...common, salesUserId: v.salesUserId, tourCode: tourItem.code, periodId: v.periodId, slipImages: v.slipImages });
        onSaved(res);
      }
    } catch (err) {
      if (applyFieldErrors(form, err)) setError('ข้อมูลบางช่องไม่ถูกต้อง กรุณาตรวจสอบ');
      else setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  const submit = async () => {
    setError('');
    try {
      await form.validateFields();
    } catch {
      setError('กรุณากรอกข้อมูลที่จำเป็นให้ครบ');
      return;
    }
    const v = form.getFieldsValue(true) as FormValues;
    const invalid = dealError(v, dealTotals(v));
    if (invalid) {
      setError(invalid);
      return;
    }
    if (booking && !booking.documentsApprovedAt) {
      modal.confirm({
        title: 'ยืนยันบันทึกการแก้ไข?',
        content: (
          <>
            <p>{financeWarning(booking)}</p>
            <p className="mb-0 text-ink-muted">แจ้งฝ่ายการเงินด้วยว่ามีการแก้ไขใบจอง เพื่อออกเอกสารตามยอดใหม่</p>
          </>
        ),
        okText: 'ยืนยันบันทึก',
        cancelText: 'กลับไปแก้ไข',
        onOk: () => save(v),
      });
      return;
    }
    await save(v);
  };

  const travellers = totals.travellers;

  return (
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
      <Form<FormValues>
        form={form}
        layout="vertical"
        requiredMark={false}
        initialValues={initial}
        onValuesChange={(_, all) => setValues(all)}
        className="flex min-w-0 flex-col gap-4"
      >
        {booking ? (
          <ReopenedSummary booking={booking} onCancel={onCancel} onChangePurpose={onChangePurpose} />
        ) : (
          <Card title="วางข้อความสรุปยอดโอน" size="small">
            <SummaryPasteBox rows={3} onApply={applySummary} />
            {pasteReport && (
              <div className="mt-3 flex flex-col gap-1 text-[13px]">
                {pasteReport.filled.length > 0 && <span className="text-green-700">กรอกแล้ว: {pasteReport.filled.join(', ')}</span>}
                {pasteReport.warnings.map((w) => (
                  <span key={w} className="text-amber-700">
                    ⚠ {w}
                  </span>
                ))}
                <span className="text-ink-muted">ตรวจสอบข้อมูลอีกครั้ง แล้วแนบสลิปก่อนบันทึก</span>
              </div>
            )}
          </Card>
        )}

        <Card title="ข้อมูลการจอง" size="small">
          {!editing && (
            <Form.Item label="เซลล์เจ้าของการจอง" name="salesUserId" rules={[{ required: true, message: 'กรุณาเลือกเซลล์เจ้าของการจอง' }]}>
              <Select
                showSearch={{ optionFilterProp: 'label' }}
                placeholder="เลือกเซลล์"
                options={lookups.sales.items.map((u) => ({ value: u.id, label: u.phone ? `${u.name} ${formatPhone(u.phone)}` : u.name }))}
                notFoundContent="ยังไม่มีผู้ใช้บทบาทเซลล์ในระบบ"
              />
            </Form.Item>
          )}
          <div className="grid gap-x-4 sm:grid-cols-3">
            <Form.Item label="ช่องทางการจอง" name="bookingSourceId">
              <Select
                allowClear
                placeholder="ไม่ระบุ"
                options={sourceOptions.map((s) => ({ value: s.id, label: s.name }))}
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

        {!editing && (
          <Card title="โปรแกรมทัวร์" size="small">
            <Form.Item label="โปรแกรม (เฉพาะที่เปิดขาย)">
              <TourPicker value={tourItem} onChange={(t) => void selectTour(t)} state={TourProgramStates.Selling} />
            </Form.Item>
            <Form.Item label="วันเดินทาง" name="periodId" className="mb-0" rules={[{ required: true, message: 'กรุณาเลือกโปรแกรมและรอบเดินทาง' }]}>
              <Select
                placeholder={tourItem ? 'เลือกรอบเดินทาง' : 'เลือกโปรแกรมก่อน'}
                disabled={!tourItem}
                loading={tourLoading}
                onChange={selectPeriod}
                options={periods.map((p) => ({ value: p.id, label: periodLabel(p) }))}
                notFoundContent="ไม่มีรอบเดินทางที่ยังว่าง"
              />
            </Form.Item>
          </Card>
        )}

        <Card title="ผู้เดินทางและราคา" size="small">
          <div className="grid grid-cols-[minmax(0,1fr)_88px_128px] items-center gap-x-3 gap-y-2 max-sm:grid-cols-[minmax(0,1fr)_72px_104px]">
            <span />
            <span className="text-[13px] text-ink-muted">จำนวน</span>
            <span className="text-[13px] text-ink-muted">ราคา/หน่วย (บาท)</span>
            <PriceRow label="ผู้ใหญ่" count="adultCount" price="adultPrice" />
            <PriceRow label="เด็กอายุต่ำกว่า 20 ปี (มีเตียง)" count="childWithBedCount" price="childWithBedPrice" />
            <PriceRow label="เด็กอายุต่ำกว่า 20 ปี (ไม่มีเตียง)" count="childNoBedCount" price="childNoBedPrice" />
            <PriceRow label="ทารก" count="infantCount" price="infantPrice" />
            <PriceRow label="วีซ่า" count="visaCount" price="visaPrice" />
            <PriceRow label="พาสปอร์ตต่างชาติ" count="foreignCount" price="foreignPrice" />
            <PriceRow label="ห้องพักเดี่ยว (ห้อง)" count="singleRoomCount" price="singleRoomPrice" />
            <PriceRow label="Join land" count="joinLandCount" price="joinLandPrice" />
            <PriceRow label="ภาษีน้ำมัน" count="fuelTaxCount" price="fuelTaxPrice" countPlaceholder={String(travellers)} />
            <PriceRow label="ภาษีสนามบิน" count="airportTaxCount" price="airportTaxPrice" countPlaceholder={String(travellers)} />
          </div>

          <Form.List name="options">
            {(fields, { add, remove }) => (
              <div className="mt-3 flex flex-col gap-2">
                {fields.map((field) => (
                  <div
                    key={field.key}
                    className="grid grid-cols-[minmax(0,1fr)_88px_128px_32px] items-center gap-x-3 max-sm:grid-cols-[minmax(0,1fr)_72px_104px_32px]"
                  >
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
                <Button type="dashed" icon={<PlusOutlined />} onClick={() => add({ name: '', qty: travellers || 1, price: null })} className="self-start">
                  เพิ่มรายการ
                </Button>
              </div>
            )}
          </Form.List>

          <Form.Item label="ส่วนลด/ท่าน (บาท)" name="discountPerPerson" className="mt-4 mb-0 w-[180px]" extra="หักจากราคาผู้ใหญ่ เด็ก และทารก">
            <InputNumber min={0} precision={0} className="w-full" placeholder="0" />
          </Form.Item>

          <p className="mt-4 mb-0 text-sm">
            ผู้เดินทางทั้งหมด <span className="font-semibold">{travellers}</span> ท่าน
            {period && <span className="text-ink-muted"> · ที่นั่งว่าง {period.seatsAvailable}</span>}
          </p>
          {period && travellers > period.seatsAvailable && <Alert type="warning" showIcon className="mt-2" title="จำนวนผู้เดินทางมากกว่าที่นั่งว่างของรอบนี้" />}
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
            {editing ? (
              <Form.Item label="แผนการชำระ (เปลี่ยนไม่ได้)">
                <Tag className="m-0">{values.isDeposit ? 'มัดจำ' : 'ชำระเต็มจำนวน'}</Tag>
              </Form.Item>
            ) : (
              <Form.Item label="มัดจำ" name="isDeposit" valuePropName="checked">
                <Switch />
              </Form.Item>
            )}
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
              <Form.Item label="ชำระส่วนที่เหลือภายในวันที่" name="remainingDueDate" className="mb-0" extra="ค่าเริ่มต้น 30 วันก่อนเดินทาง">
                <DatePicker format={DATE_FORMAT} className="w-full" />
              </Form.Item>
            </div>
          ) : (
            <div className="grid grid-cols-[minmax(0,1fr)_110px] gap-x-2 sm:max-w-[50%]">
              <Form.Item label="ชำระเต็มจำนวนภายในวันที่" name="fullDueDate" className="mb-0">
                <DatePicker format={DATE_FORMAT} className="w-full" />
              </Form.Item>
              <Form.Item label="ก่อนเวลา" name="fullDueTime" className="mb-0">
                <TimePicker format="HH:mm" minuteStep={5} needConfirm={false} className="w-full" />
              </Form.Item>
            </div>
          )}
        </Card>

        <Card title="ข้อมูลลูกค้า" size="small">
          <Form.Item
            label="แชท LINE ของลูกค้า (ไม่บังคับ)"
            extra="เลือกเมื่อลูกค้าคุยกับ OA เพื่อส่งใบเสร็จทาง LINE และค้นหาใบจองจากแชทได้ภายหลัง ถ้าคุยผ่าน LINE ส่วนตัวไม่ต้องเลือก"
          >
            <div className="flex items-center gap-2">
              <div className="min-w-0 flex-1">
                <LineContactPicker value={lineContact} onPick={pickContact} />
              </div>
              {lineContact && <OpenChatLink url={values.chatUrl} />}
              {lineContact && (
                <Button type="link" onClick={() => pickContact(null)}>
                  ไม่ผูกแชท
                </Button>
              )}
            </div>
            {lineContact && !lineContact.following && <span className="text-xs text-red-600">ลูกค้าบล็อก OA — ส่งข้อความทาง LINE ไม่ได้</span>}
          </Form.Item>
          {lineContact && (
            <Form.Item
              label="ลิงก์แชทใน LINE OA Manager (ไม่บังคับ)"
              name="chatUrl"
              extra="เปิดแชทของลูกค้าใน chat.line.biz แล้วคัดลอกลิงก์จากแถบที่อยู่มาวาง ระบบจะจำไว้กับแชทนี้ เพื่อให้กด “เปิดแชท” จากใบจองได้"
              rules={[{ pattern: /^https:\/\/chat\.line\.biz\/\S+$/, message: 'ลิงก์ต้องขึ้นต้นด้วย https://chat.line.biz/' }]}
            >
              <Input autoComplete="off" placeholder="https://chat.line.biz/U…/chat/U…" />
            </Form.Item>
          )}
          <div className="grid gap-x-4 sm:grid-cols-2">
            <Form.Item label="ชื่อลูกค้า" name="customerName" rules={[{ required: true, whitespace: true, message: 'กรุณาระบุชื่อลูกค้า' }]}>
              <Input autoComplete="off" placeholder="ชื่อ-นามสกุล" />
            </Form.Item>
            <Form.Item label="เบอร์โทรลูกค้า" name="customerPhone" rules={[{ required: true, whitespace: true, message: 'กรุณาระบุเบอร์โทรลูกค้า' }]}>
              <Input type="tel" autoComplete="off" placeholder="0xx-xxx-xxxx" />
            </Form.Item>
          </div>
          <Form.Item label="ที่อยู่ลูกค้า (ไม่บังคับ)" name="customerAddress" className="mb-0">
            <Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} maxLength={500} />
          </Form.Item>
        </Card>

        {!editing && (
          <Card title="สลิปโอนเงิน" size="small">
            <Form.Item
              name="slipImages"
              className="mb-0"
              extra="เลือกได้หลายรูป (ไม่เกิน 10 รูป)"
              rules={[{ validator: (_, v: string[] | undefined) => (v?.length ? Promise.resolve() : Promise.reject(new Error('กรุณาแนบสลิปโอนเงิน'))) }]}
            >
              <SlipUploadField />
            </Form.Item>
          </Card>
        )}

        <Card title="หมายเหตุ" size="small">
          <Form.Item name="note" className="mb-0">
            <Input.TextArea autoSize={{ minRows: 2, maxRows: 6 }} maxLength={1000} placeholder="รายละเอียดเพิ่มเติม (ถ้ามี)" />
          </Form.Item>
        </Card>
      </Form>

      {/* Phone: the total and the save button stay on a bar at the bottom of the screen. */}
      <div className="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-white px-4 pt-2 pb-[max(0.75rem,env(safe-area-inset-bottom))] shadow-[0_-6px_16px_-10px_rgba(29,43,79,0.3)] lg:hidden">
        {error && <Alert type="error" showIcon className="mb-2" title={error} />}
        <div className="flex items-center gap-3">
          <div className="min-w-0 flex-1 leading-tight">
            <div className="text-xs text-ink-muted">
              {travellers} ท่าน
              {values.isDeposit && ` · มัดจำ ${formatBaht(totals.depositToPay)}`}
            </div>
            <div className="text-lg font-semibold">{formatBaht(totals.grandTotal)}</div>
          </div>
          <Button type="primary" size="large" loading={submitting} onClick={() => void submit()}>
            {editing ? 'บันทึกการแก้ไข' : 'บันทึกการจอง'}
          </Button>
        </div>
      </div>

      {/* Desktop: the summary card stays in view below the page header. */}
      <aside className="flex flex-col gap-3 max-lg:pb-24 lg:sticky lg:top-[72px]">
        <Card title="สรุปยอด" size="small">
          <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
            <Stat label="ผู้เดินทาง" value={`${travellers} ท่าน`} />
            {totals.fee > 0 && <Stat label="ค่าธรรมเนียมบัตร" value={formatBaht(totals.fee)} />}
            {values.isDeposit && <Stat label="มัดจำ" value={formatBaht(totals.depositToPay)} />}
            {values.isDeposit && <Stat label="ส่วนที่เหลือ" value={formatBaht(totals.remaining)} />}
            <Stat label="รวมเป็นเงิน" value={formatBaht(totals.grandTotal)} strong />
          </dl>
          {pastedTotal !== null && pastedTotal !== totals.grandTotal && (
            <Alert type="warning" showIcon className="mt-3" title={`ข้อความสรุป ${formatBaht(pastedTotal)} ไม่ตรงกับยอดในฟอร์ม ตรวจสอบราคาอีกครั้ง`} />
          )}
          {error && <Alert type="error" showIcon className="mt-3 max-lg:hidden" title={error} />}
          <Button type="primary" size="large" block className="mt-4 max-lg:hidden" loading={submitting} onClick={() => void submit()}>
            {editing ? 'บันทึกการแก้ไข' : 'บันทึกการจอง'}
          </Button>
          {editing && (
            <Button block className="mt-2" onClick={onCancel} disabled={submitting}>
              ยกเลิก
            </Button>
          )}
        </Card>
      </aside>
    </div>
  );
}

function Stat({ label, value, strong }: { label: string; value: ReactNode; strong?: boolean }) {
  return (
    <>
      <dt className={strong ? 'font-medium' : 'text-ink-muted'}>{label}</dt>
      <dd className={`m-0 text-right ${strong ? 'text-lg font-semibold' : ''}`}>{value}</dd>
    </>
  );
}

type NumField = { [K in keyof FormValues]: FormValues[K] extends number | null ? K : never }[keyof FormValues];

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
