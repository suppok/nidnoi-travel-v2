# CLAUDE.md — backend

Go REST API for nidnoi travel (`module backend`). One service serves both the customer site (public endpoints) and the backoffice (staff endpoints).

**Stack:** Go · Fiber v2 · Bun (Postgres via `pkg/core/database/query`) · PostgreSQL on Supabase · Casbin (RBAC) · JWT · Zap · ozzo-validation (via `pkg/core/validation`) · guregu/null · swaggo · Viper · aws-sdk-go-v2 (R2) · bun/migrate

> Keep the stack small. Redis, a message queue, a separate batch binary, gocron and OpenTelemetry are **not** in scope. Ask before adding any of them (see "Ask before adding").

---

## Entry points (`cmd/`)

| Binary        | Purpose                                                                                   | Run                |
|---------------|-------------------------------------------------------------------------------------------|--------------------|
| `cmd/api`     | REST API + in-process background workers (site-rebuild dispatcher, media orphan sweeper)  | `task run-api`     |
| `cmd/migrate` | Database migrations (bun/migrate, SQL files in `migrations/`)                             | `task migrate-up`  |

`cmd/api` has `app/app.go` (wires everything) and `handler/` (HTTP handlers).

The API runs as **exactly one instance** per environment. Background workers live in-process and are not coordinated across replicas. Do not scale horizontally until a real queue exists.

---

## Layered architecture — the mandatory flow

```
HTTP request
  └─▶ Middleware     CORS → logger → recover → (JWT → Casbin) per route
        └─▶ Handler        cmd/api/handler/<feature>_api_handler.go
              └─▶ UseCase interface   pkg/modules/<module>/<module>_usecase.go
                    └─▶ UseCase impl  pkg/modules/<module>/usecase/<module>_<action>_usecase.go
                          ├─▶ Shared services / clients   pkg/modules/shared/...   (via ShareRegistry)
                          └─▶ Repository   pkg/repository/<table>_repository.go     (via RepositoryRegistry)
                                └─▶ Domain entity   pkg/domain/<table>.go   (one struct = one table)
```

Dependency direction is strictly top-down. Never:
- call a repository from a handler,
- import `fiber` in a usecase or repository,
- import another module's usecase from a usecase (move shared logic into `pkg/modules/shared/services/`),
- put business logic in a repository (no ID generation, timestamps, audit fields, validation or permission checks),
- start, commit or roll back a transaction inside a repository.

---

## Folder organization

```
backend/
├── cmd/
│   ├── api/
│   │   ├── main.go
│   │   ├── app/app.go               # wiring
│   │   └── handler/                 # <feature>_api_handler.go
│   └── migrate/main.go
├── pkg/
│   ├── app/initializer/             # repository_initializer.go builds RepositoryRegistry
│   ├── constants/
│   │   ├── apicode/                 # <module>_api_code.go
│   │   └── errorcode/error_code.go
│   ├── core/                        # infrastructure only — no business rules
│   │   ├── api/                     # Fiber setup, ApiHandler.Do, response handler, middleware/
│   │   ├── auth/jwt/
│   │   ├── config/                  # config schema (Viper), validated at startup
│   │   ├── database/query/          # AppIDB, DB[T], DBTx
│   │   ├── errors/                  # cerrors
│   │   ├── logger/
│   │   ├── validation/              # wrapper over ozzo-validation (+ /is)
│   │   └── utils/                   # NewID, time helpers, ...
│   ├── domain/                      # <table>.go
│   │   ├── repomodels/              # <table>_<purpose>_repo_model.go
│   │   └── domainmodels/
│   ├── modules/
│   │   ├── <module>/
│   │   │   ├── <module>_usecase.go  # interface only
│   │   │   ├── models/              # <module>_<action>_model.go
│   │   │   ├── usecase/             # <module>_usecase.go + one file per action
│   │   │   └── commons/             # optional module-local helpers
│   │   └── shared/
│   │       ├── services/            # cross-module business services (revalidation, ...)
│   │       └── client/              # outbound clients (objectstorage, deployhook, ...)
│   ├── registry/                    # CoreRegistry, RepositoryRegistry, ShareRegistry
│   └── repository/                  # <table>_repository.go
├── migrations/                      # <timestamp>_<desc>.up.sql / .down.sql
├── config/
│   ├── config_api.yml
│   ├── local.env                    # OFF-LIMITS (secrets, gitignored)
│   ├── local.env.example
│   └── permissions/{model.conf,policy.csv}
├── docs/                            # swag output — generated, do not edit
├── Dockerfile
├── Taskfile.yml
└── go.mod
```

### Folder responsibilities

| Folder | Contains | Rules |
|---|---|---|
| `pkg/domain/` | Bun row structs, one file per table | `/* DB Table Details */` block + per-field column comment. Hand-written. Persistence fields only. |
| `pkg/domain/repomodels/` | List criteria and custom query result structs | `<Name>ListCriteria`, file `<table>_<purpose>_repo_model.go` |
| `pkg/repository/` | `<Name>Repository` interface + private impl, one file per table | First method is `WithTx(tx query.AppIDB)`. Add methods only when a use case needs them. |
| `pkg/modules/<module>/` | `<module>_usecase.go` = **interface only** | Methods are `(ctx, *models.XRequest) (*models.XResponse, error)` |
| `pkg/modules/<module>/usecase/` | `<module>_usecase.go` = struct + `New<Module>UseCase`; **one file per action** | Adding an action never touches the struct file |
| `pkg/modules/<module>/models/` | Request/Response DTOs, one file per action | Request carries `UserRequestInfo *coremodels.UserRequestInfo \`json:"-"\`` |
| `pkg/modules/shared/services/` | Business services used by two or more modules | Registered in `ShareRegistry.ServiceRegistry` |
| `pkg/modules/shared/client/` | Outbound clients (R2, Cloudflare deploy hook) | Registered in `ShareRegistry.ClientRegistry` |
| `pkg/registry/` | `CoreRegistry`, `RepositoryRegistry`, `ShareRegistry` | Add a field for every new repo/service/client |
| `pkg/app/initializer/` | `repository_initializer.go` | Register every new repository here |
| `pkg/constants/apicode/` | `ApiCode("module:action", "Description")` per endpoint | Used by Casbin; mirrored in backoffice `src/constants/permissions.ts` |
| `pkg/constants/errorcode/` | `ErrCode("CODE", messages...)` business error codes | UPPER_SNAKE, grouped by module |
| `pkg/core/` | Framework/infrastructure | Generic only — no business rules |
| `config/` | YAML config, env templates, Casbin model/policy | `local.env` is off-limits |

---

## Database access — `pkg/core/database/query`

A thin project layer over Bun. Keep the surface minimal and extend only when a use case needs it.

```go
type AppIDB = bun.IDB // *bun.DB or bun.Tx

type DB[T any] interface {
    Insert(ctx context.Context, obj *T) (*T, error)                    // RETURNING *
    Update(ctx context.Context, obj *T, columns ...string) (*T, error) // RETURNING *
    FindByID(ctx context.Context, id string) (*T, error)               // nil, nil when not found
    NewSelect() *bun.SelectQuery                                        // for custom queries
}

func New[T any](db AppIDB) DB[T]

type DBTx interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, appDB AppIDB) error) error
}
```

### Repository template

```go
type CountryRepository interface {
    WithTx(tx query.AppIDB) CountryRepository
    Create(ctx context.Context, obj *domain.Country) (*domain.Country, error)
    FindByID(ctx context.Context, id string) (*domain.Country, error)
    FindPublicBySlug(ctx context.Context, slug string) (*domain.Country, error)
}

type countryRepository struct{ query query.DB[domain.Country] }

func NewCountryRepository(db query.AppIDB) CountryRepository {
    return &countryRepository{query: query.New[domain.Country](db)}
}

func (r *countryRepository) WithTx(tx query.AppIDB) CountryRepository { return NewCountryRepository(tx) }
```

Rules:
- `WithTx` returns a new instance bound to the tx; the registry instance stays non-transactional.
- Custom queries use Bun's builder with bound args (`?`, `bun.Ident`, `bun.In`). Never string-concat user input into SQL.
- Not found → `nil, nil`. The use case maps it to a business error.
- Preload associations with `.Relation("Countries")`, `.Relation("Cities")`, `.Relation("Periods")`.
- Public repository methods hard-code `status = true`. It is never a caller parameter.

### Connection

Supabase is used as **managed Postgres only** (no Supabase Auth, Storage or Realtime).

- Connect through the **session pooler** (`aws-*.pooler.supabase.com:5432`, IPv4). Not the direct host (IPv6-only by default), not transaction mode (port 6543 — for serverless).
- Keep `MaxOpenConns` small (≈10): prd runs on Micro compute.

### Transactions

```go
err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
    repo := u.repositoryRegistry.TourProgramRepository.WithTx(appDB)
    // reads and writes that must succeed or fail together
    return nil // commit; return an error to roll back
})
```

- Every repository used inside the closure must be `.WithTx(appDB)`.
- Pure reads (public list/get) use the registry repository directly — no tx.
- Side effects (rebuild enqueue, notifications) happen **after** `WithTx` returns `nil`, never inside the closure.

---

## Usecase action template

Every action follows **Validate → (CheckPermissionData, only when needed) → Process**.

```go
func (u *tourProgramUseCase) CreateValidate(request *models.TourProgramCreateRequest) error {
    return validation.ValidateStruct(request,
        validation.Field(&request.TitleTh, validation.Required),
        validation.Field(&request.CountryIDs, validation.Required),
    )
}

func (u *tourProgramUseCase) Create(ctx context.Context, request *models.TourProgramCreateRequest) (*models.TourProgramCreateResponse, error) {
    if err := u.CreateValidate(request); err != nil {
        return nil, cerrors.Validation(err)
    }

    var created *domain.TourProgram
    err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
        repo := u.repositoryRegistry.TourProgramRepository.WithTx(appDB)

        code, err := repo.NextCode(ctx)
        if err != nil {
            return err
        }

        now := time.Now()
        created, err = repo.Create(ctx, &domain.TourProgram{
            ID:         utils.NewID(),
            Code:       code,
            TitleTh:    request.TitleTh,
            TitleEn:    request.TitleEn,
            Status:     true,
            CreateDate: now,
            UpdateDate: now,
            CreateBy:   request.UserRequestInfo.UserID,
            UpdateBy:   request.UserRequestInfo.UserID,
        })
        if err != nil {
            return err
        }

        return u.repositoryRegistry.TourProgramCountryRepository.WithTx(appDB).
            ReplaceForTour(ctx, created.ID, request.CountryIDs)
    })
    if err != nil {
        return nil, err
    }

    u.shareRegistry.ServiceRegistry.Revalidator.Enqueue(ctx,
        revalidation.PathsForTourMutation(created, revalidation.MutationCreate))

    return &models.TourProgramCreateResponse{ID: created.ID, Code: created.Code}, nil
}
```

Rules:
- Permission is enforced by Casbin at the route. Add `<Action>CheckPermissionData` only when access depends on the row itself (e.g. "sales may edit only their own tours"). None exist today.
- Field rules live only in `<Action>Validate`, using `pkg/core/validation`, not raw ozzo.
- Public use case methods return `cerrors.*` errors only — never raw `fmt.Errorf`. Use `errs.Wrap` for internal context.
- IDs: `utils.NewID()` (UUID v4). Timestamps and `CreateBy`/`UpdateBy` are set by the use case from `UserRequestInfo`.
- Logging: `u.coreRegistry.Logger`.

### Response DTOs

- One `<Module><Action>Response` per action. Concrete types only — no generics (swag cannot render them reliably).
- No `Success` field. `ApiResponseHandler` wraps every result as `{"success": true, "data": <response>}`.
- Lists: `<Module>ListResponse { Items []…; Total, Page, PageSize int }`.
- Public endpoints have their own DTOs (`<Module>Public<Action>Response`). Never expose `createBy`, `updateBy` or internal notes to the customer site.

---

## Errors

Response shape (cross-project contract — frontends depend on it):

```json
{ "success": false, "error": { "code": "TAXONOMY_IN_USE", "message": "...", "fields": { "slug": "..." } } }
```

- `pkg/core/errors` (`cerrors`) defines the generic errors and their HTTP status: `VALIDATION` (400, `fields` built from ozzo `validation.Errors`), `UNAUTHENTICATED` (401), `PERMISSION_DENIED` (403), `NOT_FOUND` (404), `CONFLICT` (409), `INTERNAL` (500).
- Business codes live in `pkg/constants/errorcode/error_code.go` as `ErrCode("TAXONOMY_IN_USE", "...")`. UPPER_SNAKE, grouped by module. Frontends branch on these strings — never rename a shipped code.
- `ApiResponseHandler` maps `*cerrors.AppError` to status + envelope. Any other error becomes `INTERNAL`: log the full error, return a generic message.

---

## Handler template

```go
func (h *TourProgramAPIHandler) TourProgramCreate(c *fiber.Ctx) error {
    request := &models.TourProgramCreateRequest{}
    return h.apiHandler.Do(c, request, func(ctx context.Context, requestInfo *coremodels.RequestInfo) (interface{}, error) {
        request.UserRequestInfo = requestInfo.UserRequestInfo
        return h.tourProgramUseCase.Create(ctx, request)
    })
}
```

Handlers are thin: bind, attach `UserRequestInfo` (always from the JWT, never from the body), call one use case method, return through `apiHandler.Do`. Never write `c.Status().JSON()`.

### Route registration

```go
func (h *TourProgramAPIHandler) Init(jwtAuth middleware.Auth, authPermissions middleware.AuthPermissions) {
    r := h.router

    // --- Public (customer site) — no auth. Register BEFORE "/:id" routes. ---
    r.Get("/tour-programs/public", h.TourProgramPublicList)
    r.Get("/tour-programs/public/:code", h.TourProgramPublicGet)
    r.Get("/tour-programs/public/:code/periods", middleware.CacheControl(30*time.Second), h.TourProgramPublicPeriods)

    // --- Staff (backoffice) ---
    r.Post("/tour-programs", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramCreate), h.TourProgramCreate)
    r.Put("/tour-programs", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramUpdate), h.TourProgramUpdate)
    r.Post("/tour-programs/list", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramList), h.TourProgramList)
    r.Put("/tour-programs/status", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramStatus), h.TourProgramUpdateStatus)
    r.Get("/tour-programs/:id", jwtAuth.Auth, authPermissions.RequiresPermissions(apicode.TourProgramGet), h.TourProgramGet)
}
```

Fiber matches routes in registration order: register `/x/public…` before `/x/:id`, or `public` is captured as an `:id`.

### Route conventions

All routes live under `/api/v1`. Resources are plural kebab-case; JSON bodies are camelCase in both directions.

| Method + path | Meaning |
|---|---|
| `POST /x` | create |
| `PUT /x` | update (id in body) |
| `POST /x/list` | list/search with a filter body |
| `GET /x/:id` | get one |
| `PUT /x/status` | soft delete / restore (`status` flag) |
| `GET /x/public…` | public read for the customer site — no auth, handler name contains `Public` |

Public routes are grouped at the top of `Init()` so it is obvious what is unauthenticated.

---

## Domain conventions

Every table has `id` (UUID), `status`, `create_date`, `update_date`, `create_by`, `update_by`.

- **`status bool`** — soft-delete flag only (`true` = active). Never carries workflow meaning.
- **`state int`** — workflow state (`10` draft, `20` selling, `30` cancelled, …; gaps left for new states). **Deferred.** Until it exists, every active tour is public.
- Nullable columns use `null.String` / `null.Time` / `null.Float` with `swaggertype`.
- `bun.BaseModel` always carries `json:"-" swaggerignore:"true"`.
- Relation fields have no column comment — they are not columns.

```go
/*
DB Table Details
-------------------------------------
Table: tour_program
[ 0] id           UUID         null: false  primary: true
[ 1] code         VARCHAR(20)  null: false  unique: true
[ 2] title_th     TEXT         null: false
[ 3] title_en     TEXT         null: false
[ 4] status       BOOLEAN      null: false  default: true
[ 5] create_date  TIMESTAMPTZ  null: false
[ 6] update_date  TIMESTAMPTZ  null: false
[ 7] create_by    UUID         null: false
[ 8] update_by    UUID         null: false
*/

// TourProgram is a row of the tour_program table.
type TourProgram struct {
    bun.BaseModel `bun:"table:tour_program,alias:tp" json:"-" swaggerignore:"true"`

    ID         string    `bun:"id,pk"       json:"id"`         // [0] UUID
    Code       string    `bun:"code,unique" json:"code"`       // [1] VARCHAR(20)
    TitleTh    string    `bun:"title_th"    json:"titleTh"`    // [2] TEXT
    TitleEn    string    `bun:"title_en"    json:"titleEn"`    // [3] TEXT
    Status     bool      `bun:"status"      json:"status"`     // [4] BOOLEAN
    CreateDate time.Time `bun:"create_date" json:"createDate"` // [5] TIMESTAMPTZ
    UpdateDate time.Time `bun:"update_date" json:"updateDate"` // [6] TIMESTAMPTZ
    CreateBy   string    `bun:"create_by"   json:"createBy"`   // [7] UUID
    UpdateBy   string    `bun:"update_by"   json:"updateBy"`   // [8] UUID

    Countries []Country           `bun:"m2m:tour_program_country,join:TourProgram=Country" json:"countries,omitempty"`
    Cities    []City              `bun:"m2m:tour_program_city,join:TourProgram=City"       json:"cities,omitempty"`
    Periods   []TourProgramPeriod `bun:"rel:has-many,join:id=tour_program_id"              json:"periods,omitempty"`
}
```

---

## Tour code

- `code` is the public, stable identifier used in customer URLs (`/tour/nidn261690/`). It never changes after create and is never accepted from a request body.
- Generated in the create use case from a Postgres sequence (`tour_code_seq`) inside the create transaction: `<prefix><nextval>`. A sequence is race-free; gaps after a rollback are acceptable. The prefix lives in `config_api.yml`.
- The legacy data import sets the sequence past the highest imported code with `setval`.
- `code` has a unique index. Internal references (foreign keys, audit) always use `id`.

---

## Tour taxonomy

Country and city — one module each (`pkg/modules/country`, `pkg/modules/city`). A tour has many countries and many cities (junctions `tour_program_country`, `tour_program_city`). A city belongs to one country (`country_id`).

Editable display fields: `name_th`, `name_en`, `description` (SEO copy for the landing page), `hero_image`, `sort_order`.

### Slug rules — strictly enforced

- `slug` appears in customer URLs (`/tours/japan/`, `/tours/japan/tokyo/`). Changing it breaks inbound links and indexed pages.
- Lowercase ASCII kebab-case, max 60 chars, `^[a-z0-9]+(-[a-z0-9]+)*$`, unique per table.
- **Locked after create.** Update DTOs have no `slug` field; the update use case never writes the column.
- Soft-deleting a country/city fails with `TAXONOMY_IN_USE` (including the referencing tour count) if any active tour references it.

> **Phase 2:** a `tag` dimension (`pkg/modules/tag`, junction `tour_program_tag`) following the city pattern. Do not pre-build hooks for it.

---

## Periods, prices and seats

`tour_program_period` holds each departure's dates, price and seat availability. This is the data that changes most often (many times a day); tour content rarely changes after it is posted.

- **Starting price** = the lowest price among a tour's active periods. It is baked into the customer site's static HTML (tour cards, JSON-LD), so a change requires a site rebuild.
- **Per-period price and seats** are served live to the customer site's period island via `GET /api/v1/tour-programs/public/:code/periods`: small payload (dates, price, seats left, sold-out flag), `Cache-Control: public, max-age=30`. Seat changes never trigger a rebuild.
- Booking is inquiry-based (no online seat reservation), so up to 30 seconds of staleness is acceptable.
- Staff edit periods through the `tour-program-periods` resource (standard route conventions; codes `tour-program-period:*`). The update use case compares the old and new row (and the tour's starting price) to decide whether to enqueue a rebuild — see the rebuild table.

---

## Customer site rebuild (Revalidator)

The customer site is fully static on Cloudflare Workers. When published content changes, the backend triggers a rebuild through a Cloudflare **Deploy Hook**. Frontends never call the hook.

```go
// pkg/modules/shared/services/revalidation/revalidation.go
type Revalidator interface {
    Enqueue(ctx context.Context, paths []string) // non-blocking; logs failures, never returns an error
}
```

`DeployHookRevalidator` (in-process):
- **Debounce:** fires once no new enqueue has arrived for `quiet` (default 3 min), and at most `maxWait` (default 15 min) after the first pending enqueue. Both values live in `config_api.yml`.
- Fires `POST <CLOUDFLARE_DEPLOY_HOOK_URL>` through `shared/client/deployhook`. Retries 3× with backoff (1 s, 5 s, 30 s); a final failure is logged at ERROR with the pending paths.
- Cloudflare skips duplicate triggers while a build is still queued, and rate-limits hooks to 10 builds/minute per Worker.
- `paths` are only logged today. They are kept so a per-path invalidation strategy can return later without touching use cases.
- Use cases enqueue **after** the transaction commits, using `revalidation.PathsForTourMutation` / `PathsForTaxonomyMutation` — never hand-rolled path lists.

When to enqueue:

| Mutation | Enqueue? |
|---|---|
| Tour create / content update / status change | yes |
| Tour country/city assignment change | yes |
| Period create / delete / date change | yes |
| Period price change | only if the tour's starting price changes |
| Period seat change | **never** |
| Country/city display-field update | yes |
| Media upload not yet referenced by a record | no |

Manual trigger: `POST /api/v1/site/rebuild` (`apicode.SiteRebuild`, code `site:rebuild`) fires immediately and bypasses the debounce.

---

## Object storage (Cloudflare R2)

The backend is the only holder of R2 credentials. Upload flow: backoffice → `POST /api/v1/media/upload` (multipart) → validate → re-encode → R2 → `{ url, key, size, mime }`.

- **Client:** `pkg/modules/shared/client/objectstorage` exposes an `ObjectStorage` interface (`Put`, `Delete`, `PublicURL`); the R2 implementation uses aws-sdk-go-v2 with a custom endpoint. Use cases depend on the interface only.
- **Form fields:** `file`, `purpose` (`tour-cover | tour-gallery | country-hero | city-hero`), `refId` (optional).
- **Validation:** `image/jpeg`, `image/png`, `image/webp` only (no SVG/GIF). Max 10 MB, checked server-side. Minimum dimensions: tour cover ≥ 1200×800, hero ≥ 1600×900. Read dimensions from the header first and reject anything over 50 MP before decoding.
- **Processing:** decode → re-encode as WebP (quality 80–85, original dimensions). Re-encoding also strips EXIF. Limit concurrent decodes with a semaphore (e.g. 2) — a 24 MP image takes roughly 100 MB in memory.
- **Fiber `BodyLimit`** must be at least 12 MB (Fiber's default is 4 MB).
- **Keys are immutable:** `tours/<code>/cover-<hash>.webp`, `tours/<code>/gallery/<n>-<hash>.webp`, `countries/<slug>/hero-<hash>.webp`, `cities/<slug>/hero-<hash>.webp`. Replacing an image always creates a new key.
- **Deletion is soft:** replaced or removed keys go into `media_orphan` with `delete_after = now() + 7 days`; an in-process sweeper (`time.Ticker`) deletes expired keys from R2.
- **Permission:** `media:upload`.

---

## Public endpoints (customer site)

All unauthenticated. All filter `status = true` in the repository. All return 404 for a missing **or** soft-deleted record — the same response, so deleted records never leak.

| Endpoint | Caller |
|---|---|
| `GET /tour-programs/public?country=&city=&dateFrom=&dateTo=&page=&pageSize=` | site build |
| `GET /tour-programs/public/:code` | site build |
| `GET /tour-programs/public/:code/periods` | browser (period island), cached 30 s |
| `GET /countries/public`, `GET /countries/public/:slug`, `GET /countries/public/:slug/cities` | site build |
| `GET /cities/public`, `GET /cities/public/:slug` | site build |
| `POST /inquiries/public` | browser (inquiry form), rate-limited |

- Country/city detail returns 404 when the slug has zero active tours (empty landing pages hurt SEO).
- Build-time callers should need few requests. If site builds slow down, add a bulk endpoint rather than N per-tour calls.
- Browser-called endpoints need the customer origin in CORS. Rate-limit `POST /inquiries/public` with Fiber's built-in `limiter` middleware.

---

## Auth and permissions

- **Users have one or more roles** — fixed set `super_admin`, `admin`, `sales`, `finance` (`pkg/constants/role`), stored as `app_user.roles TEXT[]`. CHECK constraints enforce "at least one" and "known roles only" in the database itself. Login is by email (stored lowercase, unique). A request is allowed when **any** of the user's roles allows it (`pkg/core/auth/permission` enforces Casbin per role). Permissions returned by `/auth/me` are the union.
- The access token (HS256 JWT) carries **only the user id**. `jwtAuth.Auth` loads the user from the database on every request (`AuthUseCase.ResolveSession`, one primary-key lookup) to build `UserRequestInfo` (id, email, name, roles, mustChangePassword). So deactivation, role changes and a completed password change apply on the next request, not after the token expires.
- **The first user is the super admin.** `task create-super-admin -- <email> <name>` (`cmd/migrate create-super-admin`) creates them with role `super_admin` and prints a random temporary password once; it must be changed at first login. The command refuses if a super admin already exists, and a partial unique index allows at most one.
- `super_admin` has every permission (`p, super_admin, *, *`). It can't be granted or removed through the API (`role.Assignable` excludes it), and nobody else can edit, deactivate or reset the super admin (`SUPER_ADMIN_PROTECTED`). The super admin edits their own name/email via `PUT /users` and their password via `/auth/password`. Because the super admin always exists and stays active, user management can't lock itself out (this replaces the earlier `LAST_ADMIN` rule). Recovering a lost super-admin password needs database access — there is no API path.
- Errors the backoffice branches on: `INVALID_CREDENTIALS` (401, same for unknown email / wrong password / deactivated), `RATE_LIMITED` (429), `INVALID_PASSWORD`, `PASSWORD_CHANGE_REQUIRED` (403), `EMAIL_TAKEN` (409), `CANNOT_MODIFY_SELF` (409), `SUPER_ADMIN_PROTECTED` (409), `SUPER_ADMIN_REQUIRED` (403). Mirrored in `frontend/backoffice/src/constants/errorCodes.ts`.

### Auth endpoints (contract implemented by the backoffice mock, `frontend/backoffice/src/api/mock/`)

| Endpoint | Notes |
|---|---|
| `POST /auth/login` `{email, password, rememberMe}` | Rate-limited. → `{accessToken, expiresIn}`; sets the refresh cookie (persistent if `rememberMe`, else session cookie). Updates `lastLoginAt`. |
| `POST /auth/refresh` | Reads the refresh cookie, **rotates** it, → `{accessToken, expiresIn}`. 401 when missing/revoked/expired. |
| `POST /auth/logout` | Revokes the refresh token, clears the cookie. No access token required. |
| `GET /auth/me` | → `{id, email, name, roles, permissions, mustChangePassword}` |
| `PUT /auth/password` `{currentPassword, newPassword}` | Clears `mustChangePassword`. |

- Lifetimes live in `config_api.yml` (`auth.*`): access 15 min; refresh 30 days with `rememberMe` (persistent cookie), 12 h without (session cookie, server-side expiry). Both slide on every refresh.
- Refresh tokens are 256-bit random values; table `refresh_token` stores only their SHA-256. Every refresh **rotates** the token (old row: `revoked_at`, `replaced_by`), under `SELECT … FOR UPDATE`. A rotated token presented again within `refreshReuseGrace` (30 s, two tabs racing) just gets 401. After that, it counts as theft and **every session of that user is revoked**.
- Refresh cookie `nidnoi_refresh`: `HttpOnly; SameSite=Lax; Path=/api/v1/auth`, `Secure` except on `APP_ENV=local`, no `Domain` (host-only on `api.*`). `admin.*` → `api.*` is same-site, so CORS sends `Access-Control-Allow-Credentials: true` with the explicit origin list (`*` is rejected at startup). A failed refresh and logout clear the cookie.
- While `mustChangePassword` is true, every staff route returns 403 `PASSWORD_CHANGE_REQUIRED`. The exceptions are `GET /auth/me` and `PUT /auth/password` (registered with `jwtAuth.AuthAllowPasswordChange`) and the cookie-only `POST /auth/logout`. Changing the password revokes every other session and keeps the current one.
- Login: one `INVALID_CREDENTIALS` for unknown email, deactivated user and wrong password, and bcrypt runs in every case (no timing oracle). Rate-limited per client IP (`rateLimit.login`; IP from `server.clientIPHeader` = `CF-Connecting-IP`, falling back to the socket).
- Password rule (`pkg/modules/shared/services/password`, mirrored in `frontend/backoffice/src/lib/password.ts`): at least 8 characters, at most 72 bytes, at least one ASCII letter and one digit. bcrypt cost `auth.bcryptCost` (12).

### User endpoints

`POST /users/list`, `GET /users/:id`, `POST /users` (`temporaryPassword` required, sets `mustChangePassword`), `PUT /users` (`id, email, name, roles`), `PUT /users/status`, `PUT /users/password` (admin reset: new temporary password, sets `mustChangePassword`). Permissions `user:create|update|list|get|status|password-reset`.

- Deactivating a user or resetting their password **revokes all their refresh tokens**. Their access token stops working on the next request because the middleware reloads the user. Role changes need no revocation for the same reason.
- `CANNOT_MODIFY_SELF`: staff can't change their own roles or status, or reset their own password through `/users/password` (they use `/auth/password`).
- `SUPER_ADMIN_PROTECTED`: any change to the super admin by another user. Sending `super_admin` in `roles` is a `VALIDATION` error.
- `SUPER_ADMIN_REQUIRED` (403): only the super admin may grant the `admin` role (create or update) or edit, deactivate or reset an admin. An admin manages `sales`/`finance` users only — otherwise an admin could reset a fellow admin's password and sign in as them. Checked in the use case (`ensureCanGrant`, `ensureCanManage`) because it depends on the row.
- Permission codes live in `pkg/constants/apicode/<module>_api_code.go`, format `module:action` with a kebab-case module: `tour-program:create`, `tour-program:status`, `tour-program-period:update`, `country:update`, `media:upload`, `site:rebuild`, …
- Casbin policy: `config/permissions/policy.csv`, one line per role per code, e.g. `p, sales, tour-program, create`.
- The backoffice mirrors every code in `frontend/backoffice/src/constants/permissions.ts` — update both in the same PR.
- Client-side checks are UX only; the route middleware is the source of truth.

### Middleware order

- Global (in `app.go`): CORS → request logger → recover.
- Per staff route: `jwtAuth.Auth` → `authPermissions.RequiresPermissions(apicode.X)`. Auth always runs before the permission check.
- Public routes: neither. Optional `CacheControl` and `limiter`.

---

## Registries

| Registry | Contains |
|---|---|
| `CoreRegistry` | Logger, Config, `DBTx`, `AppDB`, JWT, Permissions (Casbin checker) |
| `RepositoryRegistry` | All repositories (built in `pkg/app/initializer/repository_initializer.go`) |
| `ShareRegistry.ServiceRegistry` | `Revalidator` and other cross-module services |
| `ShareRegistry.ClientRegistry` | `ObjectStorage`, `DeployHook` |

No global variables, no `init()` wiring. Background workers (revalidator, media sweeper) are started from `app.go` with a context that is cancelled on shutdown.

---

## Configuration and environments

Environments: `local`, `dev`, `prd` (`APP_ENV`). Each deployed environment has its own Supabase project, R2 bucket, deploy hook and secrets. Never share a value across environments. Never copy prd data into dev — inquiries contain customer personal data (PDPA).

| File | Purpose |
|---|---|
| `config/config_api.yml` | Non-secret settings: timeouts, debounce, tour code prefix, upload limits |
| `config/local.env` | Local secrets — **off-limits**, gitignored |
| `config/local.env.example` | Template with placeholders — safe to read and edit |
| `pkg/core/config/` | Go schema; validated at startup, fails fast on missing values |

Deployed environments read variables from Railway (set per environment; prd secrets sealed).

| Env var | Notes |
|---|---|
| `APP_ENV` | `local` / `dev` / `prd` |
| `PORT` | Provided by Railway |
| `DATABASE_URL` | Supabase **session pooler** URL |
| `JWT_SECRET` | Different per environment |
| `CORS_ALLOWED_ORIGINS` | Comma-separated, see below |
| `CLOUDFLARE_DEPLOY_HOOK_URL` | The URL itself is the credential — treat as a secret |
| `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY` | Token scoped to this environment's bucket |
| `R2_BUCKET` | `nidnoi-travel-media-dev` / `nidnoi-travel-media-prod` |
| `R2_PUBLIC_URL` | `https://dev-media.nidnoitravelv2.com` / `https://media.nidnoitravelv2.com` |

CORS origins per environment:

| Env | Origins |
|---|---|
| local | `http://localhost:4321`, `http://localhost:5173` |
| dev | `https://dev.nidnoitravelv2.com`, `https://dev-admin.nidnoitravelv2.com` |
| prd | `https://www.nidnoitravelv2.com`, `https://admin.nidnoitravelv2.com` |

> **CRITICAL — secrets protection:**
> **NEVER read, view, open, cat, grep, edit, or modify any `.env` / `*.env` file** (`config/local.env` or any variant). If a value is needed, **stop and ask the user**. Only `*.example` files are safe.
> When adding a variable: update `config/local.env.example` (placeholder), the schema in `pkg/core/config/`, the table above, then tell the user to set the real value locally and in Railway themselves.

---

## Deployment

- **Railway**, Singapore region (same region as Supabase). Project `nidnoi-travel` with environments `dev` (branch `develop`) and `production` (branch `main`). Root directory `backend/`, watch paths `backend/**`.
- One always-on replica per environment. Leave ~1 GB memory headroom for image processing.
- Migrations run as a separate step before deploy — never on API startup. **Local and dev share one Supabase database**, so dev is migrated by hand from a local machine (`go run ./cmd/migrate up` with the dev `DATABASE_URL`) before pushing to `develop`. The image also ships `/app/migrate` (usable as a Railway pre-deploy command for prd).
- The `Dockerfile` is kept so the service can move to Fly.io or a VPS without code changes.
- Swagger UI (`/swagger/index.html`) is mounted only when `APP_ENV != prd`.

---

## Migrations

- `migrations/<YYYYMMDDHHMMSS>_<description>.tx.up.sql` + `.tx.down.sql`, run by `cmd/migrate` (bun/migrate). The `.tx` infix runs each file in a transaction.
- One concern per migration. Never edit a migration that has already run in dev or prd — add a new one.
- **Local = dev database.** Running a migration locally applies it to dev immediately, while dev still runs the previous code until the next deploy. Keep migrations backward compatible (add first; drop or rename only after the code that stops using it is deployed). Try risky or draft migrations against a throwaway Postgres (e.g. Docker) first, and avoid `migrate down` on the shared database. Local testing writes into dev data — fine, since dev never holds prd data.
- `task migrate-create -- <description>` creates the pair; `task migrate-status` lists applied/pending.

---

## Commands

```sh
cd backend
task run-api                       # go run ./cmd/api with config/local.env
task migrate-up                    # also: migrate-down, migrate-status, migrate-create -- <desc>
task create-super-admin -- <email> <name> # first user (super admin); prints a temporary password once
task swag                          # swag init -g cmd/api/main.go --parseDependency --parseInternal
task lint                          # golangci-lint
go build ./...
go vet ./...
go test ./pkg/...                  # add *_test.go next to the code under test
```

Always run from `backend/` — `go.mod` lives there. In GoLand, set the run configuration's working directory to `backend\`.

---

## Adding a new feature — checklist

Example: module `inquiry`, action `create`, table `inquiry`.

1. **Migration** — `migrations/<ts>_create_inquiry.up.sql` / `.down.sql`.
2. **Domain** — `pkg/domain/inquiry.go` with the table comment block.
3. **Repo model** (if custom query) — `pkg/domain/repomodels/inquiry_list_repo_model.go`.
4. **Repository** — `pkg/repository/inquiry_repository.go` (`WithTx` first; only the methods you need).
5. **Register repo** — field in `RepositoryRegistry` + constructor call in `repository_initializer.go`.
6. **Models** — `pkg/modules/inquiry/models/inquiry_create_model.go`.
7. **UseCase interface** — `pkg/modules/inquiry/inquiry_usecase.go`.
8. **UseCase impl** — `usecase/inquiry_usecase.go` (struct + `NewInquiryUseCase(coreRegistry, shareRegistry, repositoryRegistry)`) and `usecase/inquiry_create_usecase.go`.
9. **API code** — `pkg/constants/apicode/inquiry_api_code.go`; business error codes in `errorcode/error_code.go`.
10. **Handler** — `cmd/api/handler/inquiry_api_handler.go` with `NewInquiryAPIHandler(...)` and `Init(...)`.
11. **Wire** — construct the use case and handler in `cmd/api/app/app.go`, next to related modules.
12. **Permissions** — `policy.csv` lines for staff routes; mirror the codes in backoffice `permissions.ts`.
13. **Swagger** — annotations on every handler method, then `task swag`.
14. **Customer site impact** — if the data is rendered on the customer site, enqueue a rebuild after commit (see the rebuild table).
15. **Verify** — `go build ./...`, `go vet ./...`, `task lint`.

## Modifying an existing domain — checklist

When a field is added, removed or renamed, propagate it through every layer:

1. Migration.
2. `pkg/domain/<entity>.go` — struct field and comment block.
3. Repository queries — `Relation`, `Where`, `Order`, explicit column lists.
4. Request/response DTOs, including public DTOs.
5. Use cases — struct literals, validation, business logic.
6. Swagger annotations, then `task swag`.
7. `revalidation.Paths…` helpers and the rebuild table, if the field affects what the customer site renders.
8. `go build ./...`.

---

## Naming conventions

- **Files:** `snake_case` with a layer suffix: `_api_handler.go`, `_usecase.go`, `_model.go`, `_repository.go`, `_repo_model.go`, `_service.go`, `_client.go`.
- **Packages:** lowercase, no underscores (`tourprogram`, `country`, `city`, `media`, `inquiry`).
- **Types:** exported interface (`TourProgramUseCase`), unexported impl (`tourProgramUseCase`), `NewXxx` returns the interface.
- **DTOs:** `<Module><Action>Request` / `<Module><Action>Response`; public: `<Module>Public<Action>Response`.
- **Import aliases** on name clashes: `coremodels`, `cerrors`, `errs`, `<module>usecase`.
- All code, comments and identifiers in English. Keep comments sparse and match the surrounding style.

---

## Ask before adding

- Redis (cache, locks, sessions)
- A message queue (RabbitMQ/Redis) and a separate `cmd/batch` binary — required before running more than one API replica
- gocron or any scheduler library (an in-process `time.Ticker` is enough today)
- OpenTelemetry, HMAC partner auth, error notifier (Discord/email)
- Server-side image variants (thumbnails, fixed-size OG images)
- Any new Go dependency

---

## Open questions

- **WebP encoder** — `disintegration/imaging` decodes but cannot encode WebP. Choose an encoder (cgo-based `chai2010/webp` vs a pure-Go/wasm option) and confirm it builds on Railway and on Windows (cgo needs gcc there).
- **Sales/finance permissions** — rows in `policy.csv` are placeholders; confirm with the business before prd.
- **golangci-lint** — the locally installed binary was built with Go 1.25 and refuses this Go 1.26 module; reinstall it with Go 1.26 before relying on `task lint`.
- **Inquiry spam protection** — rate limiting only for now; consider Cloudflare Turnstile before launch.
- **Legacy data import** — preserve legacy `code` values and `setval` the tour code sequence past the highest one.
