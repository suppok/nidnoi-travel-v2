import type { UploadProps } from 'antd';
import { mediaApi } from '../../../api/media.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { MediaPurpose, MediaUploadResponse } from '../../../types/media.ts';
import { ApiError } from '../../../api/client.ts';
import { ErrorCodes } from '../../../constants/errorCodes.ts';

export const MAX_UPLOAD_BYTES = 10 * 1024 * 1024;

export const ACCEPT: Record<MediaPurpose, string> = {
  'tour-banner': 'image/jpeg,image/png,image/webp',
  'tour-card': 'image/jpeg,image/png,image/webp',
  'tour-gallery': 'image/jpeg,image/png,image/webp',
  'country-hero': 'image/jpeg,image/png,image/webp',
  'city-hero': 'image/jpeg,image/png,image/webp',
  'airline-logo': 'image/jpeg,image/png,image/webp',
  'tour-pdf': 'application/pdf,.pdf',
  'tour-docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document,.docx',
};

// Minimum sizes enforced by the backend, shown as hints.
export const MIN_SIZE: Partial<Record<MediaPurpose, string>> = {
  'tour-banner': '1200×630',
  'tour-card': '800×600',
  'tour-gallery': '800×600',
  'country-hero': '1600×900',
  'city-hero': '1600×900',
  'airline-logo': '120×40',
};

// Client-side guard only; the backend re-checks type (by content), size and dimensions.
export function checkFile(file: File, purpose: MediaPurpose): string | null {
  if (file.size > MAX_UPLOAD_BYTES) return 'ไฟล์ใหญ่เกินไป (ไม่เกิน 10 MB)';
  const accepted = ACCEPT[purpose].split(',');
  const ext = `.${file.name.split('.').pop()?.toLowerCase() ?? ''}`;
  if (!accepted.includes(file.type) && !accepted.includes(ext)) return 'ไฟล์ประเภทนี้อัปโหลดในช่องนี้ไม่ได้';
  return null;
}

export function uploadErrorMessage(err: unknown, purpose: MediaPurpose): string {
  if (err instanceof ApiError && err.code === ErrorCodes.ImageTooSmall && MIN_SIZE[purpose]) {
    return `รูปมีขนาดเล็กเกินไป ต้องมีขนาดอย่างน้อย ${MIN_SIZE[purpose]} พิกเซล`;
  }
  return errorMessage(err);
}

// antd Upload customRequest that goes through the API client (POST /media/upload).
export function customRequest(purpose: MediaPurpose, onUploaded: (res: MediaUploadResponse, file: File) => void): UploadProps['customRequest'] {
  return ({ file, onSuccess, onError }) => {
    mediaApi
      .upload(file as File, purpose)
      .then((res) => {
        onUploaded(res, file as File);
        onSuccess?.(res);
      })
      .catch((err: unknown) => onError?.(err instanceof Error ? err : new Error(String(err))));
  };
}

// Display name of an uploaded document: the last URL segment.
export function fileNameFromUrl(url: string): string {
  try {
    const last = new URL(url).pathname.split('/').pop() ?? '';
    return decodeURIComponent(last) || url;
  } catch {
    return url;
  }
}
