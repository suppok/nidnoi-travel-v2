# Project: nidnoi travel — Tour Program Platform

A platform for displaying international tour programs to customers, with a back-office for staff to manage tour content. It replaces the legacy site at `nidnoitravel.com`; until cutover the new platform runs on the interim domain `nidnoitravelv2.com` (see "Domains").

## Architecture Overview

This is a **multi-project repository** with two top-level domains: `frontend/` and `backend/`. Each frontend app is an independent Node project with its own `package.json`. There is **no monorepo tooling** (no Turborepo, no pnpm workspaces) — keep each project self-contained.

```
.
├── .github/workflows/
│   ├── backend.yml           # paths: backend/**             → vet, test, build
│   ├── customer.yml          # paths: frontend/customer/**   → lint, typecheck, build
│   └── backoffice.yml        # paths: frontend/backoffice/** → lint, typecheck, build
├── .claude/settings.json     # deny rules: Claude may not read or edit .env files
├── backend/                  # Go API (Fiber + Bun + PostgreSQL)
├── frontend/
│   ├── customer/             # Public site (Astro, fully static, React islands)
│   └── backoffice/           # Staff admin (React + Vite + Ant Design)
├── .editorconfig
├── .gitignore                # .env*, !*.example, dist/, bin/, node_modules/
├── CLAUDE.md
└── README.md                 # local setup + deploy map
```

Each subproject has its own `CLAUDE.md` with details specific to that codebase. Read the one closest to where you are working.

### Component responsibilities

- **`frontend/customer`** — Public site where customers browse tour programs. Built with **Astro** as a **fully static site** (no adapter, no server routes) and served from **Cloudflare Workers static assets**. Tour content is baked into HTML at build time; the backend triggers a rebuild when content changes. Fast-changing data (seats, per-period prices) is fetched live by a small React island. SEO, Core Web Vitals and visual polish are hard requirements.
- **`frontend/backoffice`** — Internal tool for staff to create and edit tour programs, periods, seats and prices. Authenticated with JWT and protected by **Cloudflare Access**. Static SPA on Cloudflare Workers. Allowed to ship heavier dependencies (Ant Design).
- **`backend`** — Single Go service exposing the REST API for both frontends. Owns the PostgreSQL database and all secrets. Layered architecture (Handler → UseCase → Repository, shared services/clients via registries). Triggers customer-site rebuilds through a Cloudflare Deploy Hook — frontends never call it. See `backend/CLAUDE.md`.

### Hosting and data

| Component | Runs on | Notes |
|---|---|---|
| Customer site | Cloudflare Workers (static assets only) | Prerendered HTML; rebuilt via deploy hook |
| Backoffice | Cloudflare Workers (static assets, SPA fallback) | Behind Cloudflare Access |
| Backend | Railway, Singapore | Exactly one always-on instance per environment |
| Database | Supabase Postgres, Singapore | Postgres only (no Supabase Auth/Storage/Realtime); session pooler |
| Media | Cloudflare R2 | Backend-only credentials; public via `media.*` custom domain |
| DNS / TLS / CDN | Cloudflare, zone `nidnoitravelv2.com` (full setup) | Moves to zone `nidnoitravel.com` at cutover |

```
staff   ─▶ admin.*  (backoffice SPA, Cloudflare Access) ─▶ api.* (backend) ─▶ Supabase
                                                              ├─▶ R2 ─▶ media.*
                                                              └─ debounced deploy hook ─▶ Workers Builds ─▶ www.* rebuilt
visitor ─▶ www.*    (static HTML) ── period island / inquiry form ─▶ api.* public endpoints
```

## Environments

Three environments: `local`, `dev`, `prd`. Each deployed environment has its own Supabase project, R2 bucket, Workers, deploy hook and secrets.

| | dev | prd |
|---|---|---|
| Git branch | `develop` (auto-deploy) | `main` (merge from `develop` to release) |
| Customer site | `dev.nidnoitravelv2.com` | `www.nidnoitravelv2.com` |
| Backoffice | `dev-admin.nidnoitravelv2.com` | `admin.nidnoitravelv2.com` |
| API | `dev-api.nidnoitravelv2.com` | `api.nidnoitravelv2.com` |
| Media | `dev-media.nidnoitravelv2.com` | `media.nidnoitravelv2.com` |
| Railway environment | `dev` | `production` |
| Supabase | org `nidnoi-dev` (Free) | org `nidnoi` (Pro) |
| R2 bucket | `nidnoi-travel-media-dev` | `nidnoi-travel-media-prod` |
| Workers | `customer-dev`, `backoffice-dev` | `customer-prd`, `backoffice-prd` |

Rules:
- **Single-level subdomains only** (`dev-api.…`, never `api.dev.…`). Free Universal SSL covers the apex and one level of subdomain.
- **DNS lives on Cloudflare.** Every hostname above is on the interim zone `nidnoitravelv2.com`; the legacy site keeps running on `nidnoitravel.com` untouched until cutover (see "Domains").
- **Dev is private:** Cloudflare Access protects `dev.*`, `dev-admin.*` and `admin.*`. The dev customer site also renders `noindex` so it can never compete with prd in Google.
- **Prd customer site is public on the interim domain:** `www.nidnoitravelv2.com` launches publicly and is indexed (no Access, no `noindex`). Its rankings and inbound links must be carried to `nidnoitravel.com` at cutover, so checklist step 5 (301 every v2 URL) is **mandatory**, not optional. `media.*` / `dev-media.*` are always public — Image Transformations and link previews fetch them. `api.*` and `dev-api.*` are **not** behind Access (islands and builds call them; staff endpoints are protected by JWT).
- `api.*` is proxied through Cloudflare (SSL mode Full (strict)). A Cache Rule makes `/api/v1/tour-programs/public/*/periods` eligible for edge caching, respecting the origin's `Cache-Control` (30 s). Everything else on `api.*` bypasses cache.
- Secrets are per environment — never reuse a value across dev and prd. Prd secrets are sealed in Railway.
- **Never copy prd data into dev.** Inquiries contain customer personal data (PDPA).

### Domains

`nidnoitravelv2.com` is an **interim** domain for both dev and prd. The platform moves to `nidnoitravel.com` (replacing the legacy site) later.

- Never hardcode a domain in code. Hostnames come only from env/config (`PUBLIC_SITE_URL`, `PUBLIC_API_BASE_URL`, `VITE_API_BASE_URL`, `R2_PUBLIC_URL`, `CORS_ALLOWED_ORIGINS`), so the cutover is a config change, not a code change. The only committed exceptions are `frontend/customer/public/robots.txt` and docs.
- Keep the same subdomain names on both zones (`www`, `admin`, `api`, `media`, `dev`, `dev-admin`, `dev-api`, `dev-media`).
- Legacy references (the old site, legacy URLs, legacy data import) stay `nidnoitravel.com`.

Cutover checklist (`nidnoitravelv2.com` → `nidnoitravel.com`):
1. Add the subdomains on the `nidnoitravel.com` zone: Workers custom domains, R2 custom domains, Railway custom domains, Cloudflare Access apps, the `api.*` Cache Rule, Image Transformations ("This zone only").
2. Update env values everywhere: Workers build variables, Railway variables (`R2_PUBLIC_URL`, `CORS_ALLOWED_ORIGINS`), `.example` files, CI workflows, `robots.txt`, these docs. Rebuild both frontends.
3. Media URLs stored in the database point at `media.nidnoitravelv2.com` — rewrite them to `media.nidnoitravel.com` in a migration (keep the old R2 custom domain connected until done).
4. Switch `www.nidnoitravel.com` to the customer Worker; apex `nidnoitravel.com` 301 → `https://www.nidnoitravel.com` (Redirect Rule). Legacy URLs are handled by `public/_redirects`.
5. **Mandatory** — prd is public and indexed on v2. 301 every `nidnoitravelv2.com` host and path to the same path on `nidnoitravel.com` (Bulk Redirect, path and query preserved), and keep it for at least a year — never let `nidnoitravelv2.com` lapse while it still redirects. In Search Console, verify both domains, use **Change of Address** from `nidnoitravelv2.com` to `nidnoitravel.com`, and submit the new sitemap.

## Accounts

- One account/organization per provider; dev and prd are separated inside it (except Supabase, which uses a separate Free org for dev).
- Every provider is owned by a **shared business email** (not a personal one), with 2FA on and recovery codes in a shared password manager. People join with their own accounts as members.
- Providers: GitHub (org + repo), Railway (one project, two environments), Supabase (`nidnoi` Pro + `nidnoi-dev` Free), Cloudflare (one account: DNS, Workers, R2, Access).
- Railway Hobby cannot invite workspace members. Until the workspace moves to Pro, one person operates the Railway dashboard; everyone else deploys through `git push`.

### Cost baseline

Fixed: **~$30/month** — Railway Hobby $5 (includes $5 usage for both environments) + Supabase Pro $25 (prd). Everything on Cloudflare and the Supabase dev org is on free tiers.

Things that add cost — **ask before enabling any of them**: Railway Pro (team access), a larger Supabase compute size or PITR, Workers Paid ($5), more than 5,000 unique Cloudflare image transformations per month, any new paid service. If a Railway hard usage limit is set, remember that reaching it takes **all** workloads offline.

## Working Directory Rules

**Critical:** Always `cd` into the relevant subproject before running commands. Each subproject has its own dependencies and tooling.

- Customer site: `cd frontend/customer`
- Back-office: `cd frontend/backoffice`
- API: `cd backend`

Do **not** install npm packages at the repo root. Do **not** create a root-level `package.json` or `go.mod`.

## Quick Command Reference

### Frontend — `frontend/customer` (Astro)
```bash
npm install
npm run dev         # Astro dev server (localhost:4321)
npm run build       # static build → dist/
npm run preview     # serve the build locally
npm run lint        # ESLint
npm run typecheck   # astro check
```

### Frontend — `frontend/backoffice` (Vite + React)
```bash
npm install
npm run dev         # Vite dev server (localhost:5173)
npm run build       # production build
npm run lint        # ESLint
npm run typecheck   # tsc --noEmit
```

### Backend
```bash
task run-api        # start the API with config/local.env
task migrate-up     # apply migrations (also: migrate-down, migrate-create -- <desc>)
task swag           # regenerate Swagger docs after handler changes
go build ./...
go test ./pkg/...
```

## Cross-Cutting Conventions

### General
- All code, comments and identifiers in **English**, even when the conversation is in Thai.
- Match the style of existing code in the file before introducing new patterns.
- Ask before adding a new dependency to any subproject.

### API conventions
- REST endpoints under `/api/v1/`. Resource names plural and `kebab-case` (`/api/v1/tour-programs`).
- JSON bodies use **camelCase** in both directions.
- Route conventions (all subprojects must follow them):

| Method + path | Meaning |
|---|---|
| `POST /x` | create |
| `PUT /x` | update (id in body) |
| `POST /x/list` | list/search with a filter body |
| `GET /x/:id` | get one |
| `PUT /x/status` | soft delete / restore |
| `GET /x/public…` | public read for the customer site, no auth |

- Success: `{ "success": true, "data": … }`. Error: `{ "success": false, "error": { "code": "TAXONOMY_IN_USE", "message": "...", "fields": { … } } }`. `fields` is present only for `VALIDATION`.
- Error codes are UPPER_SNAKE strings (`VALIDATION`, `UNAUTHENTICATED`, `PERMISSION_DENIED`, `NOT_FOUND`, `TAXONOMY_IN_USE`, …). Frontends branch on them; a shipped code is never renamed.
- Permission codes are `module:action` (`tour-program:create`, `site:rebuild`). Defined in `backend/pkg/constants/apicode/`, mirrored in `frontend/backoffice/src/constants/permissions.ts`.

### Environment & Secrets

**Strict rule — do not break this:**

- **NEVER read, view, open, cat, grep, edit, or modify** any `.env` file under any circumstance. This includes `.env`, `.env.local`, `.env.production`, `frontend/*/.env.local`, `backend/config/local.env`, and any other variant. `.claude/settings.json` enforces this with deny rules — never loosen them.
- If a value from an env file is needed, **stop and ask the user** for that specific value.
- Only `*.example` files are safe to read and edit. They must never contain real secrets.
- When adding a variable: update the matching `.example` file (placeholder), update the config schema, then tell the user to set the real value locally **and** in the deploy target (Railway variables or Workers Builds build variables) themselves.
- Never commit secrets. Never paste real secrets into chat or commit messages.
- Frontends have **no secrets at all**: every `PUBLIC_*` / `VITE_*` value ships to the browser.

### Tour identity
- Every tour program has an internal UUID `id` and a public string `code` (e.g. `nidn261690`).
- `code` appears in customer URLs (`/tour/nidn261690/`) and inbound links. It is **stable** — never changes after create, even when the title changes.
- Frontends address tours by `code` on public endpoints; the backoffice uses `id` internally. Foreign keys and audit references always use `id`.
- `code` is generated by the backend from a Postgres sequence at create time and never accepted from a request body. See `backend/CLAUDE.md`.

### Tour taxonomy
- **Country** — `japan`, `korea`, `taiwan`, … **One or more per tour.**
- **City** — `tokyo`, `osaka`, `hokkaido`, … Zero or more per tour; each city belongs to one country.
- Each value has a stable `slug` and Thai/English display names. Slugs are **locked after create** — they appear in URLs (`/tours/japan/`, `/tours/japan/tokyo/`). Display names, descriptions and hero images are editable.

> **Phase 2:** a **Tag** dimension (ซากุระ, ใบไม้แดง, ครอบครัว) following the same pattern, driving `/tours/tag/<slug>/`. Keep the country/city implementation additive-friendly; don't pre-build for it.

### Periods, prices and seats
- Each tour has periods (departures) with dates, price and seat availability. **Seats and prices are the data that changes most often** (many times a day); tour content rarely changes after posting.
- The tour's **starting price** (lowest active period price) is baked into static HTML. Per-period prices and seats are served live through `GET /api/v1/tour-programs/public/:code/periods` to an island on the tour page.
- Seat changes never trigger a site rebuild. Price changes trigger one only when the starting price changes.

### SEO migration

The legacy site (`nidnoitravel.com`) has indexed landing pages at URLs that don't fit the new structure (e.g. `/intertours/japan/`, `/hokkaido/`, `/tokyo-osaka/`). Their traffic is not the main channel (most traffic comes from ads and social) but it is **not zero**, and third-party links point at them.

The agreed approach:
1. **Keep `/tour/<code>/` URLs identical** — every legacy tour-detail link works without redirects.
2. **Build country/city landing pages from day one.**
3. **301-redirect every known legacy non-detail URL** to its closest new equivalent. The map lives in `frontend/customer/public/_redirects` and is a versioned artifact.
4. **Accept a temporary ranking dip** for 4–8 weeks after launch while Google re-crawls.

Cross-cutting principle: **no public-facing URL pattern is introduced without first deciding which legacy URLs it absorbs (if any).** Details in `frontend/customer/CLAUDE.md`.

### Customer site rebuild contract
- Only the backend triggers rebuilds, by POSTing to the customer Worker's **Cloudflare Deploy Hook** after a committed mutation that changes published content.
- The backend debounces: it fires after 3 minutes without new changes, and at most 15 minutes after the first pending change. A build then takes a few minutes. Staff should expect content changes to appear within minutes, seat changes within ~30 seconds.
- `CLOUDFLARE_DEPLOY_HOOK_URL` is a backend-only secret (the URL itself is the credential). One hook per environment.
- Manual trigger for staff: `POST /api/v1/site/rebuild` (permission `site:rebuild`), exposed as a button in the backoffice.
- A failed build leaves the previous deployment live.

### CORS

| Env | Allowed origins |
|---|---|
| local | `http://localhost:4321`, `http://localhost:5173` |
| dev | `https://dev.nidnoitravelv2.com`, `https://dev-admin.nidnoitravelv2.com` |
| prd | `https://www.nidnoitravelv2.com`, `https://admin.nidnoitravelv2.com` |

Set through the backend's `CORS_ALLOWED_ORIGINS`. Workers preview URLs are not allowlisted — test against `dev`.

### Object storage (Cloudflare R2)
- All uploaded media (tour covers, gallery photos, taxonomy hero images, future PDFs) lives in R2, one bucket per environment.
- **Upload flow:** backoffice → backend → R2. The backend is the only holder of R2 credentials; it validates, re-encodes images to WebP, strips EXIF and returns the public URL.
- **Public read:** objects are served from `media.*` / `dev-media.*` (R2 custom domains). The customer site resizes them at the edge with **Cloudflare Image Transformations** (`/cdn-cgi/image/…`).
- **No presigned-URL uploads from frontends.**
- Backend-only secrets: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_PUBLIC_URL`.

## Do / Don't

### Do
- Read the subproject `CLAUDE.md` before working in that subproject.
- Run `go build ./...`, `go test ./pkg/...` and `npm run lint` / `npm run typecheck` before declaring a change complete.
- Keep the `frontend/customer` build small — it is shipped to every visitor (see its JS budget).
- Update `.example` files when adding an environment variable.
- Keep `develop` deployable; release by merging `develop` → `main`.

### Don't
- **NEVER read, open, view, edit, or modify any `.env` file** (`.env`, `.env.local`, `.env.production`, `config/local.env`, …). If a value is needed, ask the user.
- Don't introduce a monorepo tool (Turborepo, Nx, pnpm workspaces) without discussion.
- Don't share code between `frontend/customer` and `frontend/backoffice` via cross-project imports — raise it first.
- Don't bypass the API — frontends never connect to PostgreSQL.
- Don't call the Cloudflare Deploy Hook from any frontend, and never put its URL in a `PUBLIC_*` / `VITE_*` variable or client code.
- Don't run more than one backend instance until a real queue exists.
- Don't enable a paid tier or add-on without asking.
- Don't commit build output (`dist/`, `bin/`, `node_modules/`).

## Open Questions / TBD

Flag the user before assuming an answer. Subproject-specific questions live in each subproject's `CLAUDE.md`.

- **Queue infrastructure** — RabbitMQ vs Redis vs none. Not needed while the backend runs as one instance; required before scaling out.
- **Redis** — not in the stack; add only with a concrete need.
- **Legacy data import** — bulk import from the old `nidnoitravel.com` data. Must preserve tour `code` values and move the tour code sequence past the highest imported code.
- **Analytics / PDPA consent** — Cloudflare Web Analytics (cookieless) vs GA4 vs Plausible, and whether a consent banner is required.
- **Railway team access** — stay on Hobby (one operator) or move to Pro for workspace members.
- **WebP encoder** for backend image processing — see `backend/CLAUDE.md`.
- **Inquiry spam protection** — rate limiting only for now; consider Cloudflare Turnstile before launch.

### Resolved (recorded for context)
- **Customer site:** Astro, fully static, on Cloudflare Workers static assets; rebuilt through a debounced deploy hook. Seats and per-period prices served live through an island. (Vercel ISR was dropped: content rarely changes, Vercel Hobby forbids commercial use, and Pro adds $20/month for freshness this site doesn't need.)
- **Backoffice:** React + Vite + Ant Design, static SPA on Cloudflare Workers behind Cloudflare Access.
- **Backend:** Go on Railway (Singapore), one always-on instance per environment. Structure follows the SODA-U backend layout (apicode, errorcode, `query.DB[T]`, `DBTx`, response DTO per action).
- **Database:** Supabase Postgres (Pro for prd, Free org for dev), session pooler connection.
- **Object storage:** Cloudflare R2, one bucket per environment; backend-only credentials.
- **Environments:** local / dev / prd; `develop` → dev, `main` → prd; single-level subdomains under `nidnoitravelv2.com` (interim), moving to `nidnoitravel.com` at cutover.
- **Tour detail URL:** `/tour/<code>/`, inherited from the legacy site.
- **Taxonomy (phase 1):** country (one or more per tour) + city (zero or more), stable slugs, landing pages from day one. Tag deferred.
- **SEO migration:** keep `/tour/<code>/`, new landing pages, 301 legacy non-detail URLs, accept a temporary dip.
- **Customer-site i18n:** Thai-only; English only for proper nouns.
- **Route conventions:** `POST` create, `PUT` update, `POST /list`, `GET /:id`, `PUT /status`.
