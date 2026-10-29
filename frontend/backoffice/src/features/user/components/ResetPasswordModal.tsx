import { App, Form, Modal, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { usersApi } from '../../../api/users.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import { generateTemporaryPassword, passwordProblem } from '../../../lib/password.ts';
import type { User } from '../../../types/user.ts';
import TemporaryPasswordInput from './TemporaryPasswordInput.tsx';
import TemporaryPasswordNotice from './TemporaryPasswordNotice.tsx';

type Props = {
  user: User | null; // null = closed
  onClose: () => void;
  onSaved: () => void;
};

export default function ResetPasswordModal({ user, onClose, onSaved }: Props) {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<{ temporaryPassword: string }>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (user) form.setFieldsValue({ temporaryPassword: generateTemporaryPassword() });
  }, [user, form]);

  const onFinish = async ({ temporaryPassword }: { temporaryPassword: string }) => {
    if (!user) return;
    setSubmitting(true);
    try {
      await usersApi.resetPassword({ id: user.id, temporaryPassword });
      modal.success({
        title: 'รีเซ็ตรหัสผ่านแล้ว',
        content: <TemporaryPasswordNotice email={user.email} password={temporaryPassword} />,
        okText: 'เสร็จสิ้น',
        width: 480,
      });
      onSaved();
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={user !== null}
      title="รีเซ็ตรหัสผ่าน"
      okText="รีเซ็ตรหัสผ่าน"
      okButtonProps={{ danger: true }}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
    >
      <Typography.Paragraph className="pt-2">
        ตั้งรหัสผ่านชั่วคราวให้ <strong>{user?.name}</strong> ({user?.email}) ผู้ใช้จะถูกออกจากระบบทุกเครื่อง
        และต้องตั้งรหัสผ่านใหม่เมื่อเข้าสู่ระบบครั้งถัดไป
      </Typography.Paragraph>
      <Form form={form} layout="vertical" requiredMark={false} onFinish={onFinish}>
        <Form.Item
          label="รหัสผ่านชั่วคราว"
          name="temporaryPassword"
          rules={[
            {
              validator: (_, value: string | undefined) => {
                const problem = passwordProblem(value ?? '');
                return problem ? Promise.reject(new Error(problem)) : Promise.resolve();
              },
            },
          ]}
        >
          <TemporaryPasswordInput />
        </Form.Item>
      </Form>
    </Modal>
  );
}
