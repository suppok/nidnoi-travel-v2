import { env } from '../config/env';

export const SITE_NAME = 'นิดหน่อย ทราเวล';

export type SeoProps = {
  title: string;
  description: string;
  /** Path with trailing slash, e.g. "/tours/japan/". */
  path: string;
  /** Absolute image URL; defaults to the site-wide OG image. */
  ogImage?: string;
  ogType?: 'website' | 'article';
};

export function absoluteUrl(path: string): string {
  return `${env.SITE_URL}${path.startsWith('/') ? path : `/${path}`}`;
}

export function pageTitle(title: string): string {
  return title === SITE_NAME ? title : `${title} | ${SITE_NAME}`;
}

export function defaultOgImage(): string {
  return absoluteUrl('/og-default.jpg');
}
