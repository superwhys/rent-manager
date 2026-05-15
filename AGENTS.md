# AGENTS.md

This file provides guidance to Codex and other coding agents when working in this repository.

## Commands

- `go test ./...` - run all tests and compile all packages.
- `go test ./internal/app/services` - run tests for one package.
- `go test ./internal/app/services -run TestName` - run a single Go test by name.
- `go build ./...` - compile all packages.
- `go run .` - run the server; it expects WeChat, JWT, Redis, and MySQL config flags parsed by `github.com/miebyte/goutils/flags` in `main.go`.
- `gofmt -w <files>` - format changed Go files before finishing code changes.
- `make generate` or `go generate ./internal/ent` - regenerate Ent code after changing schemas under `internal/ent/schema`.
- `go generate ./api` - regenerate Swagger docs from annotations in `api/`.

## Code Style

- Follow the existing Go style and local package patterns.
- Keep changes scoped to the requested behavior.
- Run `gofmt` on changed Go files.
- Add comments only where they clarify exported identifiers, business rules, non-obvious control flow, or integration boundaries.
- Do not edit generated Ent files directly unless the task explicitly requires it; update schemas and regenerate instead.

## Architecture

This is a Go 1.26.2 Gin API server for a rent-management backend. `main.go` wires configuration, Redis rate limiting, MySQL/Ent, repositories, services, system RBAC initialization, and the HTTP server. The Gin engine is mounted under `/api` via `cores.WithHttpHandler("/api", router)`.

The repository follows a DDD-style architecture:

- Interface layer: `api/` contains HTTP router registration, handlers, middleware, and generated Swagger docs.
- Application layer: `internal/app/` contains use-case services that orchestrate domain logic and repositories.
- Domain layer: `internal/domain/` contains domain models, params, repository interfaces, and domain errors.
- Infrastructure layer: `internal/infra/` contains concrete infrastructure implementations, currently Ent-backed repositories in `internal/infra/repo`.

Routing is centralized in `api/api.go`. Route groups live in `api/routers` and are registered with `github.com/miebyte/goutils/ginutils` options rather than direct `gin.Engine.Group` calls.

Current route groups:

- Global middleware: `RateLimitMiddleware`.
- `/auth`: public auth routes.
- `/admin`: token-protected admin RBAC routes.
- Protected application routes: units, rooms, meter items, and role assignment routes, all using `TokenVerifyMiddleware`.

There is currently no permission-check middleware wired into `api/api.go`; do not document or assume runtime API authorization beyond token validation unless the code is updated.

Handlers should follow the existing flow:

- Define request DTOs in `internal/models/requests`.
- Use `ginutils.RequestHandler` for binding and validation.
- Call an app service from `internal/app/services`.
- Return response DTOs from `internal/models/response`.
- Use `response.HandleRouterError` for business/application errors.

## Persistence

Persistence uses Ent with MySQL. `main.go` currently calls auto-migration on startup. Soft deletes are represented by `deleted_at`; repository queries usually filter `DeletedAtIsNil()` and deletes set `deleted_at` rather than removing rows.

Important locations:

- `internal/ent/schema` - Ent schema definitions.
- `internal/ent` - generated Ent code.
- `internal/infra/repo` - concrete repository implementations.

## Authentication And RBAC

Authentication uses JWT utilities in `internal/jwtutils`. `TokenVerifyMiddleware` stores `AccountTokenClaims` in both Gin context and request context.

The authenticated account ID comes from `AccountTokenClaims.UserID`; application services read it through `services.AccountIDFromContext`.

RBAC and business resources are currently account-scoped, not organization-scoped. There is no `organization_id` field or organization model in the current Ent schemas.

Current RBAC model:

- `Account` owns roles through `roles.account_id`.
- `Role` belongs to one account and can bind many global permission groups.
- `PermissionGroup` is global and can bind many permissions.
- `Permission` is global and has unique `code` and `api_path`.
- `UnitAccountRole` records that one account granted roles for a unit to another account.

Effective authorization follows:

`account -> role -> permission_group -> permission`

System RBAC is initialized on startup by `RBACService.InitSystemRBAC`, which ensures the super-admin permission group and allow-all permission exist and are linked.

Admin RBAC endpoints are under `/admin`. Global permission and permission-group management stays there. Role creation and role-to-permission-group binding are account-scoped: repositories filter roles by the authenticated owner account ID. Unit-account role assignment verifies that the granting account owns the target unit and that assigned roles belong to that granting account.

Stored `permissions.api_path` values generally should not include `/api`, because the router is mounted under `/api`.

## Resource Ownership

Business resources are scoped by the authenticated account:

- `Unit`, `Role`, and `MeterItem` use `account_id` through `AccountMixin`.
- `Room` is scoped through its owning unit; room repository queries use a unit account predicate.
- Unit default meter items and room meter items must belong to the same authenticated account.
- Soft-deleted records should not be returned or mutated unless the specific restore/init flow requires `ClearDeletedAt`.

## API Docs

Swagger generation is configured by `api/generate.go`. Keep annotations near handlers in `api/routers` and regenerate docs with `go generate ./api` when API contracts change.
