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
    DealSection.tsx    — price table (label · จำนวน · ราคา/หน่วย) like the backoffice tool, + เพิ่มรายการ, discount
    PaymentSection.tsx — transfer/card, full/deposit, due dates, totals
    fields.tsx         — small shared inputs (NumberInput, Toggle, Label)
    Select.tsx         — styled dropdown (native <select> popups can't be styled); use it instead of <select>
    TourSelector.tsx   — search tours via public API
    PeriodSelector.tsx — load periods for selected tour
    SlipUpload.tsx     — upload slip image to /media/upload/liff
    SuccessView.tsx    — booking confirmed, show booking_no
    SummaryPaste.tsx   — paste a backoffice "สรุปยอดโอน" text to prefill the form
  lib/parseSummary.ts  — parser for that text
  lib/deal.ts          — deal state, totals and defaults (mirrors summary.ts and the backend quote)
  config/env.ts        — VITE_APP_ENV, VITE_API_BASE_URL, VITE_LIFF_ID
  hooks/useLiff.ts     — LIFF init + profile hook
  types/index.ts       — shared type definitions
  App.tsx              — LIFF init, show form or error
  main.tsx             — entry point
```

## The booking deal

The form holds everything the backoffice **สรุปยอดโอน** summary holds, and the backend stores it on `line_booking`: booking source (+ the customer's name and booking no on it), travellers including infants, agreed unit prices, price before discount, discount/person, extra charges (visa, foreign passport, join land, fuel/airport tax — an empty tax count follows the traveller count — and named options), transfer or card (fee %), full or deposit with due dates/times, and the customer's address.

- Picking a period fills the adult, child (`childWithBedPrice`, else adult) and single-room (`singleSupplement`) prices, the before-discount price, a deposit when departure is more than 30 days away and the period has one, and the balance due date 30 days before departure (`applyPeriod`). Everything stays editable.
- Totals (`computeTotals` in `lib/deal.ts`) use the same math as `summary.ts` and the backend `quoteLine`; the backend recomputes the total, so keep the three in sync. The card fee is rounded up in basis points.
- **Rows are the summary tool's** (ผู้ใหญ่, เด็กอายุต่ำกว่า 20 ปี, ทารก, วีซ่า, พาสปอร์ตต่างชาติ, ห้องพักเดี่ยว (ห้อง), Join land, ภาษีน้ำมัน, ภาษีสนามบิน, then added rows). `dealRequest` maps them to the backend's room types: single rooms → adults in single rooms (at most the adult count; supplement = the room price), the other adults → twin, children → "with bed" (child no bed is not used). An empty fuel/airport tax count shows the traveller count as its placeholder and follows it.

## Prefill from a booking summary

Sales can paste the text from the backoffice sales tool **สรุปยอดโอน** (`frontend/backoffice/src/features/sales-tool/booking-summary/summary.ts`) into the app. **The first screen is only a paste box and "ข้าม กรอกเอง" (skip)**: a paste (a change of 20+ characters at once) is parsed and applied immediately, then the form opens prefilled; skip opens it empty. Typed text is applied only with the "กรอกข้อมูลจากข้อความ" button. A paste right after opening waits for the sales users and booking sources to load. Inside the form, "วางข้อความสรุปยอดโอนใหม่" pastes another summary the same way; "บันทึกรายการใหม่" after a booking returns to the paste screen. `parseSummary` reads its `• <label> : <value>` lines — the two files are a pair: change a label in one, update the other (no shared code between projects).

- Filled: sales owner (matched by name; the text writes "<name> <phone>"), booking source (by exact name) with its reference name and booking no, customer name (without "คุณ"), phone and address, the whole deal (`dealFromSummary`: counts, prices, charges, options, discount, payment, due dates/times), tour (searched by `รหัสทัวร์`, or by exact name for older texts) and period (matched by dates once the periods load; only open periods with seats). The auto-selected period does **not** overwrite the pasted prices.
- The pasted rows map one to one (`dealFromSummary`); adults = travellers − children − infants.
- Payment sentences ("กรุณาชำระเงินมัดจำ…", "…ส่วนเหลือ…", "…เต็มจำนวน…") are read separately from the `•` label lines. When the form's total differs from the summary's "รวมเป็นเงิน", a hint shows under the payment box.
- Nothing is submitted automatically; the slip is still uploaded by hand. The report under the paste box lists what was filled and what to check.

## Environment Variables

All `VITE_*` — shipped to the browser, no secrets.

| Variable | Example | Notes |
|---|---|---|
| `VITE_APP_ENV` | `local` / `dev` / `prd` | |
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` | Backend API base |
| `VITE_LIFF_ID` | `2006......-xxxxxxxx` | From LINE Developers console |
| `VITE_LIFF_MOCK` | `false` | `true` = local preview without LINE; startup fails unless `VITE_APP_ENV=local` |

### Local preview without LINE (`VITE_LIFF_MOCK=true`)

LIFF login only works on the HTTPS endpoint registered for a LIFF app, so plain `localhost` can't log in. With the mock flag, `useLiff` skips `liff.init` and `client.ts` lazy-loads `src/api/mock.ts`, which fakes only the endpoints that need a LIFF token (`/liff/me`, `/liff/line-link`, `/liff/sales-users`, `/liff/booking-sources` (LINE OA, Facebook, Walk-in), `/liff/line-contacts/list`, `/media/upload/liff` → `blob:` URL, `POST /line-bookings` → random booking no, logged to the console). **Public tour/period endpoints still hit the real backend**, so run `task run-api` too. Mock sales users have the backoffice mock names (สมใจ ขายเก่ง, …). For real LINE login, tunnel the dev server (ngrok) to a separate local LIFF app.

## Auth Flow

1. LIFF SDK initializes with the LIFF ID
2. If the user isn't logged in, `liff.login()` redirects to LINE Login
3. After login, the app gets the user's profile (userId, displayName) and access token
4. The access token is sent as `Authorization: Bearer <LIFF access token>` on every call (`setLiffToken` in `api/client.ts`). The backend verifies it with LINE, then maps the LINE account to the **backoffice user who linked it**.
5. **Only linked staff can use the app.** `useStaff` calls `GET /liff/me` first; it needs `line-booking:create` (sales, admin, super admin). Unlinked LINE accounts (`LINE_NOT_LINKED`), finance (`PERMISSION_DENIED`) and users with a temporary password (`PASSWORD_CHANGE_REQUIRED`) see a screen explaining what to do. The LIFF URL itself is public — the API is what is protected.
6. **Linking:** in the backoffice, user menu → "เชื่อมบัญชี LINE" shows a QR code for `https://liff.line.me/<LIFF ID>?lineLink=<code>` (one-time, 10 minutes). Opening it runs `POST /liff/line-link {code}` before `/liff/me`; the code is read from `lineLink` or from inside `liff.state`, then removed from the URL.
7. No JWT, no cookies — the LIFF access token is the only credential

## API Endpoints Used

- `GET /tour-programs/public?keyword=` — search tours by name or code (public, no auth)
- `GET /tour-programs/public/:code/periods` — get periods (public, no auth; `{items}`, hide `soldOut`)
- `POST /liff/line-link` — finish linking with the QR code's code
- `GET /liff/me` — the linked staff user (gate for the whole app)
- `GET /liff/booking-sources` — "ช่องทางการจอง" select (optional; sent as `bookingSourceId`)
- `GET /liff/sales-users` — "เซลล์เจ้าของการจอง" select (required; the submitter is preselected when they are sales), sent as `salesUserId`
- `POST /liff/line-contacts/list` — customer chat picker (`CustomerChatPicker`, optional; sent as `lineContactId`)
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
