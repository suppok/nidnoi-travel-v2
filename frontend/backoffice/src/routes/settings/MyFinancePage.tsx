import { Alert, App, Button, Card, Form, Select, Skeleton } from 'antd';
import { useEffect, useState } from 'react';
import { salesFinancesApi } from '../../api/salesFinances.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { applyFieldErrors, errorMessage } from '../../lib/errors.ts';
import type { SalesFinanceSettings } from '../../types/salesFinance.ts';

type FormValues = { financeUserIds: string[]; defaultFinanceUserId?: string };

const toForm = (s: SalesFinanceSettings): FormValues => ({ financeUserIds: s.financeUserIds, defaultFinanceUserId: s.defaultFinanceUserId || undefined });

// /settings/finance — the sales user picks the finance users their confirmed LINE bookings can go to,
// and the one preselected in the confirm dialog.
export default function MyFinancePage() {
  const { message } = App.useApp();
  const [form] = Form.useForm<FormValues>();
  const [settings, setSettings] = useState<SalesFinanceSettings | null>(null);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [saving, setSaving] = useState(false);
  const selected = Form.useWatch('financeUserIds', form) ?? [];

  const apply = (s: SalesFinanceSettings) => {
    setSettings(s);
    form.setFieldsValue(toForm(s));
  };

  useEffect(() => {
    let cancelled = false;
    salesFinancesApi
      .get()
      .then((s) => {
        if (cancelled) return;
        setSettings(s);
        form.setFieldsValue(toForm(s));
      })
      .catch((err: unknown) => !cancelled && setLoadError(err));
    return () => {
      cancelled = true;
    };
  }, [form]);

  const save = async (values: FormValues) => {
    setSaving(true);
    try {
      const ids = values.financeUserIds ?? [];
      const defaultId = values.defaultFinanceUserId && ids.includes(values.defaultFinanceUserId) ? values.defaultFinanceUserId : '';
      apply(await salesFinancesApi.update({ financeUserIds: ids, defaultFinanceUserId: defaultId }));
      message.success('บันทึกฝ่ายการเงินแล้ว');
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const nameOf = (id: string) => settings?.options.find((o) => o.id === id)?.name ?? id;

  return (
    <>
      <PageHeader title="ฝ่ายการเงินของฉัน" subtitle="ฝ่ายการเงินที่คุณส่งใบจอง LINE ให้ออกใบเสร็จ เมื่อกดยืนยันการชำระเงิน" />
      <Card className="max-w-xl">
        {loadError !== null && <Alert type="error" showIcon title={errorMessage(loadError)} />}
        {!settings && loadError === null && <Skeleton active paragraph={{ rows: 3 }} />}
        {settings && settings.options.length === 0 && <Alert type="info" showIcon title="ยังไม่มีผู้ใช้ฝ่ายการเงินในระบบ" description="ให้แอดมินเพิ่มผู้ใช้ที่มีบทบาท ฝ่ายการเงิน ก่อน" />}
        {settings && settings.options.length > 0 && (
          <Form<FormValues> form={form} layout="vertical" onFinish={save}>
            <Form.Item name="financeUserIds" label="ฝ่ายการเงิน" extra="เลือกได้หลายคน ตอนยืนยันใบจองจะเลือกจากรายชื่อนี้">
              <Select
                mode="multiple"
                placeholder="เลือกฝ่ายการเงิน"
                optionFilterProp="label"
                options={settings.options.map((o) => ({ value: o.id, label: o.name }))}
                onChange={(ids: string[]) => {
                  const current = form.getFieldValue('defaultFinanceUserId') as string | undefined;
                  if (ids.length === 1) form.setFieldValue('defaultFinanceUserId', ids[0]);
                  else if (current && !ids.includes(current)) form.setFieldValue('defaultFinanceUserId', undefined);
                }}
              />
            </Form.Item>
            <Form.Item name="defaultFinanceUserId" label="ค่าเริ่มต้น" extra="ถูกเลือกไว้ให้ตอนยืนยันใบจอง เปลี่ยนได้ทุกครั้ง">
              <Select
                allowClear
                placeholder={selected.length ? 'ไม่มีค่าเริ่มต้น' : 'เลือกฝ่ายการเงินด้านบนก่อน'}
                disabled={selected.length === 0}
                options={selected.map((id) => ({ value: id, label: nameOf(id) }))}
              />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={saving}>
              บันทึก
            </Button>
          </Form>
        )}
      </Card>
    </>
  );
}
