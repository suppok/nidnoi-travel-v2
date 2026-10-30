import type { Paging } from './api.ts';

// Fields shared by countries and cities.
type TaxonomyItem = {
  id: string;
  refId: number | null; // legacy destination id from the old site
  slug: string; // used in customer URLs; locked after create
  nameTh: string;
  nameEn: string;
  description: string; // SEO copy for the landing page
  heroImage: string | null;
  sortOrder: number;
  status: boolean; // false = soft-deleted
  createDate: string;
  updateDate: string;
};

export type Country = TaxonomyItem;

export type CityCountry = Pick<Country, 'id' | 'slug' | 'nameTh' | 'nameEn' | 'status'>;

export type City = TaxonomyItem & {
  countryId: string;
  country: CityCountry;
};

// POST /countries/list
export type CountryListRequest = Paging & {
  keyword?: string; // matches Thai/English name or slug
  status?: boolean; // omitted = all
};

// POST /cities/list
export type CityListRequest = CountryListRequest & {
  countryId?: string;
};

type TaxonomyFields = {
  refId: number | null;
  nameTh: string;
  nameEn: string;
  description: string;
  heroImage: string | null;
  sortOrder: number;
};

// POST /countries
export type CountryCreateRequest = TaxonomyFields & { slug: string };

// PUT /countries — no slug (locked after create)
export type CountryUpdateRequest = TaxonomyFields & { id: string };

// POST /cities
export type CityCreateRequest = CountryCreateRequest & { countryId: string };

// PUT /cities
export type CityUpdateRequest = CountryUpdateRequest & { countryId: string };

// PUT /countries/status, PUT /cities/status
export type TaxonomyStatusRequest = {
  id: string;
  status: boolean;
};
