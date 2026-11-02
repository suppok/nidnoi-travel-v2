import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { MediaPurpose, MediaUploadResponse } from '../../types/media.ts';
import { authorize, type MockHandler, MockFail, ok, validation } from './http.ts';

// Mirrors the purpose rules in backend pkg/modules/media/usecase. The mock trusts the browser's
// file type (the backend sniffs the content) and returns a blob: URL, which only lives until the
// page reloads — enough to try the forms.
const IMAGE = ['image/jpeg', 'image/png', 'image/webp'];
const DOCX = 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
const MAX_BYTES = 10 * 1024 * 1024;

const RULES: Record<MediaPurpose, { mimes: string[]; minWidth?: number; minHeight?: number; prefix: string }> = {
  'tour-banner': { mimes: IMAGE, minWidth: 1200, minHeight: 630, prefix: 'tours/banners/' },
  'tour-card': { mimes: IMAGE, minWidth: 800, minHeight: 600, prefix: 'tours/cards/' },
  'tour-gallery': { mimes: IMAGE, minWidth: 800, minHeight: 600, prefix: 'tours/gallery/' },
  'tour-pdf': { mimes: ['application/pdf'], prefix: 'tours/documents/' },
  'tour-docx': { mimes: [DOCX], prefix: 'tours/documents/' },
  'country-hero': { mimes: IMAGE, minWidth: 1600, minHeight: 900, prefix: 'countries/heroes/' },
  'city-hero': { mimes: IMAGE, minWidth: 1600, minHeight: 900, prefix: 'cities/heroes/' },
};

async function upload(req: Parameters<MockHandler>[0]): Promise<Response> {
  authorize(req, 'POST /media/upload', Permissions.MediaUpload);
  if (!(req.body instanceof FormData)) throw validation({ file: 'file is required' });

  const purpose = req.body.get('purpose');
  const file = req.body.get('file');
  if (typeof purpose !== 'string' || !(purpose in RULES)) throw validation({ purpose: 'must be a valid value' });
  if (!(file instanceof File) || file.size === 0) throw validation({ file: 'file is required' });

  const rule = RULES[purpose as MediaPurpose];
  if (file.size > MAX_BYTES) throw new MockFail(413, ErrorCodes.MediaTooLarge, 'The file is too large');
  if (!rule.mimes.includes(file.type)) throw new MockFail(400, ErrorCodes.MediaTypeNotAllowed, 'This file type is not allowed for this purpose');

  let width = 0;
  let height = 0;
  if (rule.minWidth && rule.minHeight) {
    const bitmap = await createImageBitmap(file).catch(() => null);
    if (!bitmap) throw new MockFail(400, ErrorCodes.MediaTypeNotAllowed, 'This file type is not allowed for this purpose');
    ({ width, height } = bitmap);
    bitmap.close();
    if (width < rule.minWidth || height < rule.minHeight) {
      throw new MockFail(400, ErrorCodes.ImageTooSmall, `The image must be at least ${rule.minWidth}×${rule.minHeight} pixels (got ${width}×${height})`);
    }
  }

  const isDocument = !rule.minWidth;
  const response: MediaUploadResponse = {
    url: URL.createObjectURL(file),
    key: `${rule.prefix}${crypto.randomUUID()}`,
    size: file.size,
    mime: file.type,
    fileName: isDocument ? file.name : '',
    width,
    height,
  };
  return ok(response);
}

export const mediaRoutes: Record<string, MockHandler> = {
  'POST /media/upload': upload,
};
