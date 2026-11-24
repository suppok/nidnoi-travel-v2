import { ApiOutlined, DeleteOutlined, SaveOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Form, Input, InputNumber, Select, Spin, Switch, Tag, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { aiSettingsApi } from '../../api/aiSettings.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { Permissions } from '../../constants/permissions.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { applyFieldErrors, errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';
import type { AISetting } from '../../types/aiSetting.ts';

type FormValues = { enabled: boolean; model: string; maxPdfSizeMb: number; apiKey: string };

// Claude API key, model and on/off for the AI tour prefill (admins only). The key is never sent back
// by the backend; the page shows only its last 4 characters.
export default function AISettingsPage() {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const canEdit = can(Permissions.AISettingUpdate);
  const [form] = Form.useForm<FormValues>();
  const [setting, setSetting] = useState<AISetting | null>(null);
  const [loadError, setLoadError] = useState('');
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ ok: boolean; text: string } | null>(null);

  const show = (s: AISetting) => {
    setSetting(s);
    form.setFieldsValue({ enabled: s.enabled, model: s.model, maxPdfSizeMb: s.maxPdfSizeMb, apiKey: '' });
  };

  useEffect(() => {
    aiSettingsApi
      .get()
      .then(show)
      .catch((err: unknown) => setLoadError(errorMessage(err)));
    // eslint-disable-next-line react-hooks/exhaustive-deps -- load once
  }, []);

  const save = async (values: FormValues, clearApiKey = false) => {
    setSaving(true);
    try {
      const apiKey = values.apiKey?.trim() || null;
      show(await aiSettingsApi.update({ enabled: values.enabled, model: values.model, maxPdfSizeMb: values.maxPdfSizeMb, apiKey, clearApiKey }));
      setTestResult(null);
      message.success(clearApiKey ? 'ลบ API key แล้ว' : 'บันทึกการตั้งค่า AI แล้ว');
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const test = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const { apiKey, model } = form.getFieldsValue();
      const res = await aiSettingsApi.test({ apiKey: apiKey?.trim() ?? '', model });
      setTestResult({ ok: true, text: `เชื่อมต่อสำเร็จ — ใช้โมเดล ${res.model} ได้` });
    } catch (err) {
      if (!applyFieldErrors(form, err)) setTestResult({ ok: false, text: errorMessage(err) });
    } finally {
      setTesting(false);
    }
  };

  const clearKey = () =>
    modal.confirm({
      title: 'ลบ API key?',
      content: 'AI จะถูกปิดจนกว่าจะใส่ API key ใหม่',
      okText: 'ลบ API key',
      okButtonProps: { danger: true },
      cancelText: 'ยกเลิก',
      onOk: () => save({ ...form.getFieldsValue(), enabled: false, apiKey: '' }, true),
    });

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="ตั้งค่า AI" subtitle="ใช้ Claude อ่านไฟล์ PDF โปรแกรมทัวร์แล้วกรอกฟอร์มสร้างโปรแกรมให้ — พนักงานตรวจและบันทึกเอง" />
      {loadError && <Alert type="error" showIcon message={loadError} />}
      {!setting && !loadError && <Spin />}

      {setting && (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <Card
            title={
              <span className="flex items-center gap-2">
                การเชื่อมต่อ Claude {setting.ready ? <Tag color="green">พร้อมใช้งาน</Tag> : <Tag>ปิดอยู่</Tag>}
              </span>
            }
          >
            {!setting.encryptionConfigured && (
              <Alert
                className="mb-4"
                type="warning"
                showIcon
                message="เซิร์ฟเวอร์ยังไม่ได้ตั้งค่า AI_SETTINGS_ENCRYPTION_KEY"
                description="บันทึก API key ไม่ได้จนกว่าผู้ดูแลเซิร์ฟเวอร์จะตั้งค่าตัวแปรนี้"
              />
            )}
            {setting.keyUnreadable && (
              <Alert className="mb-4" type="warning" showIcon message="ถอดรหัส API key ที่บันทึกไว้ไม่ได้ (คีย์เข้ารหัสบนเซิร์ฟเวอร์เปลี่ยน) กรุณาใส่ API key ใหม่" />
            )}
            <Form form={form} layout="vertical" disabled={!canEdit || saving} onFinish={(v) => save(v)} requiredMark={false}>
              <Form.Item label="เปิดใช้ AI กรอกโปรแกรมจาก PDF" name="enabled" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item
                label="Anthropic API key"
                name="apiKey"
                extra={
                  setting.hasApiKey
                    ? `บันทึกไว้แล้ว (ลงท้าย …${setting.apiKeyLast4}) — เว้นว่างไว้เพื่อใช้คีย์เดิม`
                    : 'สร้างได้ที่ Claude Console (platform.claude.com) → API keys'
                }
              >
                <Input.Password placeholder={setting.hasApiKey ? `•••• ${setting.apiKeyLast4}` : 'sk-ant-…'} autoComplete="off" />
              </Form.Item>
              <Form.Item label="โมเดล" name="model" rules={[{ required: true, message: 'กรุณาเลือกโมเดล' }]}>
                <Select options={setting.models.map((m) => ({ value: m.id, label: m.label }))} className="max-w-[420px]" />
              </Form.Item>
              <Form.Item
                label="ขนาดไฟล์ PDF สูงสุด (MB)"
                name="maxPdfSizeMb"
                rules={[{ required: true, message: 'กรุณากรอกขนาดไฟล์' }]}
                extra={`ไม่เกิน ${setting.maxPdfSizeMbLimit} MB`}
              >
                <InputNumber min={1} max={setting.maxPdfSizeMbLimit} precision={0} className="w-[140px]" />
              </Form.Item>

              {testResult && <Alert className="mb-4" type={testResult.ok ? 'success' : 'error'} showIcon message={testResult.text} />}

              <div className="flex flex-wrap gap-2">
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
                  บันทึก
                </Button>
                <Button icon={<ApiOutlined />} onClick={test} loading={testing} disabled={!canEdit || saving}>
                  ทดสอบการเชื่อมต่อ
                </Button>
                {setting.hasApiKey && (
                  <Button danger icon={<DeleteOutlined />} onClick={clearKey} disabled={!canEdit || saving}>
                    ลบ API key
                  </Button>
                )}
              </div>
            </Form>
            {setting.updateDate && (
              <p className="mb-0 mt-4 text-[13px] text-ink-muted">
                แก้ไขล่าสุด {formatDateTime(setting.updateDate)}
                {setting.updateBy && ` โดย ${setting.updateBy}`}
              </p>
            )}
          </Card>

          <Card title="หมายเหตุ" size="small">
            <Typography.Paragraph className="text-[13px]">
              API key ถูกเข้ารหัสก่อนเก็บในฐานข้อมูล และไม่ถูกส่งกลับมาที่หน้าเว็บ
            </Typography.Paragraph>
            <Typography.Paragraph className="text-[13px]">
              "ทดสอบการเชื่อมต่อ" ใช้คีย์ที่พิมพ์ไว้ (หรือคีย์ที่บันทึกไว้ถ้าเว้นว่าง) ตรวจกับ Anthropic โดยไม่เสียค่าใช้จ่าย และยังไม่บันทึก
            </Typography.Paragraph>
            <Typography.Paragraph className="mb-0 text-[13px]">
              ค่าใช้จ่ายคิดตามการใช้งานจริงในบัญชี Claude Console — ไฟล์ PDF 20–30 หน้าประมาณครั้งละ 7–10 บาทกับ Sonnet (Opus ประมาณสองเท่า)
            </Typography.Paragraph>
          </Card>
        </div>
      )}
    </div>
  );
}
