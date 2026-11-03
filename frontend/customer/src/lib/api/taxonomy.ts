// Build-time taxonomy reads.
import type { CityPublic, CountryPublic } from '../../types/taxonomy';
import { apiGet } from './client';
import { USE_MOCK_DATA } from './mock-mode';

const bySortOrder = <T extends { sortOrder: number }>(a: T, b: T) => a.sortOrder - b.sortOrder;

// The header and footer read taxonomy on every page; fetch once per build.
let countriesOnce: Promise<CountryPublic[]> | undefined;
let citiesOnce: Promise<CityPublic[]> | undefined;

/** All active countries, including ones with zero tours (callers filter on tourCount). */
export function listPublicCountries(): Promise<CountryPublic[]> {
  countriesOnce ??= fetchCountries();
  return countriesOnce;
}

async function fetchCountries(): Promise<CountryPublic[]> {
  if (USE_MOCK_DATA) {
    const { mockCountries } = await import('./mock-data');
    return [...mockCountries].sort(bySortOrder);
  }
  return (await apiGet<CountryPublic[]>('/countries/public')).sort(bySortOrder);
}

/** All active cities across countries. */
export function listPublicCities(): Promise<CityPublic[]> {
  citiesOnce ??= fetchCities();
  return citiesOnce;
}

async function fetchCities(): Promise<CityPublic[]> {
  if (USE_MOCK_DATA) {
    const { mockCities } = await import('./mock-data');
    return [...mockCities].sort(bySortOrder);
  }
  return (await apiGet<CityPublic[]>('/cities/public')).sort(bySortOrder);
}
