# syntax=docker/dockerfile:1

# ── build ─────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS build
WORKDIR /src

# Dependencies are their own layer so code edits do not re-download them.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary: the runtime image has no libc to link against.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api


# ── runtime ───────────────────────────────────────────────────────────────
FROM alpine:3.21 AS runtime
WORKDIR /app

# ca-certificates for outbound TLS, tzdata so timestamps localise correctly.
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S brag && adduser -S -G brag brag

COPY --from=build /out/api /app/api
# Shipped for reference and for running migrations by hand against the volume.
COPY --from=build /src/migrations /app/migrations

USER brag
EXPOSE 8080

ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/health >/dev/null || exit 1

ENTRYPOINT ["/app/api"]
