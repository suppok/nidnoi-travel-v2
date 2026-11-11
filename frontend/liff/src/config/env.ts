export const env = {
  APP_ENV: import.meta.env.VITE_APP_ENV as string,
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL as string,
  LIFF_ID: import.meta.env.VITE_LIFF_ID as string,
  // Local preview without LINE: skips LIFF login and fakes the LIFF-token endpoints (src/api/mock.ts).
  MOCK: import.meta.env.VITE_LIFF_MOCK === 'true',
} as const

if (env.MOCK && env.APP_ENV !== 'local') {
  throw new Error('VITE_LIFF_MOCK is allowed only with VITE_APP_ENV=local')
}
