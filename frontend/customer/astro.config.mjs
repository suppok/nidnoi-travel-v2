// @ts-check
import react from '@astrojs/react';
import sitemap, { ChangeFreqEnum } from '@astrojs/sitemap';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'astro/config';
import { loadEnv } from 'vite';

// astro.config runs before .env is loaded into import.meta.env, so read it explicitly.
const { PUBLIC_SITE_URL } = loadEnv(process.env.NODE_ENV ?? 'production', process.cwd(), '');

/**
 * Priority / changefreq by URL pattern (see CLAUDE.md "Sitemap").
 * TODO(sitemap): set lastmod from each record's updateDate once the public API is live.
 * @param {string} url
 * @returns {{ priority: number, changefreq: import('@astrojs/sitemap').ChangeFreqEnum }}
 */
function sitemapRank(url) {
  const path = new URL(url).pathname;
  if (path === '/') return { priority: 1.0, changefreq: ChangeFreqEnum.DAILY };
  if (/^\/tours\/[^/]+\/$/.test(path)) return { priority: 0.8, changefreq: ChangeFreqEnum.DAILY };
  if (/^\/tours\/[^/]+\/[^/]+\/$/.test(path)) return { priority: 0.7, changefreq: ChangeFreqEnum.WEEKLY };
  if (path.startsWith('/tour/')) return { priority: 0.6, changefreq: ChangeFreqEnum.WEEKLY };
  return { priority: 0.4, changefreq: ChangeFreqEnum.WEEKLY };
}

export default defineConfig({
  output: 'static', // no adapter, no server routes
  site: PUBLIC_SITE_URL,
  trailingSlash: 'always',
  build: { format: 'directory' }, // /tours/japan/ → /tours/japan/index.html
  integrations: [
    react(),
    sitemap({
      // Booking forms are noindex: keep them out of the sitemap too.
      filter: (page) => !page.includes('/404') && !page.endsWith('/booking/'),
      serialize: (item) => ({ ...item, ...sitemapRank(item.url) }),
    }),
  ],
  vite: { plugins: [tailwindcss()] },
  server: { port: 4321 },
});
