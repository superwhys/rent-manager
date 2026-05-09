# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `go test ./...` — run all tests and compile all packages.
- `go test ./internal/app/services` — run tests for one package.
- `go test ./internal/app/services -run TestName` — run a single Go test by name.
- `go build ./...` — compile all packages.
- `go run .` — run the server; it expects WeChat, JWT, Redis, and MySQL config flags parsed by `github.com/miebyte/goutils/flags` in `main.go`.
- `gofmt -w <files>` — format changed Go files before finishing code changes.
- `make generate` or `go generate ./internal/ent` — regenerate Ent code after changing schemas under `internal/ent/schema`.
- `go generate ./api` — regenerate Swagger docs from annotations in `api/`.

## Architecture

This is a Go 1.26 Gin API server for a rent-management backend. `main.go` wires configuration, Redis rate limiting, MySQL/Ent, repositories, services, and the HTTP server. The Gin engine is mounted under `/api` via `cores.WithHttpHandler("/api", router)`.

This repository follows a DDD-style architecture. Keep new business capabilities aligned with these layers:

- Interface layer: `api/` contains HTTP router registration, handlers, middleware, and generated Swagger docs.
- Application layer: `internal/app/` contains use-case services that orchestrate domain logic and repositories.
- Domain layer: `internal/domain/` contains domain models, params, repository interfaces, and domain errors.
- Infrastructure layer: `internal/infra/` contains concrete infrastructure implementations, currently Ent-backed repositories in `internal/infra/repo`.

Routing is centralized in `api/api.go`. Public auth routes are mounted separately from protected routes. Protected routes run `TokenVerifyMiddleware` followed by `PermissionMiddleware`; route groups themselves live in `api/routers` and are registered with `github.com/miebyte/goutils/ginutils` options rather than direct `gin.Engine.Group` calls.

Handlers follow a consistent flow: define request DTOs in `internal/models/requests`, use `ginutils.RequestHandler` for binding and validation, call an app service from `internal/app/services`, and return response DTOs from `internal/models/response`. Business errors should go through `response.HandleRouterError`, which returns HTTP 200 with an app-level error code; middleware gate failures use real HTTP statuses such as 401, 403, 429, or 500.

Supporting packages:

- `internal/ent/schema` — Ent schema definitions; most files under `internal/ent` are generated and should be updated via `go generate ./internal/ent`.
- `internal/models/requests` and `internal/models/response` — transport DTOs and response mapping helpers.

Persistence uses Ent with MySQL. `internal/entutils/client.go` creates the Ent client and `main.go` currently calls auto-migration on startup. Soft deletes are represented by `deleted_at`; repository queries usually filter `DeletedAtIsNil()` and deletes set `deleted_at` rather than removing rows.

Authentication uses JWT utilities in `internal/jwtutils`. `TokenVerifyMiddleware` stores `AccountTokenClaims` in both Gin context and the request context. Permission checks use the authenticated `UserID` from claims and compare permissions against Gin route templates from `c.FullPath()` such as `/rooms/:id`; because the router is mounted under `/api`, stored `permissions.api_path` values generally should not include `/api`.

RBAC currently consists of accounts, roles, permissions, `account_roles`, and `role_permissions`. Admin endpoints are under `/admin`; a bootstrap allowance exists in `PermissionMiddleware` so initial role/permission/account-role setup can happen before any account-role bindings exist.

Swagger generation is configured by `api/generate.go`. Keep annotations near handlers in `api/routers` and regenerate docs with `go generate ./api` when API contracts change.
