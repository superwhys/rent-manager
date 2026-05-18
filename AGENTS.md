# AGENTS.md

This file provides guidance to Codex and other coding agents when working in this repository.

## Commands

- `go test ./...` - run all tests and compile all packages.
- `go test ./internal/app/services` - run tests for one package.
- `go test ./internal/app/services -run TestName` - run a single Go test by name.
- `go build ./...` - compile all packages.
- `go run .` - run the server on `0.0.0.0:8080`; it currently requires MySQL and Redis flags parsed by `github.com/miebyte/goutils/flags`.
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

This is a Go 1.26.2 Gin API server for a rent-management backend. The current codebase is being rebuilt from a reduced skeleton. `main.go` wires MySQL/Ent, Redis rate limiting, `services.AppService`, `services.RBACService`, creates the HTTP router, and mounts it under `/api` via `cores.WithHttpHandler("/api", router)`.

The repository follows a DDD-style architecture:

- Interface layer: `api/` contains HTTP router registration, handlers, middleware, and generated Swagger docs.
- Application layer: `internal/app/` contains use-case services, currently `AppService` and permission/permission-group focused `RBACService`.
- Domain layer: `internal/domain/` contains domain models, params, repository interfaces, and domain errors.
- Infrastructure layer: `internal/infra/` contains supporting infrastructure packages and Ent-backed repositories under `internal/infra/repo`.

Routing is centralized in `api/api.go`. Route groups live in `api/routers` and are registered with `github.com/miebyte/goutils/ginutils` options rather than direct `gin.Engine.Group` calls.

Current route surface:

- `GET /api/health` calls `AppService.Health`.
- `/api/admin/permission-groups/*` and `/api/admin/permissions/*` call the RBAC service and Ent repository.

Global Redis-backed rate limiting is wired in `api/api.go` via `RateLimitMiddleware`. There is currently no authentication or permission-check middleware wired into `api/api.go`; `/admin` is only a route prefix at the moment.

Handlers should follow the existing flow:

- Define request DTOs in `internal/models/requests`.
- Use `ginutils.RequestHandler` for binding and validation.
- Call an app service from `internal/app/services`.
- Return response DTOs from `internal/models/response`.
- Use `response.HandleRouterError` for business/application errors once business services are reintroduced.

## Persistence

Persistence uses Ent with MySQL. `main.go` opens the MySQL client, runs Ent migration, and wires the RBAC repository. Repository implementations should live under `internal/infra/repo`.

Important locations:

- `internal/ent/schema` - Ent schema definitions.
- `internal/ent` - generated Ent code.
- `internal/infra/repo` - intended location for future concrete repository implementations.

## Authentication And RBAC

Authentication utilities still exist in `internal/jwtutils` and `api/middlewares`, but they are not currently wired into the active API router. The previous `services.AccountIDFromContext` helper was removed with the old application services.

RBAC and business resources are currently account-scoped, not organization-scoped. There is no `organization_id` field or organization model in the current Ent schemas.

Current RBAC model:

- `Account` owns roles through `roles.account_id`.
- `Role` belongs to one account and can bind many global permission groups.
- `PermissionGroup` is global and can bind many permissions.
- `Permission` is global and has unique `code` and `api_path`.
- `UnitAccountRole` records that one account granted roles for a unit to another account.

Effective authorization follows:

`account -> role -> permission_group -> permission`

The currently wired RBAC API covers global permission and permission-group management only: create/list permissions, create/list permission groups, and add/remove permissions on permission groups.

Stored `permissions.api_path` values generally should not include `/api`, because the router is mounted under `/api`.

## Resource Ownership

The remaining Ent schemas model business resources as account-scoped:

- `Unit`, `Role`, and `MeterItem` use `account_id` through `AccountMixin`.
- `Room` is scoped through its owning unit.
- Future repository queries should enforce account ownership and filter soft-deleted records.

## API Docs

Swagger generation is configured by `api/generate.go`. Keep annotations near handlers in `api/routers` and regenerate docs with `go generate ./api` when API contracts change.
