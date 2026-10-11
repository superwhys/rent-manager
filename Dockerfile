ARG NODE_VERSION=24.19.0
ARG GO_VERSION=1.27.1
ARG ALPINE_VERSION=3.22

FROM node:${NODE_VERSION}-alpine AS web-builder

WORKDIR /src/web

RUN npm install --global pnpm@9.15.4

COPY web/package.json web/pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

FROM golang:${GO_VERSION}-alpine AS go-builder

ARG BINARY_NAME=rent-manager
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
COPY --from=web-builder /src/web/dist ./web/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags "-s -w -X github.com/miebyte/goutils/buildinfo.Version=${VERSION} -X github.com/miebyte/goutils/buildinfo.ServiceName=${BINARY_NAME}" \
      -o /out/${BINARY_NAME} .

FROM alpine:${ALPINE_VERSION}

ARG BINARY_NAME=rent-manager

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app \
    && chown -R app:app /app

WORKDIR /app

COPY --from=go-builder /out/${BINARY_NAME} /app/rent-manager

USER app

EXPOSE 8080

ENTRYPOINT ["/app/rent-manager"]
CMD ["--configFile", "/app/config.json"]
