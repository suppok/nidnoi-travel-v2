import { App, Form, Input, InputNumber, Modal } from 'antd';
import { useEffect, useState } from 'react';
import { airlinesApi } from '../../../api/airlines.ts';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { Airline } from '../../../types/airline.ts';
import ImageUploadField from '../../tour-program/components/ImageUploadField.tsx';

const CODE_RE = /^[A-Z0-9]{2,3}$/;

type FormValues = {
  code: string;
  name: string;
  refId: number | null;
  logo: string | null;
};

type Props = {
  open: boolean;
  airline: Airline | null; // null = create
  onClose: () => void;
  onSaved: () => void;
};

export default function AirlineFormModal({ open, airline, onClose, onSaved }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);
  const creating = airline === null;

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue(
      airline
        ? { code: airline.code, name: airline.name, refId: airline.refId, logo: airline.logo }
        : { code: '', name: '', refId: null, logo: null },
    );
  }, [open, airline, form]);

  const onFinish = async (values: FormValues) => {
    setSubmitting(true);
    const body = {
      code: values.code.trim().toUpperCase(),
      name: values.name.trim(),
      refId: values.refId ?? null,
      logo: values.logo ?? null,
    };
    try {
      if (creating) await airlinesApi.create(body);
      else await airlinesApi.update({ ...body, id: airline.id });
      message.success('บันทึกแล้ว');
      onSaved();
    } catch (err) {
      if (err instanceof ApiError && (err.code === ErrorCodes.CodeTaken || err.code === ErrorCodes.RefIdTaken)) {
        form.setFields([{ name: err.code === ErrorCodes.CodeTaken ? 'code' : 'refId', errors: [errorMessage(err)] }]);
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
      title={creating ? 'เพิ่มสายการบิน' : 'แก้ไขสายการบิน'}
      okText={creating ? 'เพิ่มสายการบิน' : 'บันทึก'}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={560}
    >
      <Form<FormValues> form={form} layout="vertical" requiredMark={false} onFinish={onFinish} className="pt-2">
        <div className="grid grid-cols-[1fr_2fr] gap-x-4 max-sm:grid-cols-1">
          <Form.Item
            label="รหัส"
            name="code"
            extra="IATA 2 ตัว เช่น TG (ใช้ ICAO 3 ตัวถ้าไม่มี IATA)"
            normalize={(v?: string) => (v ?? '').toUpperCase()}
            rules={[
              { required: true, whitespace: true, message: 'กรุณากรอกรหัส' },
              { pattern: CODE_RE, message: 'ใช้ได้ 2–3 ตัว เฉพาะ A-Z และ 0-9' },
            ]}
          >
            <Input autoComplete="off" spellCheck={false} maxLength={3} className="font-mono" placeholder="TG" />
          </Form.Item>
          <Form.Item label="ชื่อสายการบิน" name="name" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อ' }, { max: 200, message: 'ยาวได้ไม่เกิน 200 ตัวอักษร' }]}>
            <Input autoComplete="off" placeholder="Thai Airways" />
          </Form.Item>
        </div>

        <Form.Item label="Ref ID" name="refId" extra="รหัสจากเว็บเดิม (ไม่บังคับ)" className="max-w-[200px]">
          <InputNumber min={1} precision={0} className="w-full" placeholder="—" />
        </Form.Item>

        <Form.Item label="โลโก้" name="logo" extra="แนะนำ PNG พื้นใส">
          <LogoField />
        </Form.Item>
      </Form>
    </Modal>
  );
}

// Form.Item injects value/onChange into its direct child, so the sizing wrapper lives inside the control.
function LogoField(props: { value?: string | null; onChange?: (url: string | null) => void }) {
  return (
    <div className="max-w-[280px]">
      <ImageUploadField {...props} purpose="airline-logo" aspect="3 / 1" fit="contain" />
    </div>
  );
}
