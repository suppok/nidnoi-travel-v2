# Back-Office — Development Guide

Internal staff tool for creating and editing tour programs, periods, seats, prices and the tour taxonomy. **Authenticated** (JWT) and **behind Cloudflare Access**. Optimized for **productivity over bundle size** — staff on desktop, not anonymous visitors.

**Stack:** React · Vite · TypeScript · Tailwind CSS · Ant Design (antd) · React Router · deployed as a static SPA on Cloudflare Workers

> Unlike `frontend/customer`, this app may ship heavier dependencies (Ant Design, rich editors, charts). SEO and bundle size are not primary concerns — correctness, validation and clear UX are.

---

## Folder Organization

```
frontend/backoffice/
├── public/
│   ├── _headers                # X-Robots-Tag: noindex, nofollow
│   └── robots.txt              # Disallow: /
├── src/
│   ├── main.tsx
│   ├── App.tsx                 # Router + global providers (antd ConfigProvider, auth)
│   ├── routes/                 # One folder per route
│   │   ├── login/
│   │   ├── dashboard/
│   │   ├── tour-programs/
│   │   │   ├── list/
│   │   │   ├── create/
│   │   │   └── [id]/edit/      # Includes the periods table
│   │   ├── taxonomy/
│   │   │   ├── countries/
│   │   │   └── cities/
│   │   ├── users/
│   │   └── not-found/
│   ├── features/               # Domain UI grouped by feature
│   │   ├── tour-program/
│   │   │   ├── components/     # TourProgramForm, TourProgramTable, PeriodTable, ...
│   │   │   ├── hooks/
│   │   │   └── schemas/        # zod schemas for forms
│   │   └── taxonomy/
│   ├── components/             # Reusable presentational pieces (incl. SiteRebuildButton)
│   ├── layouts/
│   │   ├── AuthLayout.tsx      # /login — no sidebar
│   │   └── AppLayout.tsx       # Sidebar + header + outlet, requires auth
│   ├── api/
│   │   ├── client.ts           # fetch wrapper: JWT attach, 401 handling, error mapping
│   │   ├── mock/               # Mock handlers used until the real backend is ready
│   │   ├── tour-programs.ts
│   │   ├── periods.ts
│   │   ├── taxonomy.ts
│   │   ├── media.ts
│   │   ├── site.ts             # POST /site/rebuild
│   │   ├── users.ts
│   │   └── auth.ts
│   ├── auth/
│   │   ├── AuthProvider.tsx    # Context: current user, permissions, login, logout
│   │   ├── RequireAuth.tsx
│   │   ├── RequirePermission.tsx
│   │   └── token.ts
│   ├── types/                  # TS types matching backend DTOs
│   ├── hooks/
│   ├── lib/
│   ├── constants/
│   │   └── permissions.ts      # MUST mirror backend apicode
│   ├── styles/
│   └── config/
│       └── env.ts
├── index.html
├── vite.config.ts
├── wrangler.jsonc              # Workers static assets, SPA fallback (env.dev / env.prd)
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

Routes mirror URL paths. A feature folder owns its forms, tables, schemas and hooks. Promote shared pieces to `components/` or `lib/` only when reused across features.

---

## Quick Commands

```bash
cd frontend/backoffice
npm install
npm run dev         # Vite dev server (localhost:5173)
npm run build
npm run preview
npm run lint
npm run typecheck
```

Always `cd frontend/backoffice` first.

---

## Responsibilities of This App

The back-office is the only place where data is mutated. It:

1. Authenticates staff via the backend JWT endpoints.
2. Lists, creates, edits and soft-deletes tour programs.
3. Manages **periods, prices and seats** — the most frequent staff task (many times a day).
4. Manages the **taxonomy** — countries and cities.
5. Uploads media through the backend.
6. Lets staff force a customer-site rebuild.
7. Manages users and roles (where the caller's permissions allow).
8. Shows operational dashboards (inquiries, …).

It does **not**:
- Serve any UI without login other than `/login` and `/not-found`.
- Talk to the database — only to the backend REST API.
- Call the Cloudflare Deploy Hook or any customer-site endpoint.
- Get indexed by search engines.

---

## Authentication

Backend issues a JWT (see `backend/CLAUDE.md`):

1. `POST /api/v1/auth/login` with credentials → access token.
2. The token is attached as `Authorization: Bearer <token>` on every request.
3. On `401`, the client wrapper clears the session and redirects to `/login`.
4. `GET /api/v1/auth/me` is called once after login; the user (id, name, role, permission codes) is held in `AuthProvider`.

### Token storage

Current choice: access token **in memory only** (`AuthProvider` state). A page reload means logging in again until a refresh flow exists.

- If refresh tokens are added, they go in an **HttpOnly cookie set by the backend** — never readable from JS.
- Don't use `localStorage` for tokens without an explicit, documented decision.

Flag the user before changing this.

### Route guards

```tsx
<Route element={<RequireAuth />}>
  <Route element={<AppLayout />}>
    <Route path="/dashboard" element={<Dashboard />} />
    <Route element={<RequirePermission code={Permissions.TourProgramList} />}>
      <Route path="/tour-programs" element={<TourProgramList />} />
    </Route>
  </Route>
</Route>
```

- `RequireAuth` checks for a session; otherwise redirects to `/login`.
- `RequirePermission` checks the user's permission codes from `/auth/me`.
- **Client-side checks are UX only.** The backend re-checks every request with Casbin.

### Permission codes

Live in `src/constants/permissions.ts` and **must match** `backend/pkg/constants/apicode/`. Format `module:action`. Mirror new codes in the same PR as the backend change.

```ts
export const Permissions = {
  TourProgramCreate: 'tour-program:create',
  TourProgramUpdate: 'tour-program:update',
  TourProgramList:   'tour-program:list',
  TourProgramGet:    'tour-program:get',
  TourProgramStatus: 'tour-program:status',
  PeriodUpdate:      'tour-program-period:update',
  CountryCreate:     'country:create',
  CountryUpdate:     'country:update',
  CountryStatus:     'country:status',
  CityCreate:        'city:create',
  CityUpdate:        'city:update',
  CityStatus:        'city:status',
  MediaUpload:       'media:upload',
  SiteRebuild:       'site:rebuild',
  // ...
} as const;
```

Never hardcode permission strings in components — always use `Permissions.*`.

---

## Backend Integration

All API calls go through `src/api/client.ts`. Never call `fetch` directly from a component.

- **Base URL:** `VITE_API_BASE_URL` via `src/config/env.ts`.
- **Mock first:** until the real backend is ready, `client.ts` routes requests to `src/api/mock/`. The mocks follow the real contracts exactly, so switching to the real backend changes only `client.ts`.
- **JWT attach:** the wrapper reads the access token from `AuthProvider`.
- **401:** wrapper clears the session and redirects to `/login`. Components don't handle 401.
- **403:** show a "not authorized" message; don't redirect.
- **JSON casing:** camelCase in both directions.

### Endpoint conventions

| Method + path | Meaning |
|---|---|
| `POST /api/v1/<resource>` | create |
| `PUT /api/v1/<resource>` | update (id in body) |
| `POST /api/v1/<resource>/list` | list/search with a filter body |
| `GET /api/v1/<resource>/:id` | get one |
| `PUT /api/v1/<resource>/status` | soft delete / restore |

Resources: `tour-programs`, `tour-program-periods`, `countries`, `cities`, `users`, `inquiries`. Plus `POST /media/upload` and `POST /site/rebuild`.

### Errors

Response shape: `{ success: false, error: { code, message, fields? } }`.

- `code` is an UPPER_SNAKE string: `VALIDATION`, `UNAUTHENTICATED`, `PERMISSION_DENIED`, `NOT_FOUND`, `CONFLICT`, `TAXONOMY_IN_USE`, … Branch on `code`, never on `message`.
- `fields` (only for `VALIDATION`) maps field names to messages — show them on the matching `Form.Item`.

### Client skeleton (reference)
```ts
// src/api/client.ts
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

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = tokenStore.get();
  const isForm = body instanceof FormData;
  const res = await fetch(`${env.API_BASE_URL}${path}`, {
    method,
    headers: {
      ...(isForm ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body ? (isForm ? body : JSON.stringify(body)) : undefined,
  });

  if (res.status === 401) {
    authBus.emit('unauthenticated');
    throw new ApiError('UNAUTHENTICATED', 'Session expired', 401);
  }

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
```

---

## UI Library — Ant Design

Use **antd** for complex UI: `Table`, `Form`, `DatePicker`, `Select`, `Modal`, `Drawer`, `Upload`, `Tabs`, `Menu`.

- One `<ConfigProvider>` in `App.tsx`; theme tokens live there.
- Prefer antd components over building your own.
- **Tailwind for layout and spacing**, **antd for components**. Don't fight antd styling with `!important` — use theme tokens or documented `className` props.
- Import directly (`import { Button } from 'antd'`); antd v5 tree-shakes with Vite.
- Use antd `Form` for any form with more than three fields.
- Dates and money: `DatePicker` and `InputNumber` with explicit `format` / `precision`.

Ant Design is **only** used in this app — never in code that could end up in `frontend/customer`.

---

## Tour Programs

### Assigning taxonomy

- **Countries** — multi-select (`mode="multiple"`), **required (at least one)**. A tour can span several countries.
- **Cities** — multi-select; options are the cities belonging to the selected countries. Removing a country removes its cities from the selection.
- Use `Select` with `showSearch` and `optionFilterProp="label"` — destination lists grow over time.

### Soft delete

"Delete" means `PUT /tour-programs/status` with `status: false` (restore = `true`). Confirm with `Modal.confirm`. Deleted tours disappear from the customer site on the next rebuild.

---

## Periods, Prices and Seats

Editing seats and prices is the most frequent staff task — optimize this screen for speed.

- The tour edit page has a **periods table** (`features/tour-program/components/PeriodTable.tsx`): antd editable `Table` with inline `InputNumber` for price and seats, and `DatePicker.RangePicker` for dates.
- Save per row through `PUT /tour-program-periods`. Disable the row's save while pending; keep the user on the page.
- Feedback after save:
  - **Seats only:** "บันทึกแล้ว — หน้าเว็บอัปเดตภายในประมาณ 30 วินาที" (served live by the customer site's period island; no rebuild).
  - **Price, dates, or adding/removing a period:** "บันทึกแล้ว — หน้าเว็บจะอัปเดตภายในไม่กี่นาที" (may change the starting price, which triggers a debounced rebuild).
- Never poll the customer site to confirm. Trust the backend.

---

## Tour Taxonomy

Countries and cities drive customer-site landing pages (`/tours/japan/`, `/tours/japan/tokyo/`). Data model in `backend/CLAUDE.md`.

> **Phase 2:** a tag dimension (ซากุระ, ใบไม้แดง, …). The tour form then gains a "Tags" multi-select.

### UI surfaces

- `/taxonomy/countries/` — list with reorder; create/edit modal.
- `/taxonomy/cities/` — list grouped by country, reorder within a country.

Create/edit fields: `slug` (create only), `nameTh`, `nameEn`, `description` (long text — SEO copy for the landing page), `heroImage` (upload), `sortOrder`. Cities also have `countryId` (Select).

### Slug is locked after create — the UI must enforce it

- **Create form:** editable input with live validation — lowercase ASCII kebab-case, `^[a-z0-9]+(-[a-z0-9]+)*$`, max 60 chars. Helper line: "ใช้ใน URL หน้าลิสต์ทัวร์ เช่น /tours/japan/ เปลี่ยนภายหลังไม่ได้". Show a live preview of the resulting URL.
- **Edit form:** render the slug as read-only text (`<Typography.Text code>`, not a disabled input) with the tooltip "Slug ถูกล็อกหลังสร้างเพราะใช้ในลิงก์ภายนอกและ Google index".
- The backend never writes `slug` on update. Both layers must agree.

### Soft delete and `TAXONOMY_IN_USE`

"Delete" = `PUT /countries/status` / `PUT /cities/status` with `status: false`. If any active tour still references the record, the backend returns `TAXONOMY_IN_USE` with the count. Show: "ลบไม่ได้เพราะมีทัวร์ที่ใช้อยู่ X รายการ" and link to the tour list filtered to those tours so staff can reassign them first.

---

## Customer Site Rebuild

The customer site is fully static on Cloudflare Workers. **The backend** rebuilds it through a Cloudflare Deploy Hook after content changes, debounced (fires after 3 minutes without new changes, at most 15 minutes after the first). A build then takes a few minutes.

- This app never calls the deploy hook or any customer-site endpoint.
- After a successful save, show a normal success toast with the hint from the Periods section. Don't poll.
- **Manual rebuild:** `SiteRebuildButton` in the header ("อัปเดตหน้าเว็บตอนนี้"), visible only with `Permissions.SiteRebuild`. It confirms with `Modal.confirm`, calls `POST /api/v1/site/rebuild`, then disables itself for 60 seconds.

---

## Image Uploads

All media is uploaded through the **backend**, which validates it, re-encodes it to WebP and stores it in Cloudflare R2. Frontends never get R2 credentials.

### Component

antd `<Upload>` with a `customRequest` that goes through `src/api/client.ts` (`POST /media/upload`, `FormData`):

- `name` = `"file"`.
- Extra fields: `purpose` (`tour-cover | tour-gallery | country-hero | city-hero`), `refId`.
- `accept` = `"image/jpeg,image/png,image/webp"` (no SVG, no GIF).
- `maxCount` = 1 for cover/hero; unlimited for galleries.
- `beforeUpload` = client-side guard (≤ 10 MB, image types only). The backend re-checks size, type and minimum dimensions (cover ≥ 1200×800, hero ≥ 1600×900).

### Response handling

The backend returns `{ url, key, size, mime }` (`mime` is `image/webp`). Store `url` in form state — it is what the tour/country/city update sends. Preview with the same URL (`media.*` on prd, `dev-media.*` on dev).

### UX rules

- Show upload progress.
- On success, replace the placeholder with the preview.
- On failure, show the backend's error message; don't retry automatically.
- `<Upload.Dragger>` for gallery uploads.
- Gallery order is stored as `sortOrder` on each image record.

### Replacing an image

Upload the new file, then update the parent record's image field to the new URL. The old object is soft-deleted and reaped by the backend after a grace period. Frontends never call a media delete endpoint.

No `VITE_R2_*` variables, no presigned-URL flow — ever.

---

## State Management

1. **Server state** — TanStack Query is the intended choice but is **deferred until the real backend is stable**. Until then, keep data fetching in feature hooks on top of `src/api/`.
2. **URL state** — filters, pagination, sort in search params so filtered views are shareable.
3. **Auth state** — `AuthProvider` only.
4. **Local state** — `useState`, `useReducer`, `Form.useForm()`.

No Redux/Zustand without a concrete need.

---

## Forms & Validation

- antd `Form` + zod schemas in `features/<domain>/schemas/`.
- Map `ApiError.fields` onto the matching `Form.Item`s.
- Disable submit while pending; `message.success` on completion.
- Confirm destructive actions (soft delete) with `Modal.confirm`.

---

## Routing

React Router (data router API), wired in `src/App.tsx` or `src/router.tsx`.

- Paths are lowercase kebab-case (`/tour-programs/:id/edit`).
- Every route under `AppLayout` is behind `RequireAuth`.
- A `not-found` route catches everything else. On Workers, unknown paths serve `index.html` (SPA fallback), so React Router owns 404s.

---

## Environment Variables

> **CRITICAL — secrets protection:**
> **NEVER read, view, open, cat, grep, edit, or modify** any `.env` file (`.env`, `.env.local`, `.env.production`, or any variant). If a value is needed, **stop and ask the user**. Only `.env.example` is safe to read and edit.

- Vite exposes only `VITE_*` variables, and every one of them ships to the browser. This app has **no secrets**.
- Read them through `src/config/env.ts`, which validates them at startup.
- They are read **at build time** — on Cloudflare they are set as **build variables** in Workers Builds.

```
# .env.example
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_NAME=nidnoi Back-Office
VITE_APP_ENV=local              # local | dev | prd — non-prd shows an environment badge in the header
```

| Variable | dev | prd |
|---|---|---|
| `VITE_API_BASE_URL` | `https://dev-api.nidnoitravelv2.com/api/v1` | `https://api.nidnoitravelv2.com/api/v1` |
| `VITE_APP_ENV` | `dev` | `prd` |

When adding a variable: update `.env.example` and `src/config/env.ts`, then tell the user to set it locally and as a build variable on both Workers.

---

## TypeScript Conventions

- `strict: true`. Don't loosen it.
- Avoid `any`; prefer `unknown` + narrowing.
- Backend DTO types in `src/types/`. Nullable backend fields (`null.String`, `null.Time`, `null.Float`) are `string | null`, `string | null` (ISO date), `number | null`.
- `type` for shapes and unions; `interface` only for declaration merging.

---

## Deployment (Cloudflare Workers)

Two Workers built from this folder by **Workers Builds**:

| | `backoffice-dev` | `backoffice-prd` |
|---|---|---|
| Branch | `develop` | `main` |
| Domain | `dev-admin.nidnoitravelv2.com` | `admin.nidnoitravelv2.com` |
| Deploy command | `npx wrangler deploy --env dev` | `npx wrangler deploy --env prd` |

```jsonc
// wrangler.jsonc — assets only, SPA fallback
{
  "name": "backoffice",
  "compatibility_date": "2026-10-01",
  "assets": {
    "directory": "./dist",
    "not_found_handling": "single-page-application"
  },
  "env": {
    "dev": { "name": "backoffice-dev" },
    "prd": { "name": "backoffice-prd" }
  }
}
```

- Workers Builds: root directory `frontend/backoffice`, build command `npm ci && npm run build`, build watch paths `frontend/backoffice/**`.
- **Cloudflare Access** protects both hostnames (allow-list of staff emails). This is required in addition to JWT — never rely on JWT alone for an internal tool. Access is free for up to 50 users.
- The API (`api.*` / `dev-api.*`) is a separate hostname and is **not** behind Access; it is protected by JWT. Don't try to share the Access session with it.
- Never indexed: `public/_headers` sends `X-Robots-Tag: noindex, nofollow` on every path, and `public/robots.txt` disallows everything.
- This app's origins must be in the backend's `CORS_ALLOWED_ORIGINS` for each environment.

---

## Do / Don't

### Do
- Use antd for complex UI; Tailwind for layout.
- Mirror backend `apicode` permission codes in `src/constants/permissions.ts`.
- Treat client-side permission checks as UX only.
- Make the periods table fast to edit — it is used many times a day.
- Confirm destructive actions with a modal.
- Update `.env.example` when adding a variable.
- Run `npm run lint` and `npm run typecheck` before declaring a change complete.

### Don't
- **NEVER read, open, view, edit, or modify any `.env` file** — only `.env.example`.
- Don't store JWTs in `localStorage` without an explicit, documented decision.
- Don't call the Cloudflare Deploy Hook or any customer-site endpoint.
- Don't import from `frontend/customer`.
- Don't hardcode permission strings or error codes inline.
- Don't bypass the API client (no raw `fetch` in components).
- Don't introduce a global state library without a concrete need.
- Don't import Ant Design in code that could be shared with the customer site.

---

## Open Questions / TBD

Flag the user before assuming:

- **Refresh token flow** — none yet (access token in memory). Sliding session vs fixed expiry; HttpOnly cookie if added.
- **Server-state library** — TanStack Query once the real backend is stable; confirm before installing.
- **Form/validation library** — antd Form + zod is the working choice; align with the customer site's react-hook-form + zod.
- **Rich text editor** — needed if descriptions support formatting (TipTap, Lexical). Decide before the first editor is built.
- **Audit log UI** — who changed what; needs backend support first.
- **Multi-language editing** — `titleTh` / `titleEn` layout (side-by-side vs tabs).
- **Tour preview** — how staff preview a tour before it shows on the static site (dev site vs an in-app preview).
