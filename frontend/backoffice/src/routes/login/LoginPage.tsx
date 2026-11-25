import { ArrowRightOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { Alert, App, Button, Checkbox, Form, Input, Spin } from 'antd';
import { useEffect, useState } from 'react';
import { type Location, Navigate, useLocation, useNavigate } from 'react-router';
import { authApi } from '../../api/auth.ts';
import { MOCK_PASSWORD } from '../../api/mock/credentials.ts';
import { startLineLogin } from '../../auth/lineLogin.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { env } from '../../config/env.ts';
import { applyFieldErrors, errorMessage } from '../../lib/errors.ts';

type LoginForm = {
  email: string;
  password: string;
  rememberMe: boolean;
};

// The LINE speech-bubble mark, white on the LINE green button.
function LineIcon() {
  return (
    <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" aria-hidden="true">
      <path d="M12 3C6.48 3 2 6.6 2 11.04c0 3.98 3.55 7.31 8.35 7.94.33.07.77.22.88.5.1.25.07.65.03.9l-.14.85c-.04.25-.2.98.86.53 1.06-.45 5.73-3.38 7.82-5.78C21.25 14.4 22 12.8 22 11.04 22 6.6 17.52 3 12 3Zm-3.6 10.2H6.44a.52.52 0 0 1-.52-.52V8.76a.52.52 0 1 1 1.04 0v3.4H8.4a.52.52 0 1 1 0 1.04Zm2.03-.52a.52.52 0 1 1-1.04 0V8.76a.52.52 0 1 1 1.04 0v3.92Zm4.72 0a.52.52 0 0 1-.94.31l-2-2.73v2.42a.52.52 0 1 1-1.04 0V8.76a.52.52 0 0 1 .94-.31l2 2.73V8.76a.52.52 0 1 1 1.04 0v3.92Zm3.17-2.48a.52.52 0 1 1 0 1.04h-1.44v.92h1.44a.52.52 0 1 1 0 1.04h-1.96a.52.52 0 0 1-.52-.52V8.76c0-.29.23-.52.52-.52h1.96a.52.52 0 1 1 0 1.04h-1.44v.92h1.44Z" />
    </svg>
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
  // LINE Login shows only when the backend has the LINE Login channel configured.
  const [lineChannelId, setLineChannelId] = useState<string | null>(null);
  const [lineStarting, setLineStarting] = useState(false);

  useEffect(() => {
    let cancelled = false;
    authApi
      .lineConfig()
      .then((c) => !cancelled && setLineChannelId(c.enabled ? c.channelId : null))
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

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

  const loginWithLine = () => {
    if (!lineChannelId) return;
    setLineStarting(true);
    void startLineLogin(lineChannelId, form.getFieldValue('rememberMe') ?? true, target).catch(() => {
      setLineStarting(false);
      setError('เปิดหน้าเข้าสู่ระบบ LINE ไม่สำเร็จ กรุณาลองใหม่');
    });
  };

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
    <div className="flex w-full max-w-[400px] flex-col gap-4 rounded-[20px] bg-white px-8 py-7 shadow-[0_30px_60px_-30px_rgba(29,43,79,0.28),0_2px_6px_rgba(29,43,79,0.05)] max-sm:px-5">
      <h1 className="m-0 font-display text-2xl font-normal">เข้าสู่ระบบ</h1>

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
        requiredMark={false}
        initialValues={{ rememberMe: true }}
        onFinish={onFinish}
        onValuesChange={() => setError(null)}
        className="[&_.ant-form-item]:mb-4"
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
        <Form.Item name="rememberMe" valuePropName="checked">
          <Checkbox>จดจำฉันไว้ในเครื่องนี้</Checkbox>
        </Form.Item>
        <Button type="primary" htmlType="submit" block loading={submitting} className="h-10 font-semibold" icon={<ArrowRightOutlined />} iconPlacement="end">
          เข้าสู่ระบบ
        </Button>
      </Form>

      {lineChannelId && (
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-3 text-xs text-ink-muted">
            <div className="h-px flex-1 bg-line" />
            หรือ
            <div className="h-px flex-1 bg-line" />
          </div>
          <Button
            block
            loading={lineStarting}
            onClick={loginWithLine}
            icon={<LineIcon />}
            className="h-10 border-[#06C755] bg-[#06C755] font-semibold text-white hover:!border-[#05b14c] hover:!bg-[#05b14c] hover:!text-white"
          >
            ล็อกอินด้วย LINE
          </Button>
        </div>
      )}

      <p className="m-0 text-center text-sm leading-relaxed text-ink-muted">
        มีปัญหาในการเข้าสู่ระบบ?{' '}
        <Button type="link" size="small" className="px-0 font-medium" onClick={showContactAdmin}>
          ติดต่อผู้ดูแลระบบ
        </Button>
      </p>
    </div>
  );
}
