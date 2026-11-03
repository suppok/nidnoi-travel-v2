// Cloudflare Image Transformations — the ONLY place presets and parameters live.
// Free plan: 5,000 unique (image × width) transformations per month, so keep two widths per preset.
import { env } from '../config/env';

export const IMAGE_PRESETS = {
  card: [400, 800],
  hero: [960, 1920],
  gallery: [800, 1600],
} as const;

export type ImagePreset = keyof typeof IMAGE_PRESETS;

export function cfImageUrl(src: string, width: number): string {
  if (env.SITE_ENV === 'local') return src; // no /cdn-cgi on localhost
  return `/cdn-cgi/image/width=${width},quality=80,format=auto,onerror=redirect/${src}`;
}

export function cfImageSrcset(src: string, preset: ImagePreset): { src: string; srcset: string } {
  const [small, large] = IMAGE_PRESETS[preset];
  return {
    src: cfImageUrl(src, small),
    srcset: `${cfImageUrl(src, small)} ${small}w, ${cfImageUrl(src, large)} ${large}w`,
  };
}

/** Absolute 1200×630 share image for og:image / twitter:image. */
export function ogImageUrl(src: string): string {
  if (env.SITE_ENV === 'local') return src;
  return `${env.SITE_URL}/cdn-cgi/image/width=1200,height=630,fit=cover,quality=80/${src}`;
}
