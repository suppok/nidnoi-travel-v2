# nidnoi travel — Tour Program Platform

Customer site (`www.nidnoitravelv2.com`), staff backoffice (`admin.nidnoitravelv2.com`) and the Go API (`api.nidnoitravelv2.com`).

| Path | What | Stack |
|---|---|---|
| `backend/` | REST API | Go · Fiber · Bun · Supabase Postgres |
| `frontend/customer/` | Public site | Astro (static) · React islands · Tailwind |
| `frontend/backoffice/` | Staff admin | React · Vite · Ant Design · Tailwind |

Read `CLAUDE.md` (root) and the `CLAUDE.md` in each subproject before working there.

## Local setup

Prerequisites: Go, Node 22, [Task](https://taskfile.dev).

```bash
# Backend — copy config/local.env.example to config/local.env and fill in values
cd backend
task migrate-up
task run-api            # http://localhost:8080

# Customer site — copy .env.example to .env and fill in values
cd frontend/customer
npm install && npm run dev   # http://localhost:4321

# Backoffice — copy .env.example to .env and fill in values
cd frontend/backoffice
npm install && npm run dev   # http://localhost:5173
```

## Deploy map

| Component | dev (`develop`) | prd (`main`) |
|---|---|---|
| Customer site | Worker `customer-dev` → `dev.nidnoitravelv2.com` | Worker `customer-prd` → `www.nidnoitravelv2.com` |
| Backoffice | Worker `backoffice-dev` → `dev-admin.nidnoitravelv2.com` | Worker `backoffice-prd` → `admin.nidnoitravelv2.com` |
| API | Railway `dev` → `dev-api.nidnoitravelv2.com` | Railway `production` → `api.nidnoitravelv2.com` |
| Media | R2 `nidnoi-travel-media-dev` → `dev-media.nidnoitravelv2.com` | R2 `nidnoi-travel-media-prod` → `media.nidnoitravelv2.com` |

Release: merge `develop` → `main`.
