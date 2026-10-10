.PHONY: generate build build-web

generate:
	go generate ./cmd/generate
	go generate ./api

build-web:
	cd web && pnpm install --frozen-lockfile
	cd web && pnpm build

build: build-web
	go build -o rent-manager-server .
