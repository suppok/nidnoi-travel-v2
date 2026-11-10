# Frontend: LIFF Booking App

A LINE Front-end Framework (LIFF) app that sales staff use inside LINE OA chat to submit bookings with payment slips. Opens as a full-screen web view inside the LINE app.

## Stack

- **Vite + React 19 + TypeScript** (same toolchain as backoffice)
- **Tailwind CSS v4** for styling
- **@line/liff** SDK for LINE integration
- Deployed to **Cloudflare Workers** (static assets, SPA fallback)

## Project Structure

```
src/
  api/client.ts       — fetch wrapper (no JWT; sends LIFF access token)
  components/
    BookingForm.tsx    — main form orchestration
    TourSelector.tsx   — search tours via public API
    PeriodSelector.tsx — load periods for selected tour
    SlipUpload.tsx     — upload slip image to /media/upload/liff
    SuccessView.tsx    — booking confirmed, show booking_no
  config/env.ts        — VITE_APP_ENV, VITE_API_BASE_URL, VITE_LIFF_ID
  hooks/useLiff.ts     — LIFF init + profile hook
  types/index.ts       — shared type definitions
  App.tsx              — LIFF init, show form or error
  main.tsx             — entry point
```

## Environment Variables

All `VITE_*` — shipped to the browser, no secrets.

| Variable | Example | Notes |
|---|---|---|
| `VITE_APP_ENV` | `local` / `dev` / `prd` | |
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` | Backend API base |
| `VITE_LIFF_ID` | `2006......-xxxxxxxx` | From LINE Developers console |

## Auth Flow

1. LIFF SDK initializes with the LIFF ID
2. If the user isn't logged in, `liff.login()` redirects to LINE Login
3. After login, the app gets the user's profile (userId, displayName) and access token
4. The access token is sent as `Authorization: Bearer <LIFF access token>` on every call (`setLiffToken` in `api/client.ts`). The backend verifies it with LINE, then maps the LINE account to the **backoffice user who linked it**.
5. **Only linked staff can use the app.** `useStaff` calls `GET /liff/me` first; it needs `line-booking:create` (sales, admin, super admin). Unlinked LINE accounts (`LINE_NOT_LINKED`), finance (`PERMISSION_DENIED`) and users with a temporary password (`PASSWORD_CHANGE_REQUIRED`) see a screen explaining what to do. The LIFF URL itself is public — the API is what is protected.
6. **Linking:** in the backoffice, user menu → "เชื่อมบัญชี LINE" shows a QR code for `https://liff.line.me/<LIFF ID>?lineLink=<code>` (one-time, 10 minutes). Opening it runs `POST /liff/line-link {code}` before `/liff/me`; the code is read from `lineLink` or from inside `liff.state`, then removed from the URL.
7. No JWT, no cookies — the LIFF access token is the only credential

## API Endpoints Used

- `POST /tour-programs/public/list` — search tours (public, no auth)
- `GET /tour-programs/public/:code/periods` — get periods (public, no auth)
- `POST /liff/line-link` — finish linking with the QR code's code
- `GET /liff/me` — the linked staff user (gate for the whole app)
- `POST /media/upload/liff` — upload slip image (`purpose=booking-slip` only)
- `POST /line-bookings` — create booking; the backend records the staff user as `create_by`

## CORS

The LIFF app is served from `dev-liff.nidnoitravelv2.com` (dev) and `liff.nidnoitravelv2.com` (prd). These must be added to the backend's `CORS_ALLOWED_ORIGINS`.

## Commands

```bash
npm install
npm run dev         # Vite dev server (localhost:5174)
npm run build       # production build
npm run lint        # oxlint
npm run typecheck   # tsc -b
```

## Cloudflare Workers

Workers names: `liff-dev` (dev), `liff-prd` (prd). Custom domains: `dev-liff.nidnoitravelv2.com`, `liff.nidnoitravelv2.com`.

## LINE Setup

Two channels are needed in the LINE Developers console:

1. **Messaging API channel** — for future webhook/push messaging (not used by this app)
2. **LINE Login channel** — hosts the LIFF app; set size to "Full", scope to "profile"

The LIFF app endpoint URL points to the deployed Workers URL.
