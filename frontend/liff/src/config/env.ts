export const env = {
  APP_ENV: import.meta.env.VITE_APP_ENV as string,
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL as string,
  LIFF_ID: import.meta.env.VITE_LIFF_ID as string,
} as const
