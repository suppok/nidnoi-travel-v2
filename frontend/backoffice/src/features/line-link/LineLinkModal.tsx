import { CheckCircleFilled } from '@ant-design/icons';
import { App, Button, Modal, Typography } from 'antd';
import { useState } from 'react';
import { useLocation } from 'react-router';
import { authApi } from '../../api/auth.ts';
import { startLineLink } from '../../auth/lineLogin.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';

type Props = {
  open: boolean;
  onClose: () => void;
};

// Links the signed-in user's own LINE account, so they can sign in with "ล็อกอินด้วย LINE" (handy on
// a phone). The browser goes to LINE Login and comes back to LineCallbackPage, which finishes the link.
export default function LineLinkModal({ open, onClose }: Props) {
  const { message, modal } = App.useApp();
  const { user, reloadUser } = useAuth();
  const location = useLocation();
  const [starting, setStarting] = useState(false);
  const linked = user?.line ?? null;

  const link = async () => {
    setStarting(true);
    try {
      const config = await authApi.lineConfig();
      if (!config.enabled) {
        message.error('ยังไม่ได้ตั้งค่าการเชื่อมต่อ LINE กรุณาติดต่อผู้ดูแลระบบ');
        setStarting(false);
        return;
      }
      await startLineLink(config.channelId, `${location.pathname}${location.search}`);
    } catch (err) {
      message.error(errorMessage(err));
      setStarting(false);
    }
  };

  const confirmUnlink = () => {
    modal.confirm({
      title: 'ยกเลิกการเชื่อมบัญชี LINE?',
      content: 'จะเข้าสู่ระบบด้วย LINE ไม่ได้จนกว่าจะเชื่อมใหม่',
      okText: 'ยกเลิกการเชื่อม',
      okButtonProps: { danger: true },
      cancelText: 'ปิด',
      onOk: async () => {
        try {
          await authApi.lineUnlink();
          await reloadUser();
          message.success('ยกเลิกการเชื่อมบัญชี LINE แล้ว');
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  return (
    <Modal open={open} onCancel={onClose} title="บัญชี LINE" footer={null} destroyOnHidden>
      <div className="flex flex-col gap-4">
        <Typography.Paragraph type="secondary" className="mb-0">
          เชื่อมบัญชี LINE ของคุณเพื่อเข้าสู่ระบบด้วยปุ่ม "ล็อกอินด้วย LINE" ได้ทันที สะดวกเมื่อใช้งานบนมือถือ
        </Typography.Paragraph>

        {linked && (
          <div className="flex items-center gap-2">
            <CheckCircleFilled className="text-green-600" />
            <span>
              เชื่อมกับ <strong>{linked.displayName}</strong> แล้ว
              <span className="ml-1 text-xs text-ink-muted">({formatDateTime(linked.linkedAt)})</span>
            </span>
          </div>
        )}

        <div className="flex justify-end gap-2">
          {linked && (
            <Button danger onClick={confirmUnlink}>
              ยกเลิกการเชื่อม
            </Button>
          )}
          <Button type="primary" loading={starting} onClick={() => void link()}>
            {linked ? 'เชื่อมบัญชีอื่นแทน' : 'เชื่อมบัญชี LINE'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
