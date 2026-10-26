const SITE_ENVS = ['local', 'dev', 'prd'] as const;

export type SiteEnv = (typeof SITE_ENVS)[number];

// Fails the build when a variable is missing — never ship a site with a wrong API or canonical URL.
function required(name: string, value: string | undefined): string {
  if (!value) {
    throw new Error(`Missing environment variable ${name} — copy .env.example to .env`);
  }
  return value;
}

function siteEnv(value: string | undefined): SiteEnv {
  const v = required('PUBLIC_SITE_ENV', value);
  if (!(SITE_ENVS as readonly string[]).includes(v)) {
    throw new Error(`PUBLIC_SITE_ENV must be one of ${SITE_ENVS.join(', ')}, got "${v}"`);
  }
  return v as SiteEnv;
}

export const env = {
  API_BASE_URL: required('PUBLIC_API_BASE_URL', import.meta.env.PUBLIC_API_BASE_URL),
  SITE_URL: required('PUBLIC_SITE_URL', import.meta.env.PUBLIC_SITE_URL).replace(/\/+$/, ''),
  SITE_ENV: siteEnv(import.meta.env.PUBLIC_SITE_ENV),
} as const;
