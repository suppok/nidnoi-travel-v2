import { LockOutlined } from '@ant-design/icons';
import { Alert, App, Button, Form, Input } from 'antd';
import { useState } from 'react';
import { useNavigate } from 'react-router';
import { authApi } from '../../api/auth.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { applyFieldErrors, errorMessage } from '../../lib/errors.ts';
import { PASSWORD_MIN_LENGTH, passwordProblem } from '../../lib/password.ts';

type ChangePasswordForm = {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
};

// Forced after login with a temporary password (new user or admin reset); also reachable from the user menu.
export default function ChangePasswordPage() {
  const { message } = App.useApp();
  const { user, reloadUser, logout } = useAuth();
  const navigate = useNavigate();
  const [form] = Form.useForm<ChangePasswordForm>();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const forced = user?.mustChangePassword ?? false;

  const onFinish = async (values: ChangePasswordForm) => {
    setSubmitting(true);
    setError(null);
    try {
      await authApi.changePassword({ currentPassword: values.currentPassword, newPassword: values.newPassword });
      await reloadUser();
      message.success('เปลี่ยนรหัสผ่านแล้ว');
      navigate('/dashboard', { replace: true });
    } catch (err) {
      if (!applyFieldErrors(form, err)) setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex w-full max-w-[480px] flex-col gap-5 rounded-[28px] bg-white px-10 py-10 shadow-[0_30px_60px_-30px_rgba(29,43,79,0.28),0_2px_6px_rgba(29,43,79,0.05)] max-md:px-6">
      <div className="flex flex-col gap-2">
        <h1 className="m-0 font-display text-[28px] font-medium">{forced ? 'ตั้งรหัสผ่านใหม่' : 'เปลี่ยนรหัสผ่าน'}</h1>
        <p className="m-0 text-[15px] leading-relaxed text-ink-muted">
          {forced
            ? `สวัสดี ${user?.name ?? ''} บัญชีนี้ใช้รหัสผ่านชั่วคราว กรุณาตั้งรหัสผ่านของคุณเองก่อนเริ่มใช้งาน`
            : 'กรอกรหัสผ่านปัจจุบันและรหัสผ่านใหม่'}
        </p>
      </div>

      {error && <Alert type="error" showIcon title={error} />}

      <Form<ChangePasswordForm> form={form} layout="vertical" size="large" requiredMark={false} onFinish={onFinish} onValuesChange={() => setError(null)}>
        {/* Hidden username helps password managers save the new password against the right account. */}
        <input type="email" name="username" autoComplete="username" value={user?.email ?? ''} readOnly hidden />
        <Form.Item
          label={forced ? 'รหัสผ่านชั่วคราว' : 'รหัสผ่านปัจจุบัน'}
          name="currentPassword"
          rules={[{ required: true, message: 'กรุณากรอกรหัสผ่าน' }]}
        >
          <Input.Password prefix={<LockOutlined className="text-ink-muted" />} autoComplete="current-password" autoFocus />
        </Form.Item>
        <Form.Item
          label="รหัสผ่านใหม่"
          name="newPassword"
          extra={`อย่างน้อย ${PASSWORD_MIN_LENGTH} ตัวอักษร มีทั้งตัวอักษรภาษาอังกฤษและตัวเลข`}
          dependencies={['currentPassword']}
          rules={[
            { required: true, message: 'กรุณากรอกรหัสผ่านใหม่' },
            ({ getFieldValue }) => ({
              validator: (_, value: string | undefined) => {
                if (!value) return Promise.resolve();
                const problem = passwordProblem(value);
                if (problem) return Promise.reject(new Error(problem));
                if (value === getFieldValue('currentPassword')) return Promise.reject(new Error('ต้องไม่ซ้ำกับรหัสผ่านเดิม'));
                return Promise.resolve();
              },
            }),
          ]}
        >
          <Input.Password prefix={<LockOutlined className="text-ink-muted" />} autoComplete="new-password" />
        </Form.Item>
        <Form.Item
          label="ยืนยันรหัสผ่านใหม่"
          name="confirmPassword"
          dependencies={['newPassword']}
          rules={[
            { required: true, message: 'กรุณายืนยันรหัสผ่าน' },
            ({ getFieldValue }) => ({
              validator: (_, value: string | undefined) =>
                !value || value === getFieldValue('newPassword') ? Promise.resolve() : Promise.reject(new Error('รหัสผ่านไม่ตรงกัน')),
            }),
          ]}
        >
          <Input.Password prefix={<LockOutlined className="text-ink-muted" />} autoComplete="new-password" />
        </Form.Item>

        <div className="flex flex-col gap-2">
          <Button type="primary" htmlType="submit" block loading={submitting} className="h-[52px] font-semibold">
            บันทึกรหัสผ่านใหม่
          </Button>
          {forced ? (
            <Button type="text" block onClick={() => void logout()}>
              ออกจากระบบ
            </Button>
          ) : (
            <Button type="text" block onClick={() => navigate(-1)}>
              ยกเลิก
            </Button>
          )}
        </div>
      </Form>
    </div>
  );
}
