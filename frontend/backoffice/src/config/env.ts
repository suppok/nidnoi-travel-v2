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

export const env = {
  API_BASE_URL: required('VITE_API_BASE_URL', import.meta.env.VITE_API_BASE_URL),
  APP_NAME: required('VITE_APP_NAME', import.meta.env.VITE_APP_NAME),
  APP_ENV: appEnv(import.meta.env.VITE_APP_ENV),
} as const;
