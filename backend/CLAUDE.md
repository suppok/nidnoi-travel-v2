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
- Keep the pool small (`database.*` in `config_api.yml`: 5 open, 2 idle, idle conns closed after 5 min). Session mode gives every client connection its own pooler slot (`pool_size` 15 on the Free org) and **local + dev share one database**, so two API instances plus `migrate` must fit; exceeding it fails requests with `EMAXCONNSESSION`.

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
- **`state int`** — workflow state with gaps left for new states, one constants package per table (`pkg/constants/wholesalestate`, `pkg/constants/tourprogramstate`). Tour programs: `10` draft, `20` selling, `30` closed, `40` cancelled; the customer site shows `tourprogramstate.Public` (20 and 30) only.
- Nullable columns use `null.String` / `null.Time` / `null.Float` with `swaggertype`.
- Strings are `TEXT`, never `VARCHAR(n)` (same storage and speed in Postgres). Length limits live in the use case's `<Action>Validate` (`RuneLength`, which counts characters like Postgres `char_length`). When the database must also guarantee a limit (e.g. before a bulk import), add a `CHECK (char_length(col) <= n)` constraint in a new migration.
- `bun.BaseModel` always carries `json:"-" swaggerignore:"true"`.
- Relation fields have no column comment — they are not columns.

```go
/*
DB Table Details
-------------------------------------
Table: tour_program
[ 0] id           UUID         null: false  primary: true
[ 1] code         TEXT         null: false  unique: true
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
    Code       string    `bun:"code,unique" json:"code"`       // [1] TEXT
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
- Format: `NIDN` + 6 digits, zero-padded (`NIDN000001`). Generated in the create use case from a Postgres sequence (`tour_code_seq`, `MAXVALUE 999999`) inside the create transaction (`FormatTourCode`). A sequence is race-free; gaps after a rollback are acceptable. The prefix lives in `config_api.yml` (`tourCode.prefix`).
- The legacy data import sets the sequence past the highest imported code with `setval`.
- `code` has a unique index. Internal references (foreign keys, audit) always use `id`.

---

## Tour taxonomy

Country and city — one module each (`pkg/modules/country`, `pkg/modules/city`). A tour has many countries and many cities (junctions `tour_program_country`, `tour_program_city`). A city belongs to one country (`country_id`).

Editable display fields: `name_th`, `name_en`, `description` (SEO copy for the landing page), `hero_image`, `sort_order`.

- `ref_id` (nullable, unique per table) is the legacy site's destination id; it is editable and maps legacy tours on import.
- Seed data (`migrations/*_seed_country_city`) comes from the legacy destination list: continents and multi-country regions (ยุโรป, สแกนดิเนเวีย, ตะวันออกกลาง) are stored as **countries**; Thai/Vietnamese regions, provinces and sub-country regions (ภาคเหนือ, ฮอกไกโด, ทิเบต) as **cities**. Seeded rows use the nil UUID as `create_by`.
- Staff endpoints: `POST /countries|/cities` (create), `PUT` (update, no slug), `POST …/list` (`pageSize` up to 500 so dropdowns load every country), `GET …/:id`, `PUT …/status`. Cities embed their country summary. A city's country must be active on create, on country change and on restore (`countryId` VALIDATION error otherwise).
- Duplicate slug / ref ID → `SLUG_TAKEN` / `REF_ID_TAKEN` (409), mapped from the unique index name.
- Permissions: `country:list|get|create|update|status`, same for `city`. `sales` has list/get only.

### Slug rules — strictly enforced

- `slug` appears in customer URLs (`/tours/japan/`, `/tours/japan/tokyo/`). Changing it breaks inbound links and indexed pages.
- Lowercase ASCII kebab-case, max 60 chars, `^[a-z0-9]+(-[a-z0-9]+)*$`, unique per table.
- **Locked after create.** Update DTOs have no `slug` field; the update use case never writes the column.
- Soft-deleting a country/city fails with `TAXONOMY_IN_USE` (including the referencing tour count) if any active tour references it.

> **Phase 2:** a `tag` dimension (`pkg/modules/tag`, junction `tour_program_tag`) following the city pattern. Do not pre-build hooks for it.

---

## Wholesales

Tour operators we source tours from (`pkg/modules/wholesale`, tables `wholesale` + `wholesale_contact`).

- Fields: `ref_id` (legacy id, nullable, unique), `code` (stored uppercase, unique, editable — not used in URLs), `name`, `state` (`pkg/constants/wholesalestate`: 10 on request, 20 connected, 30 blocked, 40 declined; default 10), plus the standard columns.
- Contacts (name + phone, max 20, in display order) are sent with every create/update and **replaced as a whole** (`WholesaleContactRepository.ReplaceForWholesale`), so `wholesale_contact` has no `status`.
- Has-many relations must load through `.Model(dest)`, not `Scan(ctx, dest)` (Bun rejects the latter).
- Errors: `CODE_TAKEN`, `REF_ID_TAKEN` (409). Permissions `wholesale:create|update|list|get|status`; `sales` has list/get.

## Airlines

Airline master (`pkg/modules/airline`, table `airline`). Seeded (`migrations/*_seed_airline`) with 99 airlines from the legacy airline list (`ref_id` = legacy id, NULL for 7 rows; the legacy "(code)" name suffix dropped; no logos). Seeded rows use the nil UUID as `create_by`.

- Fields: `ref_id` (legacy id, nullable, unique), `code` (IATA designator, or the 3-letter ICAO code when there is none; `^[A-Z0-9]{2,3}$`, stored uppercase, unique, editable), `name`, `logo` (nullable URL under `R2_PUBLIC_URL` + `airlines/logos/`, uploaded with purpose `airline-logo`), plus the standard columns.
- Endpoints follow the route conventions (`/airlines`, `/list` with keyword/status, max `pageSize` 500 for dropdowns). Errors: `CODE_TAKEN`, `REF_ID_TAKEN` (409). Permissions `airline:create|update|list|get|status`; `sales` has list/get.
- Tours link to an airline through `tour_program.airline_id` (optional); tour responses embed `airline: {id, code, name, logo, status} | null`.

## Booking sources

Channels a booking comes through — LINE OA, Facebook, agents, … (`pkg/modules/bookingsource`, table `booking_source`). Picked in the backoffice booking summary tool; no other table references them yet.

- Fields: `name` (unique, case-insensitive — index `booking_source_name_key` on `lower(name)`), `short_name` (optional label for the source's own reference name, shown as "ชื่อ <short_name>"), `sort_order` (0–9999, list order), plus the standard columns. Not seeded.
- Endpoints follow the route conventions (`/booking-sources`, `/list` with keyword/status, max `pageSize` 500, ordered by `sort_order`, then name). Error: `NAME_TAKEN` (409). Permissions `booking-source:create|update|list|get|status`; `sales` has list/get.

## Tour programs

`pkg/modules/tourprogram`, tables `tour_program` + `tour_program_country` / `tour_program_city` (ordered links, first country = main country) + `tour_program_image` (gallery).

- Fields: `ref_id` (legacy, unique), `code` (generated, see "Tour code"), `name`, `wholesale_id` (required), `wholesale_tour_code` (staff only), `state`, `days`, `nights` (`0..days`), `airline_id` (optional, FK `airline`), `highlight` (card/meta summary, ≤ 500), `description` (plain text), `banner_image` (hero of the tour detail page and share image), `card_image` (tour cards on the home and listing pages), `pdf_file` (public download), `docx_file` (**staff only — never expose on public endpoints**), `internal_note` (staff only), plus the standard columns.
- Countries (1–20), cities (0–50, each must belong to a selected country) and gallery images (≤ 30) are sent with every create/update and **replaced as a whole**. Relations load as has-many links with nested belongs-to (`CountryLinks.Country`), not `m2m`, so their order is kept.
- Newly selected wholesale / airline / countries / cities must be active; ones the tour already had may stay after they were soft-deleted (same rule on create and update).
- Media fields accept only URLs under `R2_PUBLIC_URL` with the purpose's key prefix (`tours/banners/`, `tours/cards/`, `tours/gallery/`, `tours/documents/…{.pdf|.docx}`) — i.e. files uploaded through `/media/upload`.
- Endpoints follow the route conventions (`/tour-programs`, `/list` with keyword/country/city/wholesale/state/status, max `pageSize` 100). Errors: `REF_ID_TAKEN` (409).
- Public endpoints: see "Public endpoints (customer site)".

## Periods, prices and seats

`tour_program_period` holds each departure's dates, price and seat availability. This is the data that changes most often (many times a day); tour content rarely changes after it is posted.

- **Starting price** = the lowest price among a tour's active **upcoming** periods (`status = true`, `start_date` after today in Thailand — `utils.TodayInThailand`). Sold-out periods still count. It is baked into the customer site's static HTML (tour cards, JSON-LD), so a change requires a site rebuild.
- **Per-period price and seats** are served live to the customer site's period island via `GET /api/v1/tour-programs/public/:code/periods`: small payload (dates, price, seats left, sold-out flag), `Cache-Control: public, max-age=30`. Seat changes never trigger a rebuild.
- Booking is lead-based (no online seat reservation, see "Bookings"), so up to 30 seconds of staleness is acceptable.
- Staff edit periods through the `tour-program-periods` resource (`pkg/modules/tourprogramperiod`): `POST` create, `PUT` update, `POST /list` (`{tourProgramId, status?}`, every period of one tour by date, no paging), `PUT /status`. Codes `tour-program-period:create|update|list|status`. The update use case compares the old and new row (and the tour's starting price) to decide whether to enqueue a rebuild — see the rebuild table.
- Columns: `start_date`/`end_date` (`DATE`, wire format `YYYY-MM-DD`, at most 90 days apart), `price` (adult, required), optional `price_before_discount` (must be above `price`), `child_with_bed_price`, `child_no_bed_price`, `single_supplement`, `deposit`, `seats_total` (0–999), `seats_available` (`0..seats_total`, 0 = sold out), `note` (staff only). **Prices are whole baht (`INTEGER`)** — no floats.

---

## Bookings

Booking requests from the customer site (`pkg/modules/booking`, table `booking`). A booking is a **sales lead**: no seats are reserved, no payment is taken and `seats_available` is never touched. Sales contacts the customer and adjusts seats by hand.

- `POST /bookings/public` (no auth, `rateLimit.booking` per IP): `{tourCode, periodId, adultTwin, adultSingle, childWithBed, childNoBed, firstName, lastName, email, phone, note, website}`. `website` is a honeypot — when filled, the API answers success with a random `bookingNo` and stores nothing.
- Bookable = tour `status` true and state **selling**, period active, departing after today (Thailand) and `seats_available > 0`; otherwise `PERIOD_UNAVAILABLE` (409). An unknown or hidden tour code is `NOT_FOUND`.
- Prices are **quoted by the backend** from the period: adult twin = `price`, adult single = `price + single_supplement`, children = `child_with_bed_price` / `child_no_bed_price`. A count for a row the period doesn't price is a `VALIDATION` error on that field. At least one adult, at most 50 travellers.
- The row stores a snapshot (tour code/name, dates, unit prices, counts, total) so it still reads correctly after the tour or period is edited. `booking_no` = 2-digit year in Thailand + 6 random characters from `bookingstate.NoAlphabet` (no 0/O/1/I/L, e.g. `26X8K9AQ`); a collision hits the unique index `booking_booking_no_key` and the create transaction is retried (up to 3 times). `create_by` is the nil UUID. Logs carry the booking number and tour code only — never personal data.
- `state` (`pkg/constants/bookingstate`): 10 new, 20 contacted, 30 confirmed, 40 cancelled. The first user to move a booking out of "new" becomes its `assignee_id`.
- Staff endpoints: `POST /bookings/list` (`{keyword, state, status, page, pageSize}`, newest first), `GET /bookings/:id`, `PUT /bookings` (`{id, state, staffNote}`), `PUT /bookings/status` (spam/duplicates). Permissions `booking:list|get|update|status`; `sales` has list/get/update, `finance` list/get.
- The backoffice polls `POST /bookings/list` with `{state: 10, status: true, pageSize: 1}` every 60 s for its new-booking badge.

---

## Customer site rebuild (Revalidator)

The customer site is fully static on Cloudflare Workers. When published content changes, the backend triggers a rebuild through a Cloudflare **Deploy Hook**. Frontends never call the hook.

```go
// pkg/modules/shared/services/revalidation/revalidation.go
type Revalidator interface {
    Enqueue(ctx context.Context, paths []string) // non-blocking; logs failures, never returns an error
}
```

`revalidation.New` (in-process worker, started and stopped by `app.go` after the HTTP server):
- **Debounce:** fires once no new enqueue has arrived for `quiet` (default 3 min), and at most `maxWait` (default 15 min) after the first pending enqueue. Both values live in `config_api.yml`.
- **Daily rebuild** at `revalidation.dailyAt` (Thailand time, default `00:05`; `""` disables): "upcoming" periods move with the calendar, so starting prices can change without any edit.
- **Shutdown flush:** a batch still pending when the API stops (e.g. a redeploy) is fired once before exit.
- Fires `POST <CLOUDFLARE_DEPLOY_HOOK_URL>` through `shared/client/deployhook`. Retries 3× with backoff (1 s, 5 s, 30 s); a final failure is logged at ERROR with the pending paths. The URL never appears in logs or errors.
- Without `CLOUDFLARE_DEPLOY_HOOK_URL` (local) every rebuild is logged and skipped; `POST /site/rebuild` returns `SITE_REBUILD_UNAVAILABLE` (503). A rejected trigger returns `SITE_REBUILD_FAILED` (502).
- Cloudflare skips duplicate triggers while a build is still queued, and rate-limits hooks to 10 builds/minute per Worker.
- `paths` are only logged today. They are kept so a per-path invalidation strategy can return later without touching use cases.
- Use cases enqueue **after** the transaction commits, using `revalidation.PathsForTour` / `PathsForCountry` / `PathsForCity` — never hand-rolled path lists.

When to enqueue:

| Mutation | Enqueue? |
|---|---|
| Tour create / update / status change | only when the tour is public before or after (`revalidation.TourVisible`); an update also rebuilds the landings it left |
| Period create / update / delete / restore | only when the tour is public **and** its baked summary changes (`domainmodels.PublicPeriodSummary`: starting price, its struck-through price, next departure, period count), compared before/after inside the transaction |
| Period seat change, price change above the starting price | **never** (the result is served live by the period table) |
| Country/city update or status change | only when it has public tours (`CountPublicByCountry` / `CountPublicByCity`); a city moved to another country rebuilds both landings |
| Media upload not yet referenced by a record | no |

**Build history:** every trigger the backend makes (and every skip) is a row in `site_build` (`trigger` manual/auto/daily, `outcome` triggered/failed/skipped, paths, attempts, error, `create_by` = the user for manual builds, the nil UUID otherwise), written through `revalidation.BuildLog` and never failing the trigger. Builds from git pushes are not in it, and `outcome` is whether Cloudflare accepted the hook, not the build result. Read with `POST /api/v1/site-builds/list` (`{trigger?, page, pageSize}`, newest first, `site-build:list` — admins only), shown on the backoffice page `/site-builds`.

Manual trigger: `POST /api/v1/site/rebuild` (`pkg/modules/site`, `apicode.SiteRebuild`, code `site:rebuild` — admins via `*`) fires immediately, bypasses the debounce and clears the pending batch (re-queued if the trigger fails).

---

## Object storage (Cloudflare R2)

The backend is the only holder of R2 credentials. Upload flow: backoffice → `POST /api/v1/media/upload` (multipart) → validate → re-encode → R2 → `{ url, key, size, mime }`.

- **Client:** `pkg/modules/shared/client/objectstorage` exposes an `ObjectStorage` interface (`Put`, `Delete`, `PublicURL`, `PublicBaseURL`); the R2 implementation uses aws-sdk-go-v2 (`service/s3` + `credentials`, region `auto`, endpoint `https://<account>.r2.cloudflarestorage.com`). Use cases depend on the interface only. Until every `R2_*` variable is set, the API still starts (logs a warning) and uploads return `STORAGE_UNAVAILABLE` (503).
- **Form fields:** `file`, `purpose` (`tour-banner | tour-card | tour-gallery | tour-pdf | tour-docx | country-hero | city-hero | airline-logo`).
- **Validation** (`pkg/modules/media/commons`, by content — never the client's name or Content-Type): images `image/jpeg`, `image/png`, `image/webp` only (no SVG/GIF); PDF; DOCX (a zip with `word/document.xml` and the wordprocessingml main type — `.docm` is rejected). Max `media.maxUploadMB` (10 MB). Minimum dimensions: banner ≥ 1200×630, card ≥ 800×600, gallery ≥ 800×600, hero ≥ 1600×900, airline logo ≥ 120×40, read from the header without decoding pixels; over `media.maxMegapixels` (50 MP) is rejected. Errors: `MEDIA_TYPE_NOT_ALLOWED`, `MEDIA_TOO_LARGE` (413), `IMAGE_TOO_SMALL` (message carries the sizes), `STORAGE_UNAVAILABLE`.
- **Processing — current:** files are stored **as uploaded**. **Planned:** decode → re-encode images as WebP (quality 80–85, original dimensions, strips EXIF), with a decode semaphore (e.g. 2 — a 24 MP image takes ~100 MB). Blocked on the WebP encoder decision.
- **Fiber `BodyLimit`** must be at least 12 MB (Fiber's default is 4 MB).
- **Keys are immutable and content-addressed** (`<hash>` = first 24 hex of SHA-256): `tours/banners/<hash>.<ext>`, `tours/cards/<hash>.<ext>`, `tours/gallery/<hash>.<ext>`, `tours/documents/<hash>/<ascii-name>.{pdf,docx}`, `countries/heroes/<hash>.<ext>`, `cities/heroes/<hash>.<ext>`, `airlines/logos/<hash>.<ext>`. Objects get `Cache-Control: public, max-age=31536000, immutable`; documents also get `Content-Disposition` with the original (UTF-8) file name — `inline` for PDF, `attachment` for DOCX. Re-uploading the same bytes reuses the key.
- **Deletion is soft — not built yet:** replaced or removed keys should go into `media_orphan` with `delete_after = now() + 7 days`, reaped by an in-process sweeper (`time.Ticker`). Until then, unreferenced objects simply stay in R2.
- **Permission:** `media:upload`.

---

## Public endpoints (customer site)

All unauthenticated. All filter `status = true` in the repository. All return 404 for a missing **or** soft-deleted record — the same response, so deleted records never leak.

| Endpoint | Caller | Built |
|---|---|---|
| `GET /tour-programs/public?keyword=&country=&city=&dateFrom=&dateTo=&page=&pageSize=` → `{items, total, page, pageSize}` (ordered by code, max `pageSize` 100) | site build, LIFF tour search (`keyword` matches name or code) | yes |
| `GET /tour-programs/public/:code` → tour + `images` + `periods` | site build | yes |
| `GET /tour-programs/public/:code/periods` → `{items}` | browser (period table), `Cache-Control: public, max-age=30` on 200, `no-store` otherwise | yes |
| `GET /countries/public`, `GET /cities/public` → `{items}`: every active row with `tourCount` and `startingPrice` (cities only when their country is active) | site build | yes |
| `GET /countries/public/:slug`, `GET /countries/public/:slug/cities`, `GET /cities/public/:slug` | site build | no — the site uses the lists |
| `POST /bookings/public` → `{bookingNo, totalPrice}` | browser (booking form), rate-limited | yes |

Visibility rules (repository: `wherePublic`, `ListPublic…`; DTOs: `pkg/modules/tourprogram/models/tour_program_public_item_model.go`):
- A tour is public when `status = true` and `state` is in `tourprogramstate.Public` (selling, closed). Periods: `status = true` and departing after today (Thailand). Every period of a closed tour is `soldOut`.
- Soft-deleted countries/cities linked to a tour are left out of its public `countries` / `cities`.
- `:code` is matched exactly (case-sensitive); anything outside `^[A-Za-z0-9]{1,32}$` is 404 without a query.
- Public DTOs never carry wholesale, `docxFile`, `internalNote`, `state`, `status` or audit fields.
- `tourCount` / `startingPrice` per country/city come from one aggregate query each (`PublicCountryStats` / `PublicCityStats`).
- Because "upcoming" moves with the calendar, starting prices can change without any mutation — the rebuild trigger needs a daily rebuild as well (see "Customer site rebuild").

- Country/city detail returns 404 when the slug has zero active tours (empty landing pages hurt SEO).
- Build-time callers should need few requests. If site builds slow down, add a bulk endpoint rather than N per-tour calls.
- Browser-called endpoints need the customer origin in CORS. `POST /bookings/public` is rate-limited per client IP with Fiber's built-in `limiter` (`rateLimit.booking`).

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

`POST /users/list`, `GET /users/:id`, `POST /users` (`temporaryPassword` required, sets `mustChangePassword`), `PUT /users` (`id, email, name, phone, roles`), `PUT /users/status`, `PUT /users/password` (admin reset: new temporary password, sets `mustChangePassword`). Permissions `user:create|update|list|get|status|password-reset`.

- `phone` is optional (`app_user.phone`, `''` = none). Spaces, dashes, dots and parentheses are stripped; the rest must be 9–15 digits with an optional leading `+`. It is shown to customers in sales tool texts.
- `GET /users/sales` (`sales-user:list`, held by `sales`): active sales users `{id, name, phone}` by name plus `defaultId` (the caller when they are one). For the backoffice sales tools, so sales staff don't need `user:list`. Registered before `/users/:id`.

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

Environments: `local`, `dev`, `prd` (`APP_ENV`). Each deployed environment has its own Supabase project, R2 bucket, deploy hook and secrets. Never share a value across environments. Never copy prd data into dev — bookings contain customer personal data (PDPA).

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
| `CLOUDFLARE_DEPLOY_HOOK_URL` | The URL itself is the credential — treat as a secret. Optional: without it rebuilds are skipped. `https://` required outside local. One hook per Worker: `customer-dev` (dev), `customer-prd` (prd) |
| `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY` | Token scoped to this environment's bucket |
| `R2_BUCKET` | `nidnoi-travel-media-dev` / `nidnoi-travel-media-prod` |
| `R2_PUBLIC_URL` | `https://dev-media.nidnoitravelv2.com` / `https://media.nidnoitravelv2.com` |

CORS origins per environment:

| Env | Origins |
|---|---|
| local | `http://localhost:4321`, `http://localhost:5173`, `http://localhost:5174` |
| dev | `https://dev.nidnoitravelv2.com`, `https://dev-admin.nidnoitravelv2.com`, `https://dev-liff.nidnoitravelv2.com` |
| prd | `https://www.nidnoitravelv2.com`, `https://admin.nidnoitravelv2.com`, `https://liff.nidnoitravelv2.com` |

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
task seed-mock-tours               # local/dev only: ~22 multi-country tours from the legacy site (cmd/migrate/seeddata/mock_tours.json); `-- remove` deletes them and bookings on them
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

## LINE OA bookings

Bookings submitted by sales staff through a LIFF app inside LINE OA chat (`pkg/modules/linebooking`, table `line_booking`). Unlike customer bookings (which are self-service leads), LINE bookings come with a payment slip and are pre-confirmed by a sales person.

- **LINE client** (`pkg/modules/shared/client/line/line_client.go`): verifies LIFF access tokens via `GET https://api.line.me/oauth2/v2.1/verify?access_token=...` and fetches the user profile. Registered in `ShareRegistry.ClientRegistry.LINE`. Logs a warning and disables itself when `LINE_LIFF_CHANNEL_ID` is empty.
- **Config:** `LINEConfig` in `pkg/core/config/config.go` loads `LINE_CHANNEL_SECRET`, `LINE_CHANNEL_ACCESS_TOKEN`, `LINE_LIFF_CHANNEL_ID`, `LINE_LIFF_ID`. Only `LINE_LIFF_CHANNEL_ID` is required for token verification.
- **Staff LINE link** (`app_user.line_user_id`, unique): staff link their own LINE account so the LIFF app knows who they are. `POST /auth/line-link` (JWT) stores the SHA-256 of a random one-time code (10 min) and returns `https://liff.line.me/<LINE_LIFF_ID>?lineLink=<code>`, shown as a QR code; a new call replaces the code. `POST /liff/line-link {code}` (LIFF token) finishes it → `LINE_LINK_CODE_INVALID` (400), `LINE_ALREADY_LINKED` (409, one staff user per LINE account). `POST /auth/line-unlink` removes it. `/auth/me` returns `line {displayName, linkedAt}`; user items return `lineDisplayName`.
- **LIFF auth** (`middleware.LIFFAuth`, resolver `authUseCase.ResolveLIFFSession`): `Authorization: Bearer <LIFF access token>`, verified with LINE on every request, then mapped to the linked user → `UserRequestInfo` (plus `LineUserID`, `LineDisplayName`), so LIFF routes use `RequiresPermissions` like JWT routes. Not linked or deactivated → `LINE_NOT_LINKED` (403); temporary password → `PASSWORD_CHANGE_REQUIRED`. LIFF routes require `line-booking:create` (sales; admin and super admin by wildcard; not finance).
- **LIFF endpoints:**
  - `GET /liff/me` — the linked staff user; the LIFF app's gate.
  - `POST /line-bookings` — create a LINE booking. Loads tour+period, quotes prices, generates `booking_no`, inserts in tx; `create_by` is the linked staff user. `salesUserId` (required) is the sales owner (`line_booking.sales_user_id`): an active user with the `sales` role, else `VALIDATION` on `salesUserId`. Staff responses carry `salesUser {id, name}` (null on older rows).
  - **Owner scope:** staff whose roles don't include `super_admin`, `admin` or `finance` (i.e. sales) see only bookings with `sales_user_id` = themselves — list is filtered, get/update/line-contact/status return `NOT_FOUND` for others (`usecase/line_booking_owner_scope.go`). Ownerless rows (before `sales_user_id`) are visible to see-all roles only.
  - `GET /liff/sales-users` — owner options: active `sales` users by name, `defaultId` = the caller when they are sales.
  - `POST /media/upload/liff` — upload slip image (`purpose` must be `booking-slip`), delegates to the media upload use case.
- **Staff endpoints** (JWT + Casbin):
  - `POST /line-bookings/list` — `{keyword, state, status, page, pageSize}`, newest first.
  - `GET /line-bookings/:id` — get one.
  - `PUT /line-bookings` — `{id, state, staffNote}`.
  - `PUT /line-bookings/line-contact` — `{id, lineContactId}` links the customer's LINE OA chat (`line_booking.line_contact_id`, FK `line_contact`); `null` unlinks. Permission `line-booking:update`. The receipt is pushed to this chat later. Every staff response carries `lineContact {id, displayName, pictureUrl, following}` or `null`.
  - `PUT /line-bookings/status` — soft delete/restore.
- **States** (`pkg/constants/linebookingstate`): 10 new, 20 confirmed, 30 cancelled.
- **Permissions** (`line-booking:create|list|get|update|status`): admin has all; sales has create/list/get/update; finance has list/get.
- **Booking number format:** same as customer bookings — 2-digit Thai year + 6 random characters, unique index with retry.
- **Error codes:** `LINE_NOT_CONFIGURED`, `INVALID_LINE_TOKEN`, `LINE_NOT_LINKED`, `LINE_LINK_CODE_INVALID`, `LINE_ALREADY_LINKED`.
- **Media purpose:** `booking-slip` — images only, min 200x200, key prefix `bookings/slips/`.
- **CORS:** LIFF origins (`dev-liff.nidnoitravelv2.com`, `liff.nidnoitravelv2.com`) must be added to `CORS_ALLOWED_ORIGINS`.

### LINE contacts (Messaging API webhook)

Customers who chat with the OA, collected so sales can link a booking to the customer's chat and the backend can later push messages (receipts) to them (`pkg/modules/linecontact`, table `line_contact`).

- **Webhook:** `POST /line/webhook` (no JWT). The use case checks `X-Line-Signature` (base64 HMAC-SHA256 of the raw body, keyed with `LINE_CHANNEL_SECRET`) → `INVALID_LINE_SIGNATURE` (401); without the secret → `LINE_NOT_CONFIGURED` (503). An empty event list (the console's "Verify" button) answers 200.
- Only events from a **user** source are kept: `follow` (following = true, refresh profile), `unfollow` (following = false), `message` (`last_message_at`, profile refreshed when older than 24 h). Group/room events and other types are ignored. **Message text is never stored or logged** (PDPA) — only the profile.
- Rows are created race-free with `INSERT … ON CONFLICT (line_user_id) DO NOTHING` (`CreateIfAbsent`), then updated; redelivered events are harmless.
- **Profile:** `GET /v2/bot/profile/{userId}` with `LINE_CHANNEL_ACCESS_TOKEN`. 404 = not a friend (never added or blocked) → `following = false`. Without the token, contacts are kept without a profile. Other lookup failures are logged and never fail the webhook.
- `following` is what decides whether a push can reach the customer.
- **Staff:** `POST /line-contacts/list` (`{keyword, following?, page, pageSize}`, most recent chat first; `line-contact:list` — admin, sales, finance) for the contact picker.
- **Contact picker in the LIFF:** `POST /liff/line-contacts/list` (LIFF auth + `line-contact:list`, same body and result as the staff list). `POST /line-bookings` takes an optional `lineContactId` (unknown or deleted → `VALIDATION` on `lineContactId`).
- **Same provider:** the LINE Login (LIFF) channel and the Messaging API channel must be under the same LINE provider, or the user IDs from LIFF and from the webhook differ.
- **LINE console:** Messaging API → Webhook URL `https://dev-api.nidnoitravelv2.com/api/v1/line/webhook` (prd: `api.*`), "Use webhook" on. In OA Manager keep **Chat** on so sales still chat by hand.

---

## Open questions

- **WebP encoder** — `disintegration/imaging` decodes but cannot encode WebP. Choose an encoder (cgo-based `chai2010/webp` vs a pure-Go/wasm option) and confirm it builds on Railway and on Windows (cgo needs gcc there).
- **Sales/finance permissions** — rows in `policy.csv` are placeholders; confirm with the business before prd.
- **golangci-lint** — the locally installed binary was built with Go 1.25 and refuses this Go 1.26 module; reinstall it with Go 1.26 before relying on `task lint`.
- **Booking spam protection** — rate limiting + honeypot only for now; consider Cloudflare Turnstile before launch.
- **Legacy data import** — preserve legacy `code` values and `setval` the tour code sequence past the highest one.
- **Legacy tour codes** — open until the legacy data is inspected:
  - *Casing:* new codes are `NIDN` + 6 digits (`NIDN000001`), but the docs show legacy codes lowercase (`/tour/nidn261690/`). URL paths are case-sensitive, so decide whether imported codes keep their legacy casing, are normalized, or the customer site lowercases codes in URLs (which would need 301s from the other casing).
  - *Number range:* `tour_code_seq` starts at 1. If staff create tours before the import, a new code could collide with a legacy one (the unique index would then fail the import). Decide whether to `setval` past the legacy range now.
- **Not built yet:** the media orphan sweeper and the public country/city detail endpoints.
- **Booking notifications** — the backoffice polls `POST /bookings/list`; an email or LINE push to sales would need a new provider and secret (ask first).
- **`TAXONOMY_IN_USE`** is documented but not enforced yet: a country/city with public tours can still be soft-deleted (the rebuild trigger handles it, but the landing page and tour links disappear).
