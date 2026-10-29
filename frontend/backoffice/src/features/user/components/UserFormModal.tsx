import { App, Checkbox, Form, Input, Modal } from 'antd';
import { useEffect, useState } from 'react';
import { usersApi } from '../../../api/users.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { ASSIGNABLE_ROLES, canGrantRole, isSuperAdmin, type Role, ROLE_INFO } from '../../../constants/roles.ts';
import { applyFieldErrors, errorMessage } from '../../../lib/errors.ts';
import { generateTemporaryPassword, passwordProblem } from '../../../lib/password.ts';
import type { User } from '../../../types/user.ts';
import TemporaryPasswordInput from './TemporaryPasswordInput.tsx';
import TemporaryPasswordNotice from './TemporaryPasswordNotice.tsx';

type UserForm = {
  name: string;
  email: string;
  roles: Role[];
  temporaryPassword: string;
};

type Props = {
  open: boolean;
  user: User | null; // null = create
  isSelf: boolean;
  onClose: () => void;
  onSaved: () => void;
};

export default function UserFormModal({ open, user, isSelf, onClose, onSaved }: Props) {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<UserForm>();
  const [submitting, setSubmitting] = useState(false);
  const creating = user === null;
  const { user: me } = useAuth();
  const myRoles = me?.roles ?? [];
  // super_admin is shown only on the super admin's own profile, where roles are locked anyway.
  const roleOptions = user && isSuperAdmin(user.roles) ? ROLE_INFO : ASSIGNABLE_ROLES;

  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue(
      user
        ? { name: user.name, email: user.email, roles: user.roles }
        : { name: '', email: '', roles: [], temporaryPassword: generateTemporaryPassword() },
    );
  }, [open, user, form]);

  const onFinish = async (values: UserForm) => {
    setSubmitting(true);
    const profile = { name: values.name.trim(), email: values.email.trim().toLowerCase(), roles: values.roles };
    try {
      if (creating) {
        await usersApi.create({ ...profile, temporaryPassword: values.temporaryPassword });
        modal.success({
          title: 'สร้างผู้ใช้แล้ว',
          content: <TemporaryPasswordNotice email={profile.email} password={values.temporaryPassword} />,
          okText: 'เสร็จสิ้น',
          width: 480,
        });
      } else {
        await usersApi.update({ id: user.id, ...profile });
        message.success('บันทึกแล้ว');
      }
      onSaved();
    } catch (err) {
      if (!applyFieldErrors(form, err)) message.error(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={open}
      title={creating ? 'เพิ่มผู้ใช้' : 'แก้ไขผู้ใช้'}
      okText={creating ? 'สร้างผู้ใช้' : 'บันทึก'}
      cancelText="ยกเลิก"
      onOk={() => form.submit()}
      onCancel={onClose}
      confirmLoading={submitting}
      destroyOnHidden
      width={560}
    >
      <Form<UserForm> form={form} layout="vertical" requiredMark={false} onFinish={onFinish} className="pt-2">
        <Form.Item label="ชื่อที่แสดง" name="name" rules={[{ required: true, whitespace: true, message: 'กรุณากรอกชื่อ' }, { max: 100, message: 'ยาวได้ไม่เกิน 100 ตัวอักษร' }]}>
          <Input autoComplete="off" />
        </Form.Item>
        <Form.Item
          label="อีเมล (ใช้เข้าสู่ระบบ)"
          name="email"
          rules={[
            { required: true, message: 'กรุณากรอกอีเมล' },
            { type: 'email', message: 'รูปแบบอีเมลไม่ถูกต้อง' },
          ]}
        >
          <Input type="email" autoComplete="off" />
        </Form.Item>
        <Form.Item
          label="บทบาท"
          name="roles"
          extra={isSelf ? 'เปลี่ยนบทบาทของตัวเองไม่ได้ ให้ผู้ดูแลระบบคนอื่นเปลี่ยนให้' : 'เลือกได้มากกว่า 1 บทบาท สิทธิ์จะรวมกันจากทุกบทบาท'}
          rules={[{ required: true, type: 'array', min: 1, message: 'ต้องเลือกอย่างน้อย 1 บทบาท' }]}
        >
          <Checkbox.Group className="flex flex-col gap-2" disabled={isSelf}>
            {roleOptions.map((role) => (
              <Checkbox key={role.value} value={role.value} disabled={!isSelf && !canGrantRole(myRoles, role.value)}>
                <span className="flex flex-col">
                  <span className="font-medium">{role.label}</span>
                  <span className="text-[13px] text-ink-muted">
                    {!isSelf && !canGrantRole(myRoles, role.value) ? 'เฉพาะผู้ดูแลระบบสูงสุดเท่านั้นที่ให้บทบาทนี้ได้' : role.description}
                  </span>
                </span>
              </Checkbox>
            ))}
          </Checkbox.Group>
        </Form.Item>
        {creating && (
          <Form.Item
            label="รหัสผ่านชั่วคราว"
            name="temporaryPassword"
            extra="ผู้ใช้ต้องเปลี่ยนรหัสผ่านเมื่อเข้าสู่ระบบครั้งแรก"
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
        )}
      </Form>
    </Modal>
  );
}
