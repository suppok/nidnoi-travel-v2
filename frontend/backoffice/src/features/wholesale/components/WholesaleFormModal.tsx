import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { App, Button, Form, Input, InputNumber, Modal, Select } from 'antd';
import { useEffect, useState } from 'react';
import { ApiError } from '../../../api/client.ts';
import { wholesalesApi } from '../../../api/wholesales.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { WHOLESALE_STATE_INFO, type WholesaleState, WholesaleStates } from '../../../constants/wholesaleStates.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import type { Wholesale, WholesaleContactInput } from '../../../types/wholesale.ts';

const CODE_RE = /^[A-Z0-9]+(-[A-Z0-9]+)*$/;
const MAX_CONTACTS = 20;

type FormValues = {
  code: string;
  name: string;
  refId: number | null;
  state: WholesaleState;
  contacts: WholesaleContactInput[];
};

type Props = {
  open: boolean;
  wholesale: Wholesale | null; // null = create
  onClose: () => void;
  onSaved: () => void;
};

export default function WholesaleFormModal({ open, wholesale, onClose, onSaved }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);
  const creating = wholesale === null;

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue(
      wholesale
        ? {
            code: wholesale.code,
            name: wholesale.name,
            refId: wholesale.refId,
            state: wholesale.state,
            contacts: wholesale.contacts.map(({ name, phone }) => ({ name, phone })),
          }
        : { code: '', name: '', refId: null, state: WholesaleStates.OnRequest, contacts: [{ name: '', phone: '' }] },
    );
  }, [open, wholesale, form]);

  const onFinish = async (values: FormValues) => {
    setSubmitting(true);
    const body = {
      code: values.code.trim().toUpperCase(),
      name: values.name.trim(),
      refId: values.refId ?? null,
      state: values.state,
      contacts: (values.contacts ?? []).map((c) => ({ name: c.name.trim(), phone: c.phone.trim() })),
    };
    try {
      if (creating) await wholesalesApi.create(body);
      else await wholesalesApi.update({ ...body, id: wholesale.id });
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
      title={creating ? 'เพิ่มโฮลเซลล์' : 'แก้ไขโฮลเซลล์'}
      okText={creating ? 'เพิ่มโฮลเซลล์' : 'บันทึก'}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={640}
    >
      <Form<FormValues> form={form} layout="vertical" requiredMark={false} onFinish={onFinish} className="pt-2">
        <div className="grid grid-cols-[1fr_2fr] gap-x-4 max-sm:grid-cols-1">
          <Form.Item
            label="รหัส"
            name="code"
            normalize={(v?: string) => (v ?? '').toUpperCase()}
            rules={[
              { required: true, whitespace: true, message: 'กรุณากรอกรหัส' },
              { max: 20, message: 'ยาวได้ไม่เกิน 20 ตัวอักษร' },
              { pattern: CODE_RE, message: 'ใช้ได้เฉพาะ A-Z, 0-9 และขีด (-)' },
            ]}
          >
            <Input autoComplete="off" spellCheck={false} className="font-mono" placeholder="GO365" />
          </Form.Item>
          <Form.Item label="ชื่อโฮลเซลล์" name="name" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อ' }, { max: 200, message: 'ยาวได้ไม่เกิน 200 ตัวอักษร' }]}>
            <Input autoComplete="off" />
          </Form.Item>
        </div>

        <div className="grid grid-cols-2 gap-x-4 max-sm:grid-cols-1">
          <Form.Item label="สถานะการเชื่อมต่อ" name="state" rules={[{ required: true, message: 'กรุณาเลือกสถานะ' }]}>
            <Select options={WHOLESALE_STATE_INFO.map((s) => ({ value: s.value, label: s.label }))} />
          </Form.Item>
          <Form.Item label="Ref ID" name="refId" extra="รหัสจากเว็บเดิม (ไม่บังคับ)">
            <InputNumber min={1} precision={0} className="w-full" placeholder="—" />
          </Form.Item>
        </div>

        <Form.List
          name="contacts"
          rules={[
            {
              validator: (_, contacts: WholesaleContactInput[] | undefined) =>
                (contacts?.length ?? 0) > MAX_CONTACTS ? Promise.reject(new Error(`มีผู้ติดต่อได้ไม่เกิน ${MAX_CONTACTS} คน`)) : Promise.resolve(),
            },
          ]}
        >
          {(fields, { add, remove }, { errors }) => (
            <div className="flex flex-col gap-2">
              <span className="font-medium">ผู้ติดต่อ</span>
              {fields.length === 0 && <span className="text-[13px] text-ink-muted">ยังไม่มีผู้ติดต่อ</span>}
              {fields.map((field) => (
                <div key={field.key} className="flex items-start gap-2">
                  <Form.Item
                    name={[field.name, 'name']}
                    className="mb-0 flex-1"
                    rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อ' }, { max: 100, message: 'ยาวได้ไม่เกิน 100 ตัวอักษร' }]}
                  >
                    <Input placeholder="ชื่อผู้ติดต่อ" autoComplete="off" />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'phone']}
                    className="mb-0 flex-1"
                    rules={[{ required: true, whitespace: true, message: 'กรุณากรอกเบอร์โทร' }, { max: 50, message: 'ยาวได้ไม่เกิน 50 ตัวอักษร' }]}
                  >
                    <Input placeholder="เบอร์โทร เช่น 081-234-5678" autoComplete="off" inputMode="tel" />
                  </Form.Item>
                  <Button type="text" danger icon={<DeleteOutlined />} onClick={() => remove(field.name)} aria-label="ลบผู้ติดต่อ" />
                </div>
              ))}
              <Form.ErrorList errors={errors} />
              <Button
                type="dashed"
                icon={<PlusOutlined />}
                onClick={() => add({ name: '', phone: '' })}
                disabled={fields.length >= MAX_CONTACTS}
                className="self-start"
              >
                เพิ่มผู้ติดต่อ
              </Button>
            </div>
          )}
        </Form.List>
      </Form>
    </Modal>
  );
}
