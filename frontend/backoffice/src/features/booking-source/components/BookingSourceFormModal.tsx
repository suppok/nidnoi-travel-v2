import { App, Form, Input, InputNumber, Modal } from 'antd';
import { useEffect, useState } from 'react';
import { bookingSourcesApi } from '../../../api/bookingSources.ts';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { BookingSource } from '../../../types/bookingSource.ts';

type FormValues = {
  name: string;
  shortName: string;
  sortOrder: number | null;
};

type Props = {
  open: boolean;
  source: BookingSource | null; // null = create
  onClose: () => void;
  onSaved: () => void;
};

export default function BookingSourceFormModal({ open, source, onClose, onSaved }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);
  const creating = source === null;

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue(
      source ? { name: source.name, shortName: source.shortName, sortOrder: source.sortOrder } : { name: '', shortName: '', sortOrder: 0 },
    );
  }, [open, source, form]);

  const onFinish = async (values: FormValues) => {
    setSubmitting(true);
    const body = { name: values.name.trim(), shortName: (values.shortName ?? '').trim(), sortOrder: values.sortOrder ?? 0 };
    try {
      if (creating) await bookingSourcesApi.create(body);
      else await bookingSourcesApi.update({ ...body, id: source.id });
      message.success('บันทึกแล้ว');
      onSaved();
    } catch (err) {
      if (err instanceof ApiError && err.code === ErrorCodes.NameTaken) {
        form.setFields([{ name: 'name', errors: [errorMessage(err)] }]);
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
      title={creating ? 'เพิ่มช่องทางการจอง' : 'แก้ไขช่องทางการจอง'}
      okText={creating ? 'เพิ่มช่องทาง' : 'บันทึก'}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={520}
    >
      <Form<FormValues> form={form} layout="vertical" requiredMark={false} onFinish={onFinish} className="pt-2">
        <Form.Item label="ชื่อช่องทาง" name="name" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อ' }, { max: 100, message: 'ยาวได้ไม่เกิน 100 ตัวอักษร' }]}>
          <Input autoComplete="off" placeholder="LINE OA" />
        </Form.Item>
        <Form.Item
          label="ชื่อย่อ"
          name="shortName"
          extra='ใช้เป็นหัวข้อชื่อในข้อความสรุปการจอง เช่น "ชื่อ LINE" (ไม่บังคับ)'
          rules={[{ max: 50, message: 'ยาวได้ไม่เกิน 50 ตัวอักษร' }]}
        >
          <Input autoComplete="off" placeholder="LINE" />
        </Form.Item>
        <Form.Item label="ลำดับ" name="sortOrder" extra="น้อยขึ้นก่อน" className="max-w-[160px]">
          <InputNumber min={0} max={9999} precision={0} className="w-full" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
