import { DeleteOutlined, InboxOutlined, LeftOutlined, RightOutlined } from '@ant-design/icons';
import { App, Button, Image, Upload, type UploadProps } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { mediaApi } from '../../../api/media.ts';
import { ACCEPT, checkFile, MIN_SIZE, uploadErrorMessage } from './mediaUpload.ts';

export const MAX_GALLERY_IMAGES = 30;

type Props = {
  value?: string[];
  onChange?: (urls: string[]) => void;
  disabled?: boolean;
};

// Gallery as a Form control: value is the list of uploaded URLs in display order.
export default function GalleryField({ value = [], onChange, disabled }: Props) {
  const { message } = App.useApp();
  const [pendingIds, setPendingIds] = useState<ReadonlySet<string>>(new Set());
  const pending = pendingIds.size;
  // Uploads finish one by one; appending to the latest list keeps every one of them.
  const latest = useRef(value);
  useEffect(() => {
    latest.current = value;
  }, [value]);

  const emit = (urls: string[]) => {
    latest.current = urls;
    onChange?.(urls);
  };

  const move = (index: number, delta: number) => {
    const next = [...value];
    const [item] = next.splice(index, 1);
    next.splice(index + delta, 0, item);
    emit(next);
  };

  const uploadRequest: NonNullable<UploadProps['customRequest']> = ({ file, onSuccess, onError }) => {
    mediaApi
      .upload(file as File, 'tour-gallery')
      .then((res) => {
        const current = latest.current;
        if (current.length < MAX_GALLERY_IMAGES && !current.includes(res.url)) emit([...current, res.url]);
        onSuccess?.(res);
      })
      .catch((err: unknown) => onError?.(err instanceof Error ? err : new Error(String(err))));
  };

  const full = value.length + pending >= MAX_GALLERY_IMAGES;

  return (
    <div className="flex flex-col gap-3">
      {value.length > 0 && (
        <Image.PreviewGroup>
          <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-3">
            {value.map((url, i) => (
              <div key={url} className="flex flex-col gap-1.5 rounded-xl border border-line bg-white p-1.5">
                <Image src={url} alt="" className="rounded-lg object-cover" style={{ aspectRatio: '4 / 3', width: '100%' }} />
                <div className="flex items-center justify-between">
                  <span className="pl-1 text-xs text-ink-muted">{i === 0 ? 'รูปแรก' : `#${i + 1}`}</span>
                  <span className="flex">
                    <Button size="small" type="text" icon={<LeftOutlined />} disabled={disabled || i === 0} onClick={() => move(i, -1)} aria-label="เลื่อนไปก่อนหน้า" />
                    <Button size="small" type="text" icon={<RightOutlined />} disabled={disabled || i === value.length - 1} onClick={() => move(i, 1)} aria-label="เลื่อนไปถัดไป" />
                    <Button size="small" type="text" danger icon={<DeleteOutlined />} disabled={disabled} onClick={() => emit(value.filter((_, j) => j !== i))} aria-label="ลบรูป" />
                  </span>
                </div>
              </div>
            ))}
          </div>
        </Image.PreviewGroup>
      )}

      <Upload.Dragger
        multiple
        accept={ACCEPT['tour-gallery']}
        showUploadList={false}
        disabled={disabled || full}
        beforeUpload={(file) => {
          const problem = checkFile(file, 'tour-gallery');
          if (problem) message.error(`${file.name}: ${problem}`);
          return problem ? Upload.LIST_IGNORE : true;
        }}
        customRequest={uploadRequest}
        onChange={({ file }) => {
          setPendingIds((prev) => {
            const next = new Set(prev);
            if (file.status === 'uploading') next.add(file.uid);
            else next.delete(file.uid);
            return next;
          });
          if (file.status === 'error') message.error(`${file.name}: ${uploadErrorMessage(file.error, 'tour-gallery')}`);
        }}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">{pending > 0 ? `กำลังอัปโหลด ${pending} รูป…` : 'ลากรูปมาวาง หรือคลิกเพื่อเลือกหลายรูป'}</p>
        <p className="ant-upload-hint">
          JPG, PNG หรือ WebP ไม่เกิน 10 MB ต่อรูป · อย่างน้อย {MIN_SIZE['tour-gallery']} พิกเซล · สูงสุด {MAX_GALLERY_IMAGES} รูป ({value.length}/{MAX_GALLERY_IMAGES})
        </p>
      </Upload.Dragger>
    </div>
  );
}
