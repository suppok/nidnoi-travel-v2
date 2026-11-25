import { RollbackOutlined, SaveOutlined } from '@ant-design/icons';
import { Alert, App, Button, Card, Image, Skeleton, Tag, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { appConfigApi } from '../../api/appConfig.ts';
import PageHeader from '../../components/PageHeader.tsx';
import ImageUploadField from '../../features/tour-program/components/ImageUploadField.tsx';
import { errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';
import type { EmailContactCard } from '../../types/appConfig.ts';

// Width the card is shown at in emails (backend template: 504 px).
const EMAIL_WIDTH = 504;

// The contact card at the bottom of emails sent to wholesales (admins only).
export default function EmailContactCardPage() {
  const { message, modal } = App.useApp();
  const [card, setCard] = useState<EmailContactCard | null>(null);
  const [loadError, setLoadError] = useState('');
  const [draft, setDraft] = useState<string | null>(null); // a newly uploaded image, not saved yet
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    appConfigApi
      .getEmailContactCard()
      .then(setCard)
      .catch((err: unknown) => setLoadError(errorMessage(err)));
  }, []);

  const save = async (imageUrl: string, done: string) => {
    setSaving(true);
    try {
      setCard(await appConfigApi.updateEmailContactCard({ imageUrl }));
      setDraft(null);
      message.success(done);
    } catch (err) {
      message.error(errorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const resetToDefault = () =>
    modal.confirm({
      title: 'กลับไปใช้รูปเริ่มต้น?',
      content: 'อีเมลที่ส่งหลังจากนี้จะใช้รูปติดต่อเริ่มต้นของระบบ',
      okText: 'ใช้รูปเริ่มต้น',
      cancelText: 'ยกเลิก',
      onOk: () => save('', 'กลับไปใช้รูปเริ่มต้นแล้ว'),
    });

  const current = card?.imageUrl || card?.defaultImage || '';

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="รูปติดต่อท้ายอีเมล" subtitle="รูปที่แสดงท้ายอีเมลยืนยันการจองที่ส่งถึงโฮลเซลล์ ใต้เส้น --" />
      {loadError && <Alert type="error" showIcon message={loadError} />}
      {!card && !loadError && <Skeleton active paragraph={{ rows: 4 }} />}

      {card && (
        <div className="grid gap-6 lg:grid-cols-2">
          <Card
            title={
              <span className="flex items-center gap-2">
                รูปที่ใช้อยู่ {card.imageUrl ? <Tag color="blue">อัปโหลดเอง</Tag> : <Tag>รูปเริ่มต้น</Tag>}
              </span>
            }
          >
            {current ? (
              <Image src={current} alt="รูปติดต่อท้ายอีเมล" style={{ width: '100%', maxWidth: EMAIL_WIDTH }} />
            ) : (
              <Typography.Text type="secondary">ใช้รูปเริ่มต้นของระบบ (ดูตัวอย่างไม่ได้ในโหมดทดสอบ)</Typography.Text>
            )}
            {card.updateDate && (
              <p className="mb-0 mt-3 text-[13px] text-ink-muted">
                แก้ไขล่าสุด {formatDateTime(card.updateDate)}
                {card.updateBy && ` โดย ${card.updateBy}`}
              </p>
            )}
            {card.imageUrl && (
              <Button className="mt-4" icon={<RollbackOutlined />} onClick={resetToDefault} disabled={saving}>
                ใช้รูปเริ่มต้น
              </Button>
            )}
          </Card>

          <Card title="เปลี่ยนรูป">
            <div className="flex flex-col gap-4">
              <Typography.Text type="secondary">
                JPG หรือ PNG ไม่เกิน 10 MB กว้างอย่างน้อย 600 พิกเซล (แสดงในอีเมลกว้าง {EMAIL_WIDTH} พิกเซล) — มีผลกับอีเมลที่ส่งหลังบันทึก
                อีเมลที่อยู่ในคิวแล้วใช้รูปเดิม
              </Typography.Text>
              <div style={{ maxWidth: EMAIL_WIDTH }}>
                <ImageUploadField value={draft} onChange={setDraft} purpose="email-contact-card" aspect="1600 / 982" fit="contain" disabled={saving} />
              </div>
              <Button type="primary" icon={<SaveOutlined />} disabled={!draft} loading={saving} onClick={() => draft && save(draft, 'บันทึกรูปติดต่อแล้ว')} className="self-start">
                บันทึกรูปนี้
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}
