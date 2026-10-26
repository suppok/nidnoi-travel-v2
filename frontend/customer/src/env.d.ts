interface ImportMetaEnv {
  readonly PUBLIC_API_BASE_URL?: string;
  readonly PUBLIC_SITE_URL?: string;
  readonly PUBLIC_SITE_ENV?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
