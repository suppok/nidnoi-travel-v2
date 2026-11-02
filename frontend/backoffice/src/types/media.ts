// POST /media/upload purposes. Mirrors backend pkg/modules/media/models.
export type MediaPurpose = 'tour-banner' | 'tour-card' | 'tour-gallery' | 'tour-pdf' | 'tour-docx' | 'country-hero' | 'city-hero';

export type MediaUploadResponse = {
  url: string;
  key: string;
  size: number;
  mime: string;
  fileName: string; // download name for documents; '' for images
  width: number; // 0 for documents
  height: number;
};
