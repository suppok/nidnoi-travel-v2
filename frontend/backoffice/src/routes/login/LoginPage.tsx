import { ArrowRightOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { Alert, App, Button, Checkbox, Form, Input, Spin } from 'antd';
import { useState } from 'react';
import { type Location, Navigate, useLocation, useNavigate } from 'react-router';
import { MOCK_PASSWORD } from '../../api/mock/credentials.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { PlaneIcon } from '../../components/BrandMark.tsx';
import { env } from '../../config/env.ts';
import { applyFieldErrors, errorMessage } from '../../lib/errors.ts';

type LoginForm = {
  email: string;
  password: string;
  rememberMe: boolean;
};

function StubField({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs font-semibold tracking-[0.08em] text-brand-dark">{label}</span>
      <span className="text-base font-medium">{value}</span>
    </div>
  );
}

// Decorative "boarding pass" stub from the design example.
function BoardingPassStub() {
  return (
    <div className="relative flex flex-[1_1_320px] flex-col gap-7 border-b-2 border-dashed border-[#F2B8CE] bg-brand-soft px-9 py-10 md:border-r-2 md:border-b-0">
      <div aria-hidden="true" className="absolute -top-4 -right-[17px] hidden h-8 w-8 rounded-full bg-sky md:block" />
      <div aria-hidden="true" className="absolute -right-[17px] -bottom-4 hidden h-8 w-8 rounded-full bg-sky md:block" />

      <div className="flex items-center justify-between gap-3">
        <span className="text-[13px] font-semibold tracking-[0.16em] text-brand-dark">BOARDING PASS</span>
        <span className="rounded-full bg-white px-2.5 py-1 text-[13px] font-semibold text-brand-dark">STAFF ONLY</span>
      </div>

      <div className="flex items-center justify-between gap-3">
        <div className="flex flex-col gap-0.5">
          <span className="font-display text-[44px] leading-none font-medium">BKK</span>
          <span className="text-sm text-ink-muted">กรุงเทพฯ</span>
        </div>
        <div aria-hidden="true" className="flex min-w-[60px] flex-auto items-center gap-1.5 text-brand">
          <div className="flex-auto border-t-2 border-dashed border-brand-line" />
          <PlaneIcon />
          <div className="flex-auto border-t-2 border-dashed border-brand-line" />
        </div>
        <div className="flex flex-col items-end gap-0.5">
          <span className="font-display text-[44px] leading-none font-medium">NDN</span>
          <span className="text-sm text-ink-muted">หลังบ้าน</span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-x-4 gap-y-[18px]">
        <StubField label="ผู้โดยสาร" value="ทีมงาน nidnoi" />
        <StubField label="ประตู" value="Backoffice" />
        <StubField label="ปลายทาง" value="ทัวร์ · รอบเดินทาง · ราคา" />
        <StubField label="ที่นั่ง" value="STAFF" />
      </div>

      <div
        aria-hidden="true"
        className="mt-auto h-[52px] rounded-md bg-white bg-clip-content px-2.5 py-2 opacity-85"
        style={{
          backgroundImage:
            'repeating-linear-gradient(90deg, #1D2B4F 0 2px, transparent 2px 5px, #1D2B4F 5px 6px, transparent 6px 9px, #1D2B4F 9px 12px, transparent 12px 14px)',
        }}
      />
    </div>
  );
}

export default function LoginPage() {
  const { modal } = App.useApp();
  const { status, user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm<LoginForm>();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const from = (location.state as { from?: Location } | null)?.from;
  const target = from ? `${from.pathname}${from.search}` : '/dashboard';

  if (status === 'loading') return <Spin size="large" />;
  if (user) return <Navigate to={target} replace />;

  const showContactAdmin = () =>
    modal.info({
      title: 'ติดต่อผู้ดูแลระบบ',
      content: 'หากลืมรหัสผ่านหรือเข้าสู่ระบบไม่ได้ ให้ผู้ดูแลระบบรีเซ็ตรหัสผ่านชั่วคราวให้ที่เมนู "ผู้ใช้และสิทธิ์" แล้วเปลี่ยนรหัสผ่านใหม่หลังเข้าสู่ระบบ',
      okText: 'เข้าใจแล้ว',
    });

  const onFinish = async (values: LoginForm) => {
    setSubmitting(true);
    setError(null);
    try {
      await login({ email: values.email.trim(), password: values.password, rememberMe: values.rememberMe });
      navigate(target, { replace: true });
    } catch (err) {
      if (!applyFieldErrors(form, err)) setError(errorMessage(err));
      form.setFieldValue('password', '');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex w-full max-w-[900px] flex-wrap overflow-hidden rounded-[28px] bg-white shadow-[0_30px_60px_-30px_rgba(29,43,79,0.28),0_2px_6px_rgba(29,43,79,0.05)]">
      <BoardingPassStub />

      <div className="flex flex-[999_1_380px] flex-col gap-5 px-12 py-11 max-md:px-6">
        <div className="flex flex-col gap-2">
          <h1 className="m-0 font-display text-[32px] font-medium">เข้าสู่ระบบ</h1>
          <p className="m-0 text-[15px] leading-relaxed text-ink-muted">
            ยินดีต้อนรับกลับ ลงชื่อเข้าใช้เพื่อจัดการทัวร์ รอบเดินทาง และการจองของ nidnoi travel
          </p>
        </div>

        {env.API_MOCK && (
          <Alert
            type="info"
            showIcon
            title="โหมดจำลอง (VITE_API_MOCK)"
            description={`ทดลองด้วย admin@ / sales@ / finance@ / new@nidnoi.test รหัสผ่าน ${MOCK_PASSWORD}`}
          />
        )}
        {error && <Alert type="error" showIcon title={error} />}

        <Form<LoginForm>
          form={form}
          layout="vertical"
          size="large"
          requiredMark={false}
          initialValues={{ rememberMe: true }}
          onFinish={onFinish}
          onValuesChange={() => setError(null)}
        >
          <Form.Item
            label="อีเมล"
            name="email"
            rules={[
              { required: true, message: 'กรุณากรอกอีเมล' },
              { type: 'email', message: 'รูปแบบอีเมลไม่ถูกต้อง' },
            ]}
          >
            <Input prefix={<MailOutlined className="text-ink-muted" />} type="email" autoComplete="username" placeholder="you@nidnoitravel.com" autoFocus />
          </Form.Item>
          <Form.Item
            label={
              <span className="flex w-full items-baseline justify-between gap-3">
                <span>รหัสผ่าน</span>
                <Button type="link" size="small" className="px-0" onClick={showContactAdmin}>
                  ลืมรหัสผ่าน?
                </Button>
              </span>
            }
            name="password"
            rules={[{ required: true, message: 'กรุณากรอกรหัสผ่าน' }]}
            className="[&_.ant-form-item-label>label]:w-full"
          >
            <Input.Password prefix={<LockOutlined className="text-ink-muted" />} autoComplete="current-password" placeholder="กรอกรหัสผ่าน" />
          </Form.Item>
          <Form.Item name="rememberMe" valuePropName="checked" extra="ไม่ต้องเข้าสู่ระบบใหม่เมื่อปิดเบราว์เซอร์ อย่าเลือกบนเครื่องที่ใช้ร่วมกัน">
            <Checkbox>จดจำฉันไว้ในเครื่องนี้</Checkbox>
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={submitting} className="h-[52px] font-semibold" icon={<ArrowRightOutlined />} iconPlacement="end">
            เข้าสู่ระบบ
          </Button>
        </Form>

        <p className="m-0 text-center text-sm leading-relaxed text-ink-muted">
          มีปัญหาในการเข้าสู่ระบบ?{' '}
          <Button type="link" size="small" className="px-0 font-medium" onClick={showContactAdmin}>
            ติดต่อผู้ดูแลระบบ
          </Button>
        </p>
      </div>
    </div>
  );
}
