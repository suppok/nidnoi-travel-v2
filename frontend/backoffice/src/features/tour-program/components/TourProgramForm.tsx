import { App, Button, Card, Form, Input, InputNumber, Select } from 'antd';
import { type ReactNode, useEffect, useMemo, useState } from 'react';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { TOUR_PROGRAM_STATE_INFO, type TourProgramState, TourProgramStates } from '../../../constants/tourProgramStates.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { TourProgram, TourProgramCreateRequest } from '../../../types/tourProgram.ts';
import { useCountryOptions } from '../../taxonomy/hooks/useCountryOptions.ts';
import { useAirlineOptions } from '../hooks/useAirlineOptions.ts';
import { useCityOptions } from '../hooks/useCityOptions.ts';
import { useWholesaleOptions } from '../hooks/useWholesaleOptions.ts';
import DocumentUploadField from './DocumentUploadField.tsx';
import GalleryField, { MAX_GALLERY_IMAGES } from './GalleryField.tsx';
import ImageUploadField from './ImageUploadField.tsx';

type FormValues = Omit<TourProgramCreateRequest, 'refId' | 'days' | 'nights'> & {
  refId: number | null;
  days: number | null;
  nights: number | null;
};

const EMPTY: FormValues = {
  refId: null,
  name: '',
  wholesaleId: '',
  wholesaleTourCode: '',
  state: TourProgramStates.Draft,
  days: null,
  nights: null,
  airlineId: null,
  highlight: '',
  description: '',
  bannerImage: null,
  cardImage: null,
  pdfFile: null,
  docxFile: null,
  internalNote: '',
  countryIds: [],
  cityIds: [],
  images: [],
};

function toValues(tour: TourProgram): FormValues {
  return {
    refId: tour.refId,
    name: tour.name,
    wholesaleId: tour.wholesale.id,
    wholesaleTourCode: tour.wholesaleTourCode,
    state: tour.state,
    days: tour.days,
    nights: tour.nights,
    airlineId: tour.airline?.id ?? null,
    highlight: tour.highlight,
    description: tour.description,
    bannerImage: tour.bannerImage,
    cardImage: tour.cardImage,
    pdfFile: tour.pdfFile,
    docxFile: tour.docxFile,
    internalNote: tour.internalNote,
    countryIds: tour.countries.map((c) => c.id),
    cityIds: tour.cities.map((c) => c.id),
    images: tour.images.map((i) => i.url),
  };
}

type Props = {
  tour: TourProgram | null; // null = create
  readOnly?: boolean;
  submitText: string;
  onSubmit: (body: TourProgramCreateRequest) => Promise<void>;
  extraActions?: ReactNode;
};

function Section({ title, extra, children }: { title: string; extra?: ReactNode; children: ReactNode }) {
  return (
    <Card title={title} extra={extra}>
      {children}
    </Card>
  );
}

export default function TourProgramForm({ tour, readOnly = false, submitText, onSubmit, extraActions }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);
  const countryIds = Form.useWatch('countryIds', form) ?? [];

  const { countries, loading: countriesLoading } = useCountryOptions();
  const { wholesales, loading: wholesalesLoading } = useWholesaleOptions();
  const { airlines, loading: airlinesLoading } = useAirlineOptions();
  const { cities, loading: citiesLoading } = useCityOptions(countryIds);

  useEffect(() => {
    form.resetFields();
    form.setFieldsValue(tour ? toValues(tour) : EMPTY);
  }, [tour, form]);

  // Deleted options stay selectable only when the tour already uses them.
  const initialCountries = useMemo(() => new Set(tour?.countries.map((c) => c.id)), [tour]);
  const initialCities = useMemo(() => new Set(tour?.cities.map((c) => c.id)), [tour]);
  const countryOptions = countries
    .filter((c) => c.status || initialCountries.has(c.id))
    .map((c) => ({ value: c.id, label: `${c.nameTh} (${c.nameEn})${c.status ? '' : ' — ลบแล้ว'}` }));
  const wholesaleOptions = wholesales
    .filter((w) => w.status || w.id === tour?.wholesale.id)
    .map((w) => ({ value: w.id, label: `${w.name} (${w.code})${w.status ? '' : ' — ลบแล้ว'}` }));
  const airlineOptions = airlines
    .filter((a) => a.status || a.id === tour?.airline?.id)
    .map((a) => ({ value: a.id, label: `${a.name} (${a.code})${a.status ? '' : ' — ลบแล้ว'}`, logo: a.logo }));
  const cityOptions = countryIds.map((countryId) => {
    const country = countries.find((c) => c.id === countryId);
    return {
      label: country?.nameTh ?? '…',
      options: cities
        .filter((c) => c.countryId === countryId && (c.status || initialCities.has(c.id)))
        .map((c) => ({ value: c.id, label: `${c.nameTh} (${c.nameEn})${c.status ? '' : ' — ลบแล้ว'}` })),
    };
  });

  // City → country, from loaded options plus the tour's own cities (loaded before the options).
  const cityCountry = useMemo(() => {
    const map = new Map<string, string>(tour?.cities.map((c) => [c.id, c.countryId]));
    for (const c of cities) map.set(c.id, c.countryId);
    return map;
  }, [cities, tour]);

  const onValuesChange = (changed: Partial<FormValues>, all: FormValues) => {
    if (changed.countryIds) {
      // Removing a country removes its cities.
      const keep = all.cityIds.filter((id) => {
        const countryId = cityCountry.get(id);
        return !countryId || changed.countryIds?.includes(countryId);
      });
      if (keep.length !== all.cityIds.length) form.setFieldValue('cityIds', keep);
    }
    if (changed.days && (all.nights === null || all.nights >= changed.days)) {
      form.setFieldValue('nights', Math.max(0, changed.days - 1));
    }
  };

  const onFinish = async (values: FormValues) => {
    setSubmitting(true);
    try {
      await onSubmit({
        ...values,
        refId: values.refId ?? null,
        days: values.days ?? 0,
        nights: values.nights ?? 0,
        name: values.name.trim(),
        wholesaleTourCode: values.wholesaleTourCode.trim(),
        airlineId: values.airlineId ?? null,
        highlight: values.highlight.trim(),
        description: values.description.trim(),
        internalNote: values.internalNote.trim(),
      });
    } catch (err) {
      if (err instanceof ApiError && err.code === ErrorCodes.RefIdTaken) {
        form.setFields([{ name: 'refId', errors: [errorMessage(err)] }]);
      } else if (!applyFieldErrors(form, err)) {
        message.error(errorMessage(err));
      } else {
        message.error('ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบช่องที่มีเครื่องหมายสีแดง');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form<FormValues>
      form={form}
      layout="vertical"
      requiredMark={false}
      onFinish={onFinish}
      onValuesChange={onValuesChange}
      onFinishFailed={() => message.error('กรุณากรอกข้อมูลให้ครบ')}
      disabled={readOnly}
      scrollToFirstError
      className="flex flex-col gap-6"
    >
      <Section title="ข้อมูลโปรแกรม">
        <Form.Item label="ชื่อโปรแกรม" name="name" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อโปรแกรม' }, { max: 200, message: 'ยาวได้ไม่เกิน 200 ตัวอักษร' }]}>
          <Input placeholder="เช่น ทัวร์ญี่ปุ่น โตเกียว ฟูจิ 5 วัน 3 คืน" autoComplete="off" showCount maxLength={200} />
        </Form.Item>

        <div className="grid grid-cols-3 gap-x-4 max-lg:grid-cols-2 max-sm:grid-cols-1">
          <Form.Item label="สถานะโปรแกรม" name="state" rules={[{ required: true, message: 'กรุณาเลือกสถานะ' }]}>
            <Select<TourProgramState>
              options={TOUR_PROGRAM_STATE_INFO.map((s) => ({
                value: s.value,
                label: (
                  <span className="flex flex-col leading-tight">
                    <span>{s.label}</span>
                    <span className="text-xs text-ink-muted">{s.hint}</span>
                  </span>
                ),
              }))}
              labelRender={({ value }) => TOUR_PROGRAM_STATE_INFO.find((s) => s.value === value)?.label ?? String(value)}
            />
          </Form.Item>
          <Form.Item label="จำนวนวัน" name="days" rules={[{ required: true, message: 'กรุณากรอกจำนวนวัน' }]}>
            <InputNumber min={1} max={60} precision={0} className="w-full" suffix="วัน" />
          </Form.Item>
          <Form.Item
            label="จำนวนคืน"
            name="nights"
            dependencies={['days']}
            rules={[
              { required: true, message: 'กรุณากรอกจำนวนคืน' },
              ({ getFieldValue }) => ({
                validator: (_, v: number | null) =>
                  v === null || v === undefined || v <= (getFieldValue('days') ?? 0) ? Promise.resolve() : Promise.reject(new Error('จำนวนคืนต้องไม่เกินจำนวนวัน')),
              }),
            ]}
          >
            <InputNumber min={0} max={60} precision={0} className="w-full" suffix="คืน" />
          </Form.Item>
        </div>

        <div className="grid grid-cols-3 gap-x-4 max-lg:grid-cols-2 max-sm:grid-cols-1">
          <Form.Item label="โฮลเซลล์" name="wholesaleId" rules={[{ required: true, message: 'กรุณาเลือกโฮลเซลล์' }]}>
            <Select showSearch optionFilterProp="label" options={wholesaleOptions} loading={wholesalesLoading} placeholder="เลือกโฮลเซลล์" />
          </Form.Item>
          <Form.Item label="รหัสทัวร์ของโฮลเซลล์" name="wholesaleTourCode" rules={[{ max: 50, message: 'ยาวได้ไม่เกิน 50 ตัวอักษร' }]} extra="เห็นเฉพาะพนักงาน">
            <Input autoComplete="off" spellCheck={false} className="font-mono" placeholder="เช่น GO1TYO-XJ001" />
          </Form.Item>
          <Form.Item label="สายการบิน" name="airlineId" extra="ไม่บังคับ (เว้นว่างได้ถ้าไม่มีเที่ยวบิน)">
            <Select
              showSearch
              allowClear
              optionFilterProp="label"
              options={airlineOptions}
              loading={airlinesLoading}
              placeholder="เลือกสายการบิน"
              optionRender={(option) => (
                <span className="flex items-center gap-2">
                  <span className="flex h-5 w-12 shrink-0 items-center justify-center">
                    {option.data.logo && <img src={option.data.logo} alt="" className="max-h-full max-w-full object-contain" />}
                  </span>
                  {option.label}
                </span>
              )}
            />
          </Form.Item>
        </div>

        <Form.Item label="Ref ID" name="refId" extra="รหัสจากเว็บเดิม (ไม่บังคับ)" className="max-w-[240px]">
          <InputNumber min={1} precision={0} className="w-full" placeholder="—" />
        </Form.Item>
      </Section>

      <Section title="ปลายทาง">
        <div className="grid grid-cols-2 gap-x-4 max-md:grid-cols-1">
          <Form.Item
            label="ประเทศ"
            name="countryIds"
            extra="ประเทศแรกคือประเทศหลักของทัวร์"
            rules={[{ required: true, type: 'array', min: 1, message: 'กรุณาเลือกอย่างน้อย 1 ประเทศ' }]}
          >
            <Select mode="multiple" showSearch optionFilterProp="label" options={countryOptions} loading={countriesLoading} placeholder="เลือกประเทศ" />
          </Form.Item>
          <Form.Item label="เมือง" name="cityIds" extra={countryIds.length === 0 ? 'เลือกประเทศก่อน' : 'ไม่บังคับ'}>
            <Select
              mode="multiple"
              showSearch
              optionFilterProp="label"
              options={cityOptions}
              loading={citiesLoading}
              disabled={readOnly || countryIds.length === 0}
              placeholder="เลือกเมือง"
            />
          </Form.Item>
        </div>
      </Section>

      <Section title="รายละเอียด">
        <Form.Item label="ไฮไลต์" name="highlight" extra="สรุปสั้น ๆ แสดงบนการ์ดทัวร์และผลการค้นหา Google" rules={[{ max: 500, message: 'ยาวได้ไม่เกิน 500 ตัวอักษร' }]}>
          <Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} showCount maxLength={500} />
        </Form.Item>
        <Form.Item label="รายละเอียดโปรแกรม" name="description" rules={[{ max: 20000, message: 'ยาวได้ไม่เกิน 20,000 ตัวอักษร' }]}>
          <Input.TextArea autoSize={{ minRows: 6, maxRows: 20 }} showCount maxLength={20000} />
        </Form.Item>
        <Form.Item label="บันทึกภายใน" name="internalNote" extra="เห็นเฉพาะพนักงาน ไม่แสดงบนหน้าเว็บ" rules={[{ max: 5000, message: 'ยาวได้ไม่เกิน 5,000 ตัวอักษร' }]}>
          <Input.TextArea autoSize={{ minRows: 2, maxRows: 8 }} />
        </Form.Item>
      </Section>

      <Section title="รูปภาพ">
        <div className="grid grid-cols-[3fr_2fr] gap-x-6 max-lg:grid-cols-1">
          <Form.Item label="แบนเนอร์" name="bannerImage" extra="รูปใหญ่ด้านบนหน้ารายละเอียดทัวร์ และรูปตอนแชร์ลิงก์">
            <ImageUploadField purpose="tour-banner" disabled={readOnly} />
          </Form.Item>
          <Form.Item label="รูปการ์ดทัวร์" name="cardImage" extra="รูปบนการ์ดทัวร์ในหน้าแรกและหน้ารายการทัวร์ (แนวนอน 4:3)">
            <ImageUploadField purpose="tour-card" aspect="4 / 3" disabled={readOnly} />
          </Form.Item>
        </div>
        <Form.Item
          label="แกลเลอรี"
          name="images"
          rules={[{ type: 'array', max: MAX_GALLERY_IMAGES, message: `มีรูปได้ไม่เกิน ${MAX_GALLERY_IMAGES} รูป` }]}
          className="mb-0"
        >
          <GalleryField disabled={readOnly} />
        </Form.Item>
      </Section>

      <Section title="เอกสารโปรแกรม">
        <div className="grid grid-cols-2 gap-x-4 max-md:grid-cols-1">
          <Form.Item label="ไฟล์ PDF" name="pdfFile" extra="ลูกค้าดาวน์โหลดได้จากหน้าทัวร์">
            <DocumentUploadField purpose="tour-pdf" disabled={readOnly} />
          </Form.Item>
          <Form.Item label="ไฟล์ Word (DOCX)" name="docxFile" extra="ไฟล์ต้นฉบับสำหรับแก้ไข เห็นเฉพาะพนักงาน">
            <DocumentUploadField purpose="tour-docx" disabled={readOnly} />
          </Form.Item>
        </div>
      </Section>

      {!readOnly && (
        <div className="sticky bottom-0 z-10 -mx-8 flex justify-end gap-2.5 border-t border-line bg-white/90 px-8 py-3 backdrop-blur max-md:-mx-4 max-md:px-4">
          {extraActions}
          <Button type="primary" htmlType="submit" loading={submitting}>
            {submitText}
          </Button>
        </div>
      )}
    </Form>
  );
}
