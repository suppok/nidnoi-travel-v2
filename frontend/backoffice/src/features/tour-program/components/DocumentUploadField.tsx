import { DeleteOutlined, FilePdfOutlined, FileWordOutlined, LoadingOutlined, UploadOutlined } from '@ant-design/icons';
import { App, Button, Upload } from 'antd';
import { useState } from 'react';
import { ACCEPT, checkFile, customRequest, fileNameFromUrl, uploadErrorMessage } from './mediaUpload.ts';

type Props = {
  value?: string | null;
  onChange?: (url: string | null) => void;
  purpose: 'tour-pdf' | 'tour-docx';
  disabled?: boolean;
};

// One PDF or DOCX file as a Form control: value is the uploaded URL.
export default function DocumentUploadField({ value, onChange, purpose, disabled }: Props) {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState(false);
  // The original (possibly Thai) name of a file uploaded in this session; the URL keeps an ASCII name.
  const [uploadedName, setUploadedName] = useState<{ url: string; name: string } | null>(null);
  const isPdf = purpose === 'tour-pdf';
  const Icon = isPdf ? FilePdfOutlined : FileWordOutlined;
  const name = value ? (uploadedName?.url === value ? uploadedName.name : fileNameFromUrl(value)) : '';

  return (
    <div className="flex flex-wrap items-center gap-2">
      {value && (
        <a href={value} target="_blank" rel="noreferrer" className="inline-flex max-w-[360px] items-center gap-2 rounded-lg border border-line bg-white px-3 py-1.5">
          <Icon className={isPdf ? 'text-red-600' : 'text-blue-600'} />
          <span className="truncate">{name}</span>
        </a>
      )}
      <Upload
        accept={ACCEPT[purpose]}
        showUploadList={false}
        disabled={disabled || uploading}
        beforeUpload={(file) => {
          const problem = checkFile(file, purpose);
          if (problem) message.error(problem);
          return problem ? Upload.LIST_IGNORE : true;
        }}
        customRequest={customRequest(purpose, (res, file) => {
          setUploadedName({ url: res.url, name: res.fileName || file.name });
          onChange?.(res.url);
        })}
        onChange={({ file }) => {
          setUploading(file.status === 'uploading');
          if (file.status === 'error') message.error(uploadErrorMessage(file.error, purpose));
        }}
      >
        <Button icon={uploading ? <LoadingOutlined /> : <UploadOutlined />} disabled={disabled || uploading}>
          {uploading ? 'กำลังอัปโหลด…' : value ? 'เปลี่ยนไฟล์' : `อัปโหลด ${isPdf ? 'PDF' : 'DOCX'}`}
        </Button>
      </Upload>
      {value && (
        <Button type="text" danger icon={<DeleteOutlined />} disabled={disabled || uploading} onClick={() => onChange?.(null)} aria-label="นำไฟล์ออก" />
      )}
    </div>
  );
}
