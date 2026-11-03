// Public taxonomy DTOs. Mirrors the planned backend responses for
// GET /countries/public and GET /cities/public (see backend/CLAUDE.md "Public endpoints").

export type CountryPublic = {
  slug: string;
  nameTh: string;
  nameEn: string;
  /** Landing-page intro, authored in the backoffice (plain text). */
  description: string;
  heroImage: string | null;
  sortOrder: number;
  /** Active (public) tours in this country. Landings with 0 are not generated. */
  tourCount: number;
  /** Lowest starting price among its tours; null when no tour has a priced period. */
  startingPrice: number | null;
  updateDate: string;
};

export type CityPublic = {
  slug: string;
  countrySlug: string;
  nameTh: string;
  nameEn: string;
  description: string;
  heroImage: string | null;
  sortOrder: number;
  tourCount: number;
  startingPrice: number | null;
  updateDate: string;
};

/** Compact reference embedded in tour DTOs. */
export type TaxonomyRef = {
  slug: string;
  nameTh: string;
  nameEn: string;
};

export type CityRef = TaxonomyRef & { countrySlug: string };
