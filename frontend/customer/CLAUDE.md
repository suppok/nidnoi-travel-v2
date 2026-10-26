# Customer Site — Development Guide

Public-facing website where customers browse international tour programs. **Read-only** consumption of the backend's public API. Optimized for **SEO, speed, and visual polish** above all else.

**Stack:** Astro (static output, no adapter) · TypeScript · Tailwind CSS · React (islands only) · Cloudflare Workers static assets · Cloudflare Image Transformations

> This is a **public, indexable site**. SEO is a hard requirement. Every page ships as prerendered HTML and only hydrates JavaScript where genuinely interactive. Anything that breaks the "static HTML first" model needs discussion.

---

## Why Astro, Static, on Cloudflare

Decided up front — don't relitigate without a concrete reason:

- The site is read-mostly. Tour content rarely changes after it is posted; what changes many times a day is **seats and per-period prices**.
- Astro ships **zero JS by default**; React hydrates only the islands that need it.
- The whole site is **prerendered**. When content changes, the backend triggers a rebuild through a Cloudflare Deploy Hook (debounced). Seats and per-period prices are fetched live by one small island, so they never need a rebuild.
- Requests to static assets on Cloudflare Workers are free and unlimited, and the rest of the stack (DNS, R2, Access) is already on Cloudflare. Vercel ISR was dropped: Vercel Hobby forbids commercial use, and Pro adds $20/month for per-page freshness this site doesn't need.

Revisit only if tour **content** must appear on the site within seconds of being saved. If a feature seems to need SPA behavior across many routes, the answer is usually "use an island," not "abandon Astro."

---

## Folder Organization

```
frontend/customer/
├── public/                       # Served as-is
│   ├── _redirects                # Legacy 301 map (versioned artifact)
│   ├── robots.txt
│   ├── favicon.svg, og-default.jpg
│   └── fonts/
├── src/
│   ├── pages/                    # File-based routing — every page is prerendered
│   │   ├── index.astro           # Home
│   │   ├── tour/
│   │   │   └── [code].astro      # Tour detail — /tour/nidn261690/ (matches legacy URL)
│   │   ├── tour-programs/
│   │   │   └── index.astro       # All tours, client-side filters via query string
│   │   ├── tours/
│   │   │   └── [country]/
│   │   │       ├── index.astro   # Country landing — /tours/japan/
│   │   │       └── [city].astro  # City landing    — /tours/japan/tokyo/
│   │   ├── about.astro
│   │   ├── contact.astro
│   │   └── 404.astro
│   ├── layouts/
│   │   ├── BaseLayout.astro      # <html>, <head>, SEO tags, fonts, dev noindex
│   │   └── PageLayout.astro      # Header + footer + slot
│   ├── components/               # .astro components — static, no client JS
│   │   ├── ui/                   # Button, Card, Container, CfImage, ...
│   │   ├── tour/                 # TourCard, TourHero, TourPriceBlock, ...
│   │   └── layout/               # Header, Footer, Nav
│   ├── islands/                  # React components that hydrate (interactive only)
│   │   ├── PeriodTable.tsx       # Live seats + per-period prices on /tour/<code>/
│   │   ├── TourFilters.tsx       # Filters on /tour-programs/
│   │   ├── InquiryForm.tsx       # Lead capture
│   │   └── ImageGallery.tsx      # Lightbox / carousel
│   ├── lib/
│   │   ├── api/                  # Backend client (build time + islands)
│   │   │   ├── client.ts
│   │   │   ├── tour-programs.ts
│   │   │   └── taxonomy.ts
│   │   ├── image.ts              # Cloudflare transformation URL builder + presets
│   │   ├── seo.ts                # <title>, meta, JSON-LD helpers
│   │   ├── format.ts             # formatPrice, formatDate
│   │   └── thai.ts               # Thai-specific helpers
│   ├── types/                    # TS types matching backend DTOs
│   ├── styles/
│   │   └── global.css            # Tailwind directives, font-face, resets
│   └── config/
│       └── env.ts                # Typed access to import.meta.env
├── astro.config.mjs
├── wrangler.jsonc                # Workers static assets config (env.dev / env.prd)
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

**Folder roles:**
- `pages/` is the router. One file = one URL. No server endpoints (`pages/api/` does not exist).
- `components/` is `.astro` — static by default, no client JS.
- `islands/` is `.tsx` — ships JS to every visitor. Treat it as a budget.
- `lib/api/` runs **at build time** (pages, `getStaticPaths`) and **in the browser** (islands only: periods, inquiry). Browser calls are limited to those two endpoints.

---

## Quick Commands

```bash
cd frontend/customer
npm install
npm run dev         # Astro dev server (localhost:4321)
npm run build       # static build → ./dist
npm run preview     # serve the build locally
npm run lint        # ESLint
npm run typecheck   # astro check
```

Always `cd frontend/customer` first. Never run npm at the repo root.

---

## Responsibilities of This App

1. List active tour programs with filters (country, city, date range, price).
2. Show a tour detail page as static HTML with full meta tags, plus live period availability.
3. Serve country and city landing pages (primary SEO surface).
4. Capture inquiries that POST to the public backend endpoint.
5. Serve static content (about, contact, terms).

It does **not**:
- Log users in or call authenticated endpoints.
- Create, edit, or delete any data.
- Talk to the database.
- Render on a server. There is no server runtime — only static files.

---

## Rendering & Data Freshness

Every page is prerendered at build time (`output: 'static'`, no adapter).

| Data | Where it lives | How it updates |
|---|---|---|
| Tour content, taxonomy, starting price, landing pages | Static HTML | Backend fires the deploy hook (debounced: 3 min quiet, max 15 min) → Workers Builds rebuilds → typically live within minutes |
| Per-period dates, prices, seats left, sold-out | `PeriodTable` island | Fetched on view from `GET /tour-programs/public/:code/periods` (cached ~30 s) |
| About / contact / terms | Static HTML | Rebuilt on every deploy |

Build rules:
- Fetch data in page frontmatter / `getStaticPaths` through `src/lib/api/`. Keep the number of calls per build small; if builds slow down, ask the backend for a bulk endpoint instead of adding per-tour calls.
- **Never swallow API errors at build time.** A failed build leaves the previous deployment live, which is the safe outcome. A build that silently renders empty pages is not.
- Pages that should not exist are simply not generated: a city under the wrong country, or a country/city with zero active tours. They fall through to `404.astro`.

```js
// astro.config.mjs
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  output: 'static',            // no adapter, no server routes
  site: process.env.PUBLIC_SITE_URL,
  trailingSlash: 'always',
  build: { format: 'directory' }, // /tours/japan/ → /tours/japan/index.html
  integrations: [react(), sitemap(/* see Sitemap */)],
});
```

---

## Period Island (seats & prices)

`src/islands/PeriodTable.tsx` renders the departures table on `/tour/<code>/`.

```astro
<PeriodTable client:visible code={tour.code} initialPeriods={tour.periods} />
```

- The table is rendered **at build time** with `initialPeriods`, so bots and no-JS visitors see real data and nothing shifts on hydration (CLS).
- On hydration it fetches `GET /tour-programs/public/:code/periods` once and replaces the rows. On error it keeps the build-time data silently.
- No polling. One request per page view.
- Keep the markup identical between the static render and the hydrated render (same columns, same row height).
- Cards and JSON-LD show the **starting price** from build time, labelled "ราคาเริ่มต้น". On the detail page, the island's numbers are the source of truth.

---

## URL Strategy & Legacy Redirects

The customer site has a fixed set of public URL patterns. Every new page must fit one of them or be discussed first.

### Canonical URL patterns

| Pattern | Example | Purpose |
|---|---|---|
| `/` | `/` | Home |
| `/tour/<code>/` | `/tour/nidn261690/` | Tour detail — **matches the legacy site exactly** |
| `/tour-programs/` | `/tour-programs/?country=japan&dateFrom=2026-03-01` | All tours; query-string filters applied client-side |
| `/tours/<country.slug>/` | `/tours/japan/` | Country landing |
| `/tours/<country.slug>/<city.slug>/` | `/tours/japan/tokyo/` | City landing (city must belong to the country) |
| `/about/`, `/contact/`, `/terms/` | — | Static content |

> `/tours/tag/<slug>/` is reserved for the phase-2 tag dimension — don't use this prefix for anything else.

**Rules:**
- Trailing slash is **required**. Astro builds `directory` format; Workers static assets (`html_handling: "auto-trailing-slash"`) redirect `/tours/japan` → `/tours/japan/`. Never serve the same content at both.
- Country and city slugs are lowercase ASCII (backend-enforced). Do not URL-encode or translate them.
- `getStaticPaths` for `[city].astro` only emits (country, city) pairs where the city belongs to that country. `/tours/korea/tokyo/` is therefore a 404.
- An empty landing (zero active tours) is not generated → 404. Empty stubs hurt SEO more than 404s.

### Legacy redirect map

Every known legacy URL must 301 to its closest new equivalent. The map lives in `public/_redirects`, which Workers applies to static asset requests:

```
# frontend/customer/public/_redirects
# Country hubs
/intertours/japan/*    /tours/japan/            301
/intertours/korea/*    /tours/korea/            301
/intertours/taiwan/*   /tours/taiwan/           301

# City hubs (some legacy URLs lack the /intertours/ prefix)
/intertours/tokyo/*    /tours/japan/tokyo/      301
/intertours/osaka/*    /tours/japan/osaka/      301
/intertours/nagoya/*   /tours/japan/nagoya/     301
/intertours/nikko/*    /tours/japan/nikko/      301
/hokkaido/*            /tours/japan/hokkaido/   301

# Combo pages → country landing; the user can re-filter from there
/tokyo-osaka/*         /tours/japan/            301

# /tour/<code>/ is NOT redirected — it works natively.
```

**Maintenance rules:**
- The map is a versioned artifact; every entry is reviewed in PR.
- Always `301`. Lines with `*` count as dynamic redirects (limit 100; 2,000 static). Beyond that, use Cloudflare Bulk Redirects.
- If the legacy site also served a path without a trailing slash, add that exact path as a separate static line.
- Before launch, crawl the legacy sitemap and confirm every URL category has a matching redirect. Unknown URLs fall through to the 404 page; watch Search Console for them and add entries.
- **Never remove** a redirect entry, even years later.

### When a country/city slug must "change"

It can't — the backend locks slugs after create. If a slug truly must change (e.g. a typo found after launch): create a new record with the correct slug, move tour associations, soft-delete the old record, and **add a line to `public/_redirects`** from the old path to the new one before any traffic sees the change.

---

## Deployment (Cloudflare Workers)

Two Workers built from this folder by **Workers Builds**:

| | `customer-dev` | `customer-prd` |
|---|---|---|
| Branch | `develop` | `main` |
| Domain | `dev.nidnoitravelv2.com` (behind Cloudflare Access) | `www.nidnoitravelv2.com` (public; moves to `nidnoitravel.com` at cutover — see root CLAUDE.md "Domains") |
| Deploy command | `npx wrangler deploy --env dev` | `npx wrangler deploy --env prd` |

```jsonc
// wrangler.jsonc — assets only, no Worker script
{
  "name": "customer",
  "compatibility_date": "2026-10-01",
  "assets": {
    "directory": "./dist",
    "not_found_handling": "404-page",
    "html_handling": "auto-trailing-slash"
  },
  "env": {
    "dev": { "name": "customer-dev" },
    "prd": { "name": "customer-prd" }
  }
}
```

- Workers Builds settings: root directory `frontend/customer`, build command `npm ci && npm run build`, build watch paths `frontend/customer/**`. Turn off builds for non-production branches to save build minutes.
- Each Worker has its own **Deploy Hook**. Its URL is given to the backend of the matching environment as `CLOUDFLARE_DEPLOY_HOOK_URL`. This app never calls a deploy hook.
- `PUBLIC_*` variables are read **at build time**, so they are set as **build variables** in Workers Builds, not as runtime vars.
- Free plan limits that matter: 3,000 build minutes/month, one concurrent build, 20,000 static files per version.

---

## Backend Integration

All API calls go through `src/lib/api/client.ts`. Never call `fetch` directly from a page or component.

- **Base URL:** `PUBLIC_API_BASE_URL` via `src/config/env.ts`.
- **Public endpoints only** — registered without auth in the backend, handler names containing `Public`. Never attach an `Authorization` header.
- **Browser calls** are limited to `GET /tour-programs/public/:code/periods` (period island) and `POST /inquiries/public` (inquiry form). These need this site's origin in the backend CORS allowlist; everything else runs at build time.
- **Error shape:** `{ success: false, error: { code, message, fields? } }`. The wrapper throws a typed `ApiError`.
- **JSON casing:** camelCase in both directions.

### Client skeleton (reference)
```ts
// src/lib/api/client.ts
export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
    public fields?: Record<string, string>,
  ) {
    super(message);
  }
}

async function request<T>(method: 'GET' | 'POST', path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${env.API_BASE_URL}${path}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  });
  const json = await res.json().catch(() => null);
  if (!res.ok || json?.success === false) {
    throw new ApiError(
      json?.error?.code ?? 'UNKNOWN',
      json?.error?.message ?? res.statusText,
      res.status,
      json?.error?.fields,
    );
  }
  return json.data as T;
}

export const apiGet = <T>(path: string) => request<T>('GET', path);
export const apiPost = <T>(path: string, body: unknown) => request<T>('POST', path, body);
```

---

## Images (Cloudflare Image Transformations)

Media lives in R2 and is served from `media.nidnoitravelv2.com` (`dev-media.*` on dev). The backend already stores WebP at original dimensions. This site resizes at the edge with Cloudflare Image Transformations — **nothing is processed at build time**.

- Use `src/components/ui/CfImage.astro` (and the matching helper in islands) for every R2 image. Never write a raw `<img src="https://media...">`.
- URLs come from `src/lib/image.ts`:

```ts
// src/lib/image.ts
export const IMAGE_PRESETS = {
  card: [400, 800],
  hero: [960, 1920],
  gallery: [800, 1600],
} as const;

export function cfImageUrl(src: string, width: number): string {
  if (env.SITE_ENV === 'local') return src; // no /cdn-cgi on localhost
  return `/cdn-cgi/image/width=${width},quality=80,format=auto,onerror=redirect/${src}`;
}
```

- **Budget:** the free plan includes 5,000 unique transformations per month. Every (image × width) pair counts. Keep presets to two widths, keep them in this one file, and never add per-request parameters. On the free plan, transformations beyond the limit fail (the `onerror=redirect` option falls back to the original) rather than being billed.
- Always render explicit `width`/`height` (or a fixed aspect ratio) and meaningful `alt`.
- The LCP image (hero) uses `loading="eager"` and `fetchpriority="high"`; every other image is lazy.
- `og:image` must be absolute: `https://www.nidnoitravelv2.com/cdn-cgi/image/width=1200,height=630,fit=cover,quality=80/<src>`.

---

## SEO Requirements — Hard

Every indexable page must have:

1. **`<title>`** — unique, 50–60 chars, Thai. Format: `<page title> | <site name>`.
2. **`<meta name="description">`** — unique, 140–160 chars, Thai.
3. **`<link rel="canonical">`** — absolute URL with trailing slash, built from `PUBLIC_SITE_URL`.
4. **Open Graph** — `og:title`, `og:description`, `og:image`, `og:url`, `og:type`, `og:locale="th_TH"`. **OG image matters most** — most traffic comes from LINE/Facebook shares.
5. **Twitter Card** — `twitter:card="summary_large_image"`, `twitter:title`, `twitter:description`, `twitter:image`.
6. **JSON-LD** — `TouristTrip` on tour detail pages (offer price = starting price), `BreadcrumbList` on country/city landings.
7. **`<html lang="th">`** — set in `BaseLayout.astro`.

These are built by `src/lib/seo.ts` and a `<SEO />` component inside `BaseLayout.astro`.

### Dev must never be indexed

`BaseLayout.astro` renders `<meta name="robots" content="noindex, nofollow">` **only when `PUBLIC_SITE_ENV !== 'prd'`**. The dev site is also behind Cloudflare Access. Before every release, check that a prd build contains no `noindex` anywhere.

### Landing pages — extra requirements

Country and city landings are the primary SEO surface (with `/tour/<code>/`):

- **H1 contains the destination name in Thai** — "ทัวร์ญี่ปุ่น", not "Japan Tours".
- **Above-the-fold copy** from the destination's `description` (80–200 words, authored in the backoffice). Never substitute placeholder or auto-generated copy.
- **Hero image** from `heroImage`; fall back to a generic image if missing — never ship a landing page without a hero.
- **BreadcrumbList JSON-LD** with the full path.
- **Tour list** renders every active tour in the destination in the initial HTML. Lazy-load images, never lazy-render markup.
- **Internal linking** — country links to its cities; city links back to its country.
- **No empty landings** — not generated when there are zero active tours.

### Sitemap

`@astrojs/sitemap` with custom serialization for dynamic routes:
- `/` and static pages; every active tour `/tour/<code>/`; every country/city landing with ≥ 1 active tour.
- Priority: home 1.0, country 0.8, city 0.7, tour detail 0.6, static 0.4.
- `changefreq`: home & country `daily`, others `weekly`.
- `lastmod`: the record's actual last-update timestamp.

Submit `sitemap.xml` to Google Search Console and Bing Webmaster Tools at launch; re-submit on structural URL changes.

### `robots.txt`

```
User-agent: *
Allow: /

Sitemap: https://www.nidnoitravelv2.com/sitemap.xml
```

### Migration monitoring (first 8 weeks after launch)

Track in Google Search Console and act on regressions:
- **Coverage** — 404 spikes usually mean a legacy URL needs a line in `public/_redirects`.
- **Performance** — old top pages should keep or recover position; audit any page that craters.
- **Core Web Vitals** — should stay green; verify monthly.
- **Indexed pages** — should climb steadily; expect 4–8 weeks for full re-indexing.

The first 8 weeks have an explicit "ranking dip is acceptable" budget. After that, treat any drop as a bug.

---

## Performance — Hard Budgets

| Metric | Target | Why |
|---|---|---|
| **LCP** | < 2.0 s on 4G | Core Web Vital |
| **CLS** | < 0.05 | Core Web Vital |
| **TBT** | < 100 ms | Core Web Vital |
| **JS per route (gzipped)** | < 50 KB | Islands add up fast |
| **Largest image** | < 200 KB | Use the transformation presets |
| **Lighthouse Performance** | ≥ 95 mobile | Build-fail threshold |

### Rules to hit those budgets

- **Default to static.** A new component is `.astro`. Add a `client:` directive only when it must react to user input.
- **Hydration directives:**
  - `client:load` — above-the-fold and immediately interactive (rare; requires a justification comment).
  - `client:idle` — most cases.
  - `client:visible` — below-the-fold islands (period table, gallery).
  - `client:media` — breakpoint-conditional islands.
- **Images** — always through `CfImage` and the presets (see Images).
- **Fonts** — self-hosted in `public/fonts/`, `font-display: swap`, at most two families, subset to Thai + Latin.
- **No CSS-in-JS.** Tailwind + scoped Astro `<style>` is enough.
- **Check `dist/_astro/*.js` sizes** after `npm run build` before merging anything that adds a dependency.

---

## Visual Design — Make It Look Good

This is a tour business — the site must look polished and trustworthy. Treat design quality as a deliverable.

When building UI, consult the `frontend-design` skill for design direction, typography and layout.

- **Typography hierarchy** — distinct sizes/weights for h1/h2/h3/body. Body 16–18 px minimum for Thai.
- **Generous whitespace** — tour cards and detail pages should breathe.
- **Imagery leads** — tour photos are the product. Large, full-bleed where appropriate, proper aspect ratios.
- **Restricted palette** — one primary, a neutral scale, one accent. No ad-hoc colors.
- **Thai typography care** — line height ≥ 1.6. Test with real Thai content.
- **Mobile-first** — most traffic is mobile.

Design tokens (colors, spacing, font sizes, radii) live in the Tailwind theme. Never hardcode hex values inline.

---

## Styling

- **Tailwind** for layout and styling.
- **No component library** (no Ant Design, MUI or shadcn copies). Ant Design is reserved for `frontend/backoffice`.
- **Astro scoped styles** are fine for component-specific tweaks.
- **One global stylesheet** — `src/styles/global.css` (Tailwind entry, font-face, resets).

---

## Forms & Lead Capture

`src/islands/InquiryForm.tsx` POSTs to `POST /api/v1/inquiries/public` (rate-limited by the backend).

- react-hook-form + zod for validation (aligned with the backoffice).
- The server validates again — never trust client validation alone. Map `error.fields` back onto inputs.
- Clear success and error states; show the backend's `error.message`.
- Disable submit while pending; show a spinner.
- After success, show a confirmation block — don't silently clear the form.

---

## State Management

Almost none — by design.

1. **URL state** — filters, pagination, sort in search params; deep links must work.
2. **Local state** inside islands — `useState`.
3. **No global store.** Raise it before adding one.

---

## Accessibility

- Interactive elements are real `<button>` / `<a>` — never `<div onClick>`.
- Meaningful `alt`; decorative images use `alt=""`.
- WCAG AA color contrast — verify, Tailwind defaults are not automatically compliant.
- Visible labels linked via `htmlFor` / `id`.
- Keyboard navigation: visible focus, tab order matches visual order, Esc closes modals.

---

## Environment Variables

> **CRITICAL — secrets protection:**
> **NEVER read, view, open, cat, grep, edit, or modify** any `.env` file (`.env`, `.env.local`, `.env.production`, or any variant). If a value is needed, **stop and ask the user**. Only `.env.example` is safe to read and edit.

This app has **no secrets**. Every variable is `PUBLIC_*`, read at build time, and visible to anyone. Access them through `src/config/env.ts`, which validates them and fails the build if one is missing.

```
# .env.example
PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
PUBLIC_SITE_URL=http://localhost:4321
PUBLIC_SITE_ENV=local            # local | dev | prd
```

| Variable | dev | prd |
|---|---|---|
| `PUBLIC_API_BASE_URL` | `https://dev-api.nidnoitravelv2.com/api/v1` | `https://api.nidnoitravelv2.com/api/v1` |
| `PUBLIC_SITE_URL` | `https://dev.nidnoitravelv2.com` | `https://www.nidnoitravelv2.com` |
| `PUBLIC_SITE_ENV` | `dev` | `prd` |

When adding a variable: update `.env.example`, the schema in `src/config/env.ts`, then tell the user to set the real value locally and as a **build variable** on both Workers in Workers Builds.

---

## TypeScript Conventions

- `strict: true` via `astro check`. Don't loosen it.
- Avoid `any`; prefer `unknown` + narrowing.
- Backend DTO types live in `src/types/`. Nullable backend fields (`null.String`, `null.Time`, `null.Float`) are `string | null`, `string | null` (ISO date), `number | null`.
- `type` for shapes and unions; `interface` only when declaration merging is needed.

---

## Internationalization

**Thai-only.** English appears only as proper nouns.

- `<html lang="th">`, `og:locale="th_TH"`.
- Dates: `Intl.DateTimeFormat('th-TH', …)`. Buddhist vs Gregorian year is still open — decide once, apply everywhere.
- Numbers: `Intl.NumberFormat('th-TH')`.
- No i18n routing, no locale prefix, no translation library. Revisit only when a second language is actually added.

---

## Do / Don't

### Do
- Default to `.astro`; promote to an island only when interactivity is required.
- Consult the `frontend-design` skill before building visible UI.
- Use `CfImage` for every R2 image, with explicit dimensions and good `alt`.
- Set `<title>`, description, canonical, OG and JSON-LD through `<SEO />` on every page.
- Let builds fail loudly when the API fails.
- Run `npm run build` and check bundle sizes before merging anything that adds a dependency.
- Update `.env.example` when adding a variable.

### Don't
- **NEVER read, open, view, edit, or modify any `.env` file** — only `.env.example`.
- Don't add an Astro adapter, on-demand rendering or server endpoints without discussion — the site is fully static by design.
- Don't call a Cloudflare Deploy Hook from this app.
- Don't add a UI component library.
- Don't add SPA-style client routing.
- Don't call authenticated backend endpoints, or add browser calls beyond the period and inquiry endpoints without discussion.
- Don't add image presets or transformation parameters outside `src/lib/image.ts`.
- Don't import from `frontend/backoffice`.
- Don't add `client:load` without a justification comment.
- Don't introduce a global state library without a concrete need.

---

## Adding a New Page — Checklist

1. Create `src/pages/<route>.astro` (with a trailing-slash URL that fits the canonical patterns).
2. Use `BaseLayout` / `PageLayout` and pass SEO props (title, description, ogImage).
3. For dynamic routes, implement `getStaticPaths` from the public API; don't emit pages that should 404.
4. Build interactive parts as islands in `src/islands/` with the right `client:*` directive.
5. Confirm the route appears in the sitemap with the right priority.
6. If the page shows backend data, confirm the backend enqueues a rebuild when that data changes (see the rebuild table in `backend/CLAUDE.md`).
7. If the page absorbs legacy URLs, add them to `public/_redirects`.
8. Test mobile layout, keyboard navigation and Lighthouse before merging.

---

## Open Questions / TBD

Flag the user before assuming:

- **Buddhist vs Gregorian year** in date display.
- **Analytics provider** — Cloudflare Web Analytics (cookieless) vs GA4 vs Plausible; PDPA implications.
- **Consent / cookie banner** — required only if analytics uses cookies.
- **Preview of unpublished tours** — the static site can't show unsaved/unpublished content; decide whether staff preview on `dev` or in a backoffice preview screen.
- **A/B testing** — out of scope.
- **Inquiry spam protection** — Cloudflare Turnstile on `InquiryForm` before launch?
- **Full legacy URL audit** — the current `_redirects` map comes from search results; crawl the legacy sitemap end-to-end before launch.
