// @ts-check
import react from '@astrojs/react';
import sitemap from '@astrojs/sitemap';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'astro/config';
import { loadEnv } from 'vite';

// astro.config runs before .env is loaded into import.meta.env, so read it explicitly.
const { PUBLIC_SITE_URL } = loadEnv(process.env.NODE_ENV ?? 'production', process.cwd(), '');

export default defineConfig({
  output: 'static', // no adapter, no server routes
  site: PUBLIC_SITE_URL,
  trailingSlash: 'always',
  build: { format: 'directory' }, // /tours/japan/ → /tours/japan/index.html
  integrations: [react(), sitemap()],
  vite: { plugins: [tailwindcss()] },
  server: { port: 4321 },
});
