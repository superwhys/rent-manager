# AGENTS.md

This file provides guidance to Codex and other coding agents when working in this repository.

## Commands

- `make build-web` - install locked frontend dependencies with pnpm and build `web/dist`; required before the first `go run`, `go test`, or `go build`, and after frontend changes.
- `make build` - build the frontend and compile `rent-manager-server` with its static assets embedded.
- `go test ./...` - run all tests and compile all packages.
- `go test ./internal/app/rbac` - run tests for one package.
- `go test ./internal/app/rbac -run TestName` - run a single Go test by name.
- `go build ./...` - compile all packages.
- `go run .` - run the server on `0.0.0.0:8080`; it requires authkit session, MySQL, and Redis configuration; SMTP configuration enables email verification parsed by `github.com/miebyte/goutils/flags`.
- `gofmt -w <files>` - format changed Go files before finishing code changes.
- `make generate` or `go generate ./cmd/generate` - regenerate Gen queries after changing models under `internal/infra/persistence/models`.
- `go generate ./api` - regenerate Swagger docs from annotations in `api/`.

## Code Style

- Follow the existing Go style and local package patterns.
- Keep changes scoped to the requested behavior.
- Run `gofmt` on changed Go files.
- Add comments only where they clarify exported identifiers, business rules, non-obvious control flow, or integration boundaries.
- Do not edit generated Gen queries directly; update persistence models and regenerate instead.
- User-facing error messages should be written in Chinese. Keep error variables stable for `errors.Is`, but make their `Error()` text Chinese.
- Application services should return domain/repository errors directly. Do not add `mapXError` helpers in `internal/app/<business>` that translate domain errors into HTTP/errcode errors; HTTP response mapping belongs at the interface layer.

## Architecture

This is a Go 1.27.1 Gin API server for a rent-management backend. `main.go` wires `github.com/miebyte/authkit v1.9.0`, its email/password authentication and MySQL persistence, business repositories/services, and Redis rate limiting. It creates the HTTP router and mounts it under `/api` via `cores.WithHttpHandler("/api", router)`. The `web` Go package embeds the Vue build output from `web/dist` and serves it at `/`, with `index.html` fallback for frontend history routes. API, Swagger, and health-check routes retain their existing handlers. Build the frontend before compiling Go; production only needs the binary, service configuration, and existing backing services.

The repository follows a DDD-style architecture:

- Interface layer: `api/` contains HTTP router registration, handlers, middleware, and generated Swagger docs.
- Application layer: `internal/app/<business>/` contains business use-case services and DTOs, including account-scoped RBAC. Authentication and account lifecycle are owned by authkit. `internal/app/auth` coordinates verified email registration and password setup in one authkit transaction; there is no local account model or authentication domain.
- Domain layer: `internal/domain/` contains domain models, params, repository interfaces, and domain errors.
- Infrastructure layer: `internal/infra/` contains supporting infrastructure packages and GORM + Gen repositories under `internal/infra/persistence/mysql`.

Routing is centralized in `api/api.go`. Route groups live in `api/routers` and are registered with `github.com/miebyte/goutils/ginutils` options rather than direct `gin.Engine.Group` calls.

Current route surface:

- `POST /api/auth/email/code` sends an email verification code.
- `POST /api/auth/register` verifies the email and creates both email and password credentials atomically.
- `POST /api/auth/password-login` and `POST /api/auth/email-login` authenticate existing accounts by password or email code.
- `GET /api/auth/session` returns the current authkit account and session expiry; `DELETE /api/auth/session` revokes that session.
- `PATCH /api/auth/account` updates the current account's display name and returns the authkit account. Names are trimmed, require 1–64 Unicode characters, and reject control characters; credentials and sessions are preserved.
- `/api/roles/*`, `/api/admin/permission-groups/*`, and `/api/units/:unit_id/grants` or `/api/units/:unit_id/accounts/:account_id/roles` call the RBAC service and GORM repository. `GET /api/admin/permissions` returns the fixed backend permission catalog; permission creation APIs are removed.
- The former `/api/auth/refresh-token` and `/api/accout/self` routes are removed.

Global Redis-backed rate limiting is wired in `api/api.go` via `RateLimitMiddleware`. The router disables default proxy trust; only configure explicitly trusted proxy addresses when deploying behind a reverse proxy, so forwarded headers cannot spoof authentication rate-limit identities. Email code delivery, registration, and the two email login endpoints are public. Session and business routes require a valid authkit Bearer session token; resource-level permission checks are performed in application services, not route middleware.

Handlers should follow the existing flow:

- Define business request DTOs in `internal/app/<business>/dto.go`.
- Use `ginutils.RequestHandler` for binding and validation.
- Call an app service from `internal/app/<business>`.
- Return business response DTOs from `internal/app/<business>/dto.go`; auth endpoints expose `authkit.Account` directly without an account mapping.
- Use `response.HandleRouterError` for business/application errors.

## Persistence

Persistence uses GORM with MySQL and type-safe queries generated by `gorm.io/gen`. `main.go` opens the MySQL client, migrates `authmysql.Models()` from authkit before business `models.AllModels()`, and wires business repositories through `mysql.RepositoryFactory`. `WithTransaction` supplies repositories bound to the same transaction. Startup refuses an existing legacy `accounts` table before AutoMigrate. [The authkit migration SQL](migrations/20261009_authkit_accounts.sql) is only for empty legacy account-owned tables and refuses nonempty data; preserving historical data needs a separate explicit migration plan.

`EnsureBuiltinPermissionSchema` refuses a legacy `permissions` table or `permission_group_permissions.permission_id` column before AutoMigrate. Existing databases require [the fixed permission migration](migrations/20261010_builtin_permissions.sql) after stopping writes and backing up; new databases need no manual migration. It maps actual old IDs to codes, retains only enabled/non-deleted permission bindings, and rejects unknown codes, orphan bindings or unexpected foreign keys before modifying business tables. MySQL DDL implicitly commits; an interrupted migration requires backup recovery. Permission catalog entries are not persistence models; group bindings store exact `permission_code` strings with `utf8mb4_bin` collation.

Important locations:

- `internal/infra/persistence/models` - independent persistence models and migration list.
- `internal/infra/persistence/mysql/query` - generated Gen queries; do not edit manually.
- `cmd/generate` - reproducible model-to-query generator.
- `internal/infra/persistence/mysql` - concrete repositories, conversion and transaction handling.

Keep business table names, nullability, indexes and resource relationships compatible with existing data. Account references now use `varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin` and point directly to authkit IDs; business models do not define foreign keys or associations to the removed local `accounts` model. `reader_id` is a nullable string. Existing integer account references require an explicit historical-data migration decision; do not invent an account mapping or treat a numeric-to-string cast as identity migration. Models use `*time.Time` for explicit soft-delete filtering and `*bool` for default-true fields so explicit false values survive creation. Business CRUD must use generated queries. Preserve zero values, NULL updates and request contexts; destructive schema changes require separate versioned SQL.

Rent bills store complete meter-reading snapshots in nullable JSON `rent_bills.meter_readings`. Generation calculates each frozen line fee and the bill meter fee from the same readings and prices. Regenerating an unpaid bill replaces both amounts and snapshots; paid/cancelled bills are skipped, and status changes preserve snapshots. Detail responses expose the saved `meter_readings` under `rent_bill.view` without reading current dictionaries or meter records or requiring `meter_reading.view`; list and status responses keep the summary contract. Preserve `null` for legacy bills without snapshots and `[]` for generated bills with no meter items. Never backfill old bill details from current data. The [snapshot column migration](migrations/20261010_rent_bill_meter_reading_snapshots.sql) adds the nullable column without backfilling; new databases use AutoMigrate.

## Authentication And RBAC

Authentication, accounts, credentials and sessions are owned by `github.com/miebyte/authkit`. Configure `auth.smtp` for email code delivery and `auth.session_ttl` (for example `720h`). Zero session TTL uses authkit's 30-day default. SMTP can be omitted to run the server and use existing password credentials; code delivery and registration require SMTP. Registration verifies the email, sets its password, and returns a new session in one transaction. Both login endpoints authenticate existing accounts only. Re-registering an existing account must not change its password. The Bearer token is an opaque, revocable session token, not a JWT, and expired sessions require a new login. Preserve authkit verification rejection counters when managing its transactions. Passwords are passed verbatim and never stored or logged by the frontend.

Use authkit account IDs directly throughout business code as `string` (64 lowercase hexadecimal characters), with no local account model or ID mapping. `internal/app.AccountIDFromContext` and `SetAccountIDToContext` carry that string. Auth responses retain authkit's `Account` JSON fields `ID`, `Username`, `Email`; the outer login fields are `account`, `access_token`, `expires_at`, `created`. Business resource IDs remain `int`.

RBAC and business resources are currently account-scoped, not organization-scoped. There is no `organization_id` field or organization model in the current persistence models.

Current RBAC model:

- Accounts have full access to their own resources by default.
- Resource-level permission checks are only needed when an account accesses resources under another account's unit.
- Authkit accounts own roles through `roles.account_id`.
- `Role` belongs to one account and can bind many global permission groups.
- `PermissionGroup` is global and can bind many fixed permission codes.
- `Permission` is a read-only backend catalog entry defined in `internal/domain/rbac/permissions.go`, with a unique `code` and Chinese description; it has no database table or numeric ID.
- `UnitAccountRole` records that a unit owner granted owner-defined roles for a unit to another account.

Effective authorization follows:

`target account -> unit_account_role -> owner role -> permission_group -> fixed permission_code`

Unit lists keep only `id`, `name`, `access_type`, and `is_owner` for grants without `unit.view`; other list fields require the corresponding unit, room, or bill view permission. Home room and financial aggregates also check each unit's view permissions. Role responses expose `permission_group_ids`, and permission-group responses expose `permission_codes`, including empty arrays. Group permission binding requests use `permission_codes` and reject any code outside the fixed catalog.

Room-type and meter-item GET endpoints accept an optional `unit_id`. With it, the service checks `room_type.view` or `meter_item.view` on that unit before reading the owner's shared catalog for selection in the unit. Without it, queries remain scoped to the signed-in account. Shared definitions can only be created, changed, or deleted by their owner; per-unit grants do not confer account-wide dictionary write access. `room.manage` and `unit.update` continue to control selecting existing definitions in a unit.

The RBAC API covers a read-only fixed permission catalog, global permission-group management, account-private role management, role permission-group binding, and unit role grants. Permission-group write and permission-code binding APIs require `AdminConfig.SuperAdminAccountIDs` (`[]string` containing exact authkit account IDs); normal accounts can list fixed permissions/groups, manage their own roles, and grant roles only on units they own. Do not add permission creation APIs or restore the `permissions` persistence model.

Before granting unit roles, the RBAC application service checks the target account directly through authkit's `GetByID` account repository method. Define permission checks by code constants in `internal/domain/rbac/permissions.go` and call `domain/rbac.RbacService.RequireUnitPermission` from business application services. Do not maintain a route-to-permission map unless route-level authorization is intentionally introduced later.

## Resource Ownership

The persistence models model business resources as account-scoped:

- `Unit`, `Role`, `RoomType`, `MeterItem`, and `AccountSetting` use authkit `account_id` strings.
- `UnitAccountRole.AccountID`, `GrantedByAccountID`, and optional `MeterReading.ReaderID` use the same authkit identity. Account settings remain business data.
- `Room` is scoped through its owning unit.
- Future repository queries should enforce account ownership and filter soft-deleted records.

## API Docs

Swagger generation is configured by `api/generate.go`. Keep annotations near handlers in `api/routers` and regenerate docs with `go generate ./api` when API contracts change.
