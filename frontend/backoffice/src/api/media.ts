import type { MediaPurpose, MediaUploadResponse } from '../types/media.ts';
import { request } from './client.ts';

export const mediaApi = {
  // The backend checks the type by content, the size and the minimum image dimensions.
  upload: (file: File, purpose: MediaPurpose) => {
    const body = new FormData();
    body.append('file', file);
    body.append('purpose', purpose);
    return request<MediaUploadResponse>('POST', '/media/upload', body);
  },
};
