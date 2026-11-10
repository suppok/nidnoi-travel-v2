import { CheckCircleFilled } from '@ant-design/icons';
import { Alert, App, Button, Modal, QRCode, Typography } from 'antd';
import { useCallback, useEffect, useState } from 'react';
import { authApi } from '../../api/auth.ts';
import { useAuth } from '../../auth/useAuth.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';
import type { LineLinkResponse } from '../../types/auth.ts';

type Props = {
  open: boolean;
  onClose: () => void;
};

// While a QR code is shown, check every few seconds whether the phone finished the link.
const POLL_MS = 3000;

function remaining(expiresAt: string, now: number): string {
  const s = Math.max(0, Math.round((new Date(expiresAt).getTime() - now) / 1000));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
}

// Links the signed-in user's own LINE account, so the LIFF booking app inside LINE knows who they
// are. The backend issues a one-time LIFF URL (10 minutes); opening it in LINE finishes the link.
export default function LineLinkModal({ open, onClose }: Props) {
  const { message, modal } = App.useApp();
  const { user, reloadUser } = useAuth();
  const [link, setLink] = useState<LineLinkResponse | null>(null);
  const [linkedBefore, setLinkedBefore] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [now, setNow] = useState(() => Date.now());
  const linked = user?.line ?? null;
  const linkedAt = linked?.linkedAt ?? null;
  const expired = link !== null && new Date(link.expiresAt).getTime() <= now;

  const createLink = useCallback(async () => {
    setCreating(true);
    try {
      setLinkedBefore(linkedAt);
      setLink(await authApi.createLineLink());
      setNow(Date.now());
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setCreating(false);
    }
  }, [message, linkedAt]);

  const close = () => {
    setLink(null);
    onClose();
  };

  // Countdown and polling while the QR code is valid.
  useEffect(() => {
    if (!open || !link || expired) return;
    const tick = setInterval(() => setNow(Date.now()), 1000);
    const poll = setInterval(() => {
      authApi
        .me()
        .then((me) => {
          if (me.line && me.line.linkedAt !== linkedBefore) {
            setLink(null);
            void reloadUser();
            message.success(`เชื่อมบัญชี LINE "${me.line.displayName}" แล้ว`);
          }
        })
        .catch(() => undefined); // a failed poll is retried on the next tick
    }, POLL_MS);
    return () => {
      clearInterval(tick);
      clearInterval(poll);
    };
  }, [open, link, expired, linkedBefore, reloadUser, message]);

  const confirmUnlink = () => {
    modal.confirm({
      title: 'ยกเลิกการเชื่อมบัญชี LINE?',
      content: 'จะส่งใบจองจากแอปใน LINE ไม่ได้จนกว่าจะเชื่อมใหม่',
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
    <Modal open={open} onCancel={close} title="บัญชี LINE" footer={null} destroyOnHidden>
      <div className="flex flex-col gap-4">
        <Typography.Paragraph type="secondary" className="mb-0">
          เชื่อมบัญชี LINE ของคุณเพื่อส่งใบจองผ่านแอปใน LINE OA (เฉพาะฝ่ายขายและผู้ดูแลระบบ)
        </Typography.Paragraph>

        {linked && !link && (
          <div className="flex items-center gap-2">
            <CheckCircleFilled className="text-green-600" />
            <span>
              เชื่อมกับ <strong>{linked.displayName}</strong> แล้ว
              <span className="ml-1 text-xs text-ink-muted">({formatDateTime(linked.linkedAt)})</span>
            </span>
          </div>
        )}

        {link ? (
          <div className="flex flex-col items-center gap-3">
            <QRCode value={link.url} size={220} status={expired ? 'expired' : 'active'} onRefresh={() => void createLink()} />
            {expired ? (
              <Typography.Text type="secondary">QR code หมดอายุแล้ว กด “refresh” เพื่อสร้างใหม่</Typography.Text>
            ) : (
              <>
                <Typography.Text>สแกนด้วยกล้องมือถือหรือแอป LINE · หมดอายุใน {remaining(link.expiresAt, now)}</Typography.Text>
                <Typography.Text type="secondary" className="text-center text-xs">
                  ใช้ LINE บนคอมพิวเตอร์? เปิดลิงก์นี้ในเบราว์เซอร์แล้วเข้าสู่ระบบ LINE{' '}
                  <Typography.Text copyable={{ text: link.url, tooltips: ['คัดลอกลิงก์', 'คัดลอกแล้ว'] }} />
                </Typography.Text>
              </>
            )}
            <Alert type="info" showIcon title="ห้ามส่ง QR code หรือลิงก์นี้ให้ผู้อื่น — ใครเปิดก่อนจะได้เชื่อมกับบัญชีของคุณ" />
          </div>
        ) : (
          <div className="flex justify-end gap-2">
            {linked && (
              <Button danger onClick={confirmUnlink}>
                ยกเลิกการเชื่อม
              </Button>
            )}
            <Button type="primary" loading={creating} onClick={() => void createLink()}>
              {linked ? 'เชื่อมบัญชีอื่นแทน' : 'เชื่อมบัญชี LINE'}
            </Button>
          </div>
        )}
      </div>
    </Modal>
  );
}
