import { LockOutlined } from '@ant-design/icons';
import { App, Form, Input, InputNumber, Modal, Select, Tooltip, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { citiesApi, countriesApi } from '../../../api/taxonomy.ts';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { City, Country } from '../../../types/taxonomy.ts';

const SLUG_RE = /^[a-z0-9]+(-[a-z0-9]+)*$/;

// "Ho Chi Minh City" → "ho-chi-minh-city"; a suggestion only, staff can edit it before saving.
function slugify(text: string): string {
  return text
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 60)
    .replace(/-$/, '');
}

type FormValues = {
  countryId?: string;
  slug: string;
  nameTh: string;
  nameEn: string;
  refId: number | null;
  sortOrder: number;
  description: string;
};

type Props =
  | { kind: 'country'; open: boolean; item: Country | null; countries?: never; defaultCountryId?: never; onClose: () => void; onSaved: () => void }
  | {
      kind: 'city';
      open: boolean;
      item: City | null;
      countries: Country[];
      defaultCountryId?: string; // preselected when creating
      onClose: () => void;
      onSaved: () => void;
    };

const LABEL = { country: 'ประเทศ', city: 'เมือง' } as const;

export default function TaxonomyFormModal(props: Props) {
  const { kind, open, item, defaultCountryId, onClose, onSaved } = props;
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);
  const slugTouched = useRef(false);
  const creating = item === null;
  const countries = props.kind === 'city' ? props.countries : [];

  const slug = Form.useWatch('slug', form) ?? '';
  const countryId = Form.useWatch('countryId', form);
  const countrySlug = countries.find((c) => c.id === (creating ? countryId : (item as City | null)?.countryId))?.slug;
  const shownSlug = creating ? slug : (item?.slug ?? '');
  const urlPreview = kind === 'country' ? `/tours/${shownSlug || '…'}/` : `/tours/${countrySlug ?? '…'}/${shownSlug || '…'}/`;

  useEffect(() => {
    if (!open) return;
    slugTouched.current = false;
    form.resetFields();
    form.setFieldsValue(
      item
        ? {
            countryId: kind === 'city' ? (item as City).countryId : undefined,
            slug: item.slug,
            nameTh: item.nameTh,
            nameEn: item.nameEn,
            refId: item.refId,
            sortOrder: item.sortOrder,
            description: item.description,
          }
        : { countryId: defaultCountryId, slug: '', nameTh: '', nameEn: '', refId: null, sortOrder: 0, description: '' },
    );
  }, [open, item, kind, defaultCountryId, form]);

  const onValuesChange = (changed: Partial<FormValues>) => {
    if ('slug' in changed) slugTouched.current = true;
    if (creating && 'nameEn' in changed && !slugTouched.current) form.setFieldValue('slug', slugify(changed.nameEn ?? ''));
  };

  const onFinish = async (values: FormValues) => {
    setSubmitting(true);
    const fields = {
      refId: values.refId ?? null,
      nameTh: values.nameTh.trim(),
      nameEn: values.nameEn.trim(),
      description: (values.description ?? '').trim(),
      heroImage: item?.heroImage ?? null, // edited through media upload (not built yet); keep as is
      sortOrder: values.sortOrder ?? 0,
    };
    try {
      if (kind === 'country') {
        if (creating) await countriesApi.create({ ...fields, slug: values.slug });
        else await countriesApi.update({ ...fields, id: item.id });
      } else {
        const city = { ...fields, countryId: values.countryId ?? '' };
        if (creating) await citiesApi.create({ ...city, slug: values.slug });
        else await citiesApi.update({ ...city, id: item.id });
      }
      message.success('บันทึกแล้ว');
      onSaved();
    } catch (err) {
      if (err instanceof ApiError && (err.code === ErrorCodes.SlugTaken || err.code === ErrorCodes.RefIdTaken)) {
        form.setFields([{ name: err.code === ErrorCodes.SlugTaken ? 'slug' : 'refId', errors: [errorMessage(err)] }]);
      } else if (!applyFieldErrors(form, err)) {
        message.error(errorMessage(err));
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={open}
      title={creating ? `เพิ่ม${LABEL[kind]}` : `แก้ไข${LABEL[kind]}`}
      okText={creating ? `เพิ่ม${LABEL[kind]}` : 'บันทึก'}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={600}
    >
      <Form<FormValues> form={form} layout="vertical" requiredMark={false} onFinish={onFinish} onValuesChange={onValuesChange} className="pt-2">
        {kind === 'city' && (
          <Form.Item label="ประเทศ" name="countryId" rules={[{ required: true, message: 'กรุณาเลือกประเทศ' }]}>
            <Select
              showSearch
              optionFilterProp="label"
              placeholder="เลือกประเทศ"
              options={countries
                .filter((c) => c.status || c.id === (item as City | null)?.countryId)
                .map((c) => ({ value: c.id, label: `${c.nameTh} (${c.nameEn})`, disabled: !c.status }))}
            />
          </Form.Item>
        )}

        <div className="grid grid-cols-2 gap-x-4 max-sm:grid-cols-1">
          <Form.Item label="ชื่อภาษาไทย" name="nameTh" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อภาษาไทย' }, { max: 100, message: 'ยาวได้ไม่เกิน 100 ตัวอักษร' }]}>
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item label="ชื่อภาษาอังกฤษ" name="nameEn" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อภาษาอังกฤษ' }, { max: 100, message: 'ยาวได้ไม่เกิน 100 ตัวอักษร' }]}>
            <Input autoComplete="off" />
          </Form.Item>
        </div>

        {creating ? (
          <Form.Item
            label="Slug"
            name="slug"
            extra={
              <span>
                ใช้ใน URL หน้าลิสต์ทัวร์ เช่น <code>{urlPreview}</code> เปลี่ยนภายหลังไม่ได้
              </span>
            }
            rules={[
              { required: true, message: 'กรุณากรอก slug' },
              { max: 60, message: 'ยาวได้ไม่เกิน 60 ตัวอักษร' },
              { pattern: SLUG_RE, message: 'ใช้ได้เฉพาะ a-z, 0-9 และขีด (-) เช่น ho-chi-minh-city' },
            ]}
          >
            <Input autoComplete="off" spellCheck={false} className="font-mono" />
          </Form.Item>
        ) : (
          <Form.Item label="Slug" extra={<span>URL: <code>{urlPreview}</code></span>}>
            <Tooltip title="Slug ถูกล็อกหลังสร้างเพราะใช้ในลิงก์ภายนอกและ Google index">
              <Typography.Text code>
                <LockOutlined className="mr-1" />
                {item.slug}
              </Typography.Text>
            </Tooltip>
          </Form.Item>
        )}

        <div className="grid grid-cols-2 gap-x-4 max-sm:grid-cols-1">
          <Form.Item label="Ref ID" name="refId" extra="รหัสจากเว็บเดิม (ไม่บังคับ)">
            <InputNumber min={1} precision={0} className="w-full" placeholder="—" />
          </Form.Item>
          <Form.Item label="ลำดับการแสดง" name="sortOrder" extra="น้อยแสดงก่อน" rules={[{ required: true, message: 'กรุณากรอกลำดับ' }]}>
            <InputNumber min={0} max={100000} precision={0} className="w-full" />
          </Form.Item>
        </div>

        <Form.Item label="คำอธิบาย" name="description" extra="ข้อความแนะนำบนหน้า landing ของลูกค้า (ช่วยเรื่อง SEO)" rules={[{ max: 5000, message: 'ยาวได้ไม่เกิน 5000 ตัวอักษร' }]}>
          <Input.TextArea autoSize={{ minRows: 3, maxRows: 8 }} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
