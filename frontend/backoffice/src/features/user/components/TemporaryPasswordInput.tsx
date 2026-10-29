import { ReloadOutlined } from '@ant-design/icons';
import { Button, Input, Space, Tooltip } from 'antd';
import { generateTemporaryPassword } from '../../../lib/password.ts';

type Props = {
  value?: string;
  onChange?: (value: string) => void;
};

// Plain-text input (the admin has to hand this password over) with a "generate" button.
export default function TemporaryPasswordInput({ value, onChange }: Props) {
  return (
    <Space.Compact block>
      <Input value={value} onChange={(e) => onChange?.(e.target.value)} autoComplete="off" spellCheck={false} className="font-mono" />
      <Tooltip title="สุ่มรหัสผ่านใหม่">
        <Button icon={<ReloadOutlined />} onClick={() => onChange?.(generateTemporaryPassword())} aria-label="สุ่มรหัสผ่านใหม่" />
      </Tooltip>
    </Space.Compact>
  );
}
