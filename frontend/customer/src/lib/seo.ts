import { env } from '../config/env';
import type { TourProgramPublicDetail } from '../types/tour-program';

export const SITE_NAME = 'นิดหน่อย ทราเวล';

export type SeoProps = {
  title: string;
  description: string;
  /** Path with trailing slash, e.g. "/tours/japan/". */
  path: string;
  /** Absolute image URL; defaults to the site-wide OG image. */
  ogImage?: string;
  ogType?: 'website' | 'article';
  /** JSON-LD objects rendered into <head>. */
  jsonLd?: Record<string, unknown>[];
  /** Keeps the page out of search results on every environment (forms such as /tour/<code>/booking/). */
  noindex?: boolean;
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

/** Trims text to a meta-description length on a word boundary. */
export function metaDescription(text: string, max = 158): string {
  const clean = text.replace(/\s+/g, ' ').trim();
  if (clean.length <= max) return clean;
  const cut = clean.slice(0, max - 1);
  const space = cut.lastIndexOf(' ');
  return `${(space > max * 0.6 ? cut.slice(0, space) : cut).trimEnd()}…`;
}

export type Crumb = { name: string; path: string };

export function breadcrumbJsonLd(crumbs: Crumb[]): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: crumbs.map((c, i) => ({
      '@type': 'ListItem',
      position: i + 1,
      name: c.name,
      item: absoluteUrl(c.path),
    })),
  };
}

/** TouristTrip with the build-time starting price as the offer (no ratings until they are real). */
export function touristTripJsonLd(tour: TourProgramPublicDetail, image: string): Record<string, unknown> {
  const url = absoluteUrl(`/tour/${tour.code}/`);
  return {
    '@context': 'https://schema.org',
    '@type': 'TouristTrip',
    name: tour.name,
    description: tour.highlight || metaDescription(tour.description),
    url,
    image,
    touristType: 'Group tour',
    itinerary: {
      '@type': 'ItemList',
      itemListElement: [...tour.countries, ...tour.cities].map((p, i) => ({
        '@type': 'ListItem',
        position: i + 1,
        item: { '@type': 'Place', name: p.nameTh },
      })),
    },
    ...(tour.startingPrice !== null && {
      offers: {
        '@type': 'Offer',
        price: tour.startingPrice,
        priceCurrency: 'THB',
        url,
        availability: tour.periods.some((p) => !p.soldOut)
          ? 'https://schema.org/InStock'
          : 'https://schema.org/SoldOut',
      },
    }),
    provider: { '@type': 'TravelAgency', name: SITE_NAME, url: absoluteUrl('/') },
  };
}
