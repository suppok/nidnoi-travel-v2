import { DeleteOutlined, LoadingOutlined, PictureOutlined, SwapOutlined } from '@ant-design/icons';
import { App, Button, Image, Upload } from 'antd';
import { useState } from 'react';
import type { MediaPurpose } from '../../../types/media.ts';
import { ACCEPT, checkFile, customRequest, MIN_SIZE, uploadErrorMessage } from './mediaUpload.ts';

type Props = {
  value?: string | null;
  onChange?: (url: string | null) => void;
  purpose: MediaPurpose;
  disabled?: boolean;
  aspect?: string; // CSS aspect-ratio of the preview, e.g. "1200 / 630"
  fit?: 'cover' | 'contain'; // contain for logos, which must not be cropped
};

// A single image (banner / hero) as a Form control: value is the uploaded URL.
export default function ImageUploadField({ value, onChange, purpose, disabled, aspect = '1200 / 630', fit = 'cover' }: Props) {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState(false);

  const upload = (
    <Upload
      accept={ACCEPT[purpose]}
      showUploadList={false}
      disabled={disabled || uploading}
      beforeUpload={(file) => {
        const problem = checkFile(file, purpose);
        if (problem) message.error(problem);
        return problem ? Upload.LIST_IGNORE : true;
      }}
      customRequest={customRequest(purpose, (res) => onChange?.(res.url))}
      onChange={({ file }) => {
        setUploading(file.status === 'uploading');
        if (file.status === 'error') message.error(uploadErrorMessage(file.error, purpose));
      }}
    >
      {value ? (
        <Button icon={uploading ? <LoadingOutlined /> : <SwapOutlined />} disabled={disabled || uploading}>
          {uploading ? 'กำลังอัปโหลด…' : 'เปลี่ยนรูป'}
        </Button>
      ) : (
        <button
          type="button"
          disabled={disabled || uploading}
          className="flex w-full max-w-[560px] cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-line bg-sky-soft font-sans text-ink-muted hover:border-sky-strong disabled:cursor-not-allowed"
          style={{ aspectRatio: aspect }}
        >
          {uploading ? <LoadingOutlined className="text-2xl" /> : <PictureOutlined className="text-2xl" />}
          <span className="text-sm">{uploading ? 'กำลังอัปโหลด…' : 'คลิกเพื่ออัปโหลดรูป'}</span>
          <span className="text-xs">JPG, PNG หรือ WebP ไม่เกิน 10 MB · อย่างน้อย {MIN_SIZE[purpose]} พิกเซล</span>
        </button>
      )}
    </Upload>
  );

  if (!value) return upload;

  return (
    <div className="flex max-w-[560px] flex-col gap-2">
      <Image src={value} alt="" className={`rounded-xl ${fit === 'contain' ? 'border border-line bg-white object-contain p-3' : 'object-cover'}`} style={{ aspectRatio: aspect, width: '100%' }} />
      <div className="flex gap-2">
        {upload}
        <Button icon={<DeleteOutlined />} danger disabled={disabled || uploading} onClick={() => onChange?.(null)}>
          นำออก
        </Button>
      </div>
    </div>
  );
}
