const APP_ENVS = ['local', 'dev', 'prd'] as const;

export type AppEnv = (typeof APP_ENVS)[number];

function required(name: string, value: string | undefined): string {
  if (!value) {
    throw new Error(`Missing environment variable ${name} — copy .env.example to .env`);
  }
  return value;
}

function appEnv(value: string | undefined): AppEnv {
  const v = required('VITE_APP_ENV', value);
  if (!(APP_ENVS as readonly string[]).includes(v)) {
    throw new Error(`VITE_APP_ENV must be one of ${APP_ENVS.join(', ')}, got "${v}"`);
  }
  return v as AppEnv;
}

function flag(name: string, value: string | undefined): boolean {
  if (value === undefined || value === '' || value === 'false') return false;
  if (value === 'true') return true;
  throw new Error(`${name} must be "true" or "false", got "${value}"`);
}

const APP_ENV = appEnv(import.meta.env.VITE_APP_ENV);
const API_MOCK = flag('VITE_API_MOCK', import.meta.env.VITE_API_MOCK);

if (API_MOCK && APP_ENV === 'prd') {
  throw new Error('VITE_API_MOCK must not be enabled when VITE_APP_ENV=prd');
}

export const env = {
  API_BASE_URL: required('VITE_API_BASE_URL', import.meta.env.VITE_API_BASE_URL),
  APP_NAME: required('VITE_APP_NAME', import.meta.env.VITE_APP_NAME),
  APP_ENV,
  // Route API calls to src/api/mock/ instead of the backend (local/dev only).
  API_MOCK,
} as const;
