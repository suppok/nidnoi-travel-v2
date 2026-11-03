import { CloudSyncOutlined } from '@ant-design/icons';
import { App, Button, type ButtonProps } from 'antd';
import { useEffect, useState } from 'react';
import { siteApi } from '../../../api/site.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { errorMessage } from '../../../lib/errors.ts';

const COOLDOWN_SECONDS = 60;

// Fired after a successful manual rebuild so open views (the build history) can refresh.
export const SITE_REBUILT_EVENT = 'nidnoi:site-rebuilt';

// Shared by every instance (header + history page) and kept across route changes, not reloads.
let cooldownUntil = 0;

/** "อัปเดตหน้าเว็บตอนนี้": rebuilds the customer site now, skipping the 3-minute wait. */
export default function SiteRebuildButton(props: Omit<ButtonProps, 'onClick' | 'disabled' | 'loading'>) {
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const [now, setNow] = useState(() => Date.now());
  const remaining = Math.max(0, Math.ceil((cooldownUntil - now) / 1000));

  useEffect(() => {
    if (remaining === 0) return;
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [remaining]);

  // Another instance may start the cooldown; pick it up.
  useEffect(() => {
    const sync = () => setNow(Date.now());
    window.addEventListener(SITE_REBUILT_EVENT, sync);
    return () => window.removeEventListener(SITE_REBUILT_EVENT, sync);
  }, []);

  if (!can(Permissions.SiteRebuild)) return null;

  const confirm = () =>
    modal.confirm({
      title: 'อัปเดตหน้าเว็บตอนนี้?',
      content: 'ระบบจะสร้างหน้าเว็บลูกค้าใหม่ทันทีโดยไม่ต้องรอ 3 นาที ใช้เวลาประมาณ 2–5 นาทีจึงจะเห็นผลบนเว็บ',
      okText: 'อัปเดตเลย',
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await siteApi.rebuild();
          cooldownUntil = Date.now() + COOLDOWN_SECONDS * 1000;
          window.dispatchEvent(new Event(SITE_REBUILT_EVENT));
          message.success('สั่งอัปเดตหน้าเว็บแล้ว จะเห็นผลภายในไม่กี่นาที');
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });

  return (
    <Button icon={<CloudSyncOutlined />} onClick={confirm} disabled={remaining > 0} {...props}>
      {remaining > 0 ? `สั่งอัปเดตแล้ว (รอ ${remaining} วิ)` : 'อัปเดตหน้าเว็บตอนนี้'}
    </Button>
  );
}
