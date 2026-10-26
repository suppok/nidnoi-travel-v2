import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { App, Button, Card, Form, Input, Typography } from 'antd';
import { env } from '../../config/env.ts';

type LoginForm = {
  username: string;
  password: string;
};

export default function LoginPage() {
  const { message } = App.useApp();

  // TODO: call POST /auth/login through src/api/auth.ts once the API client exists.
  const onFinish = (_values: LoginForm) => {
    message.info('ยังไม่ได้เชื่อมต่อระบบเข้าสู่ระบบ');
  };

  return (
    <Card className="w-full max-w-sm shadow-sm">
      <Typography.Title level={3} className="!mb-1 text-center">
        {env.APP_NAME}
      </Typography.Title>
      <Typography.Paragraph type="secondary" className="text-center">
        เข้าสู่ระบบสำหรับพนักงาน
      </Typography.Paragraph>

      <Form<LoginForm> layout="vertical" requiredMark={false} onFinish={onFinish} autoComplete="on">
        <Form.Item
          label="ชื่อผู้ใช้"
          name="username"
          rules={[{ required: true, message: 'กรุณากรอกชื่อผู้ใช้' }]}
        >
          <Input prefix={<UserOutlined />} autoComplete="username" autoFocus />
        </Form.Item>
        <Form.Item
          label="รหัสผ่าน"
          name="password"
          rules={[{ required: true, message: 'กรุณากรอกรหัสผ่าน' }]}
        >
          <Input.Password prefix={<LockOutlined />} autoComplete="current-password" />
        </Form.Item>
        <Button type="primary" htmlType="submit" block>
          เข้าสู่ระบบ
        </Button>
      </Form>
    </Card>
  );
}
