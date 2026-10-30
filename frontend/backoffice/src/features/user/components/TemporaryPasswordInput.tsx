import { CopyOutlined, ReloadOutlined } from '@ant-design/icons';
import { App, Button, Input, Space, Tooltip } from 'antd';
import { generateTemporaryPassword } from '../../../lib/password.ts';

type Props = {
  value?: string;
  onChange?: (value: string) => void;
};

// Plain-text input (the admin has to hand this password over) with "copy" and "generate" buttons.
export default function TemporaryPasswordInput({ value, onChange }: Props) {
  const { message } = App.useApp();

  const copy = async () => {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      message.success('คัดลอกรหัสผ่านแล้ว');
    } catch {
      message.error('คัดลอกไม่สำเร็จ');
    }
  };

  return (
    <Space.Compact block>
      <Input value={value} onChange={(e) => onChange?.(e.target.value)} autoComplete="off" spellCheck={false} className="font-mono" />
      <Tooltip title="คัดลอกรหัสผ่าน">
        <Button icon={<CopyOutlined />} onClick={copy} disabled={!value} aria-label="คัดลอกรหัสผ่าน" />
      </Tooltip>
      <Tooltip title="สุ่มรหัสผ่านใหม่">
        <Button icon={<ReloadOutlined />} onClick={() => onChange?.(generateTemporaryPassword())} aria-label="สุ่มรหัสผ่านใหม่" />
      </Tooltip>
    </Space.Compact>
  );
}
