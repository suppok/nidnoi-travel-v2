import { PlusOutlined } from '@ant-design/icons';
import { App, Upload, type UploadFile } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { lineBookingsApi } from '../../../api/lineBookings.ts';
import { errorMessage } from '../../../lib/errors.ts';

// Slips per payment; MUST match the backend (maxSlips).
export const MAX_SLIPS = 10;
const MAX_BYTES = 10 * 1024 * 1024;
const ACCEPT = 'image/jpeg,image/png,image/webp';

type Props = {
  value?: string[];
  onChange?: (urls: string[]) => void;
  disabled?: boolean;
};

// One or more payment slips (a payment split over several transfers), as a Form control whose value
// is the list of uploaded URLs. Several files can be picked at once; each uploads on its own.
export default function SlipUploadField({ value = [], onChange, disabled }: Props) {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState<UploadFile[]>([]);
  // Uploads finish one by one: each appends to the latest list, not the one it started with.
  const valueRef = useRef(value);
  const pending = useRef(0);
  useEffect(() => {
    valueRef.current = value;
  });

  const fileList: UploadFile[] = [
    ...value.map((url, i): UploadFile => ({ uid: url, name: `สลิป ${i + 1}`, status: 'done', url, thumbUrl: url })),
    ...uploading,
  ];

  return (
    <Upload
      listType="picture-card"
      accept={ACCEPT}
      multiple
      disabled={disabled}
      fileList={fileList}
      beforeUpload={(file) => {
        if (!ACCEPT.split(',').includes(file.type)) {
          message.error('เลือกได้เฉพาะไฟล์รูป JPG, PNG หรือ WebP');
          return Upload.LIST_IGNORE;
        }
        if (file.size > MAX_BYTES) {
          message.error('ไฟล์ใหญ่เกินไป (ไม่เกิน 10 MB)');
          return Upload.LIST_IGNORE;
        }
        if (valueRef.current.length + pending.current >= MAX_SLIPS) {
          message.warning(`แนบสลิปได้ไม่เกิน ${MAX_SLIPS} รูป`);
          return Upload.LIST_IGNORE;
        }
        pending.current += 1;
        return true;
      }}
      customRequest={({ file }) => {
        const f = file as File & { uid: string };
        setUploading((list) => [...list, { uid: f.uid, name: f.name, status: 'uploading' }]);
        lineBookingsApi
          .uploadSlip(f)
          .then((res) => {
            const next = [...valueRef.current, res.url];
            valueRef.current = next;
            onChange?.(next);
          })
          .catch((err: unknown) => message.error(errorMessage(err)))
          .finally(() => {
            pending.current -= 1;
            setUploading((list) => list.filter((x) => x.uid !== f.uid));
          });
      }}
      onRemove={(file) => {
        if (file.url) onChange?.(valueRef.current.filter((u) => u !== file.url));
        return false;
      }}
      onPreview={(file) => file.url && window.open(file.url, '_blank', 'noopener')}
    >
      {fileList.length < MAX_SLIPS && !disabled && (
        <button type="button" className="flex flex-col items-center gap-1 border-0 bg-transparent text-ink-muted">
          <PlusOutlined />
          <span className="text-xs">เพิ่มสลิป</span>
        </button>
      )}
    </Upload>
  );
}
