# ── Stage 1: base ───────────────────────────────────────────────
FROM golang:1.26-alpine AS base
WORKDIR /moneymate

# install dependencies
RUN apk add --no-cache git

ENV GONOSUMCHECK=* GONOSUMDB=*

# copy go workspace files first for layer caching
COPY go.work go.work.sum ./

# copy all go.mod files so workspace resolves
COPY shared/go.mod shared/go.mod
COPY services/auth/go.mod services/auth/go.mod
COPY services/gateway/go.mod services/gateway/go.mod

# download dependencies
RUN go work sync

# ── Stage 2: dev ────────────────────────────────────────────────
FROM base AS dev
WORKDIR /moneymate/services/auth

# install air for hot reload
RUN go install github.com/air-verse/air@latest

# air watches for changes and rebuilds
CMD ["air", "-c", ".air.toml"]

# ── Stage 3: builder ────────────────────────────────────────────
FROM base AS builder
WORKDIR /moneymate

# copy all source
COPY shared/ shared/
COPY services/auth/ services/auth/

# build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /bin/auth \
    ./services/auth/cmd/main.go

# ── Stage 4: prod ───────────────────────────────────────────────
FROM alpine:3.19 AS prod

# security: don't run as root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /bin/auth /bin/auth

EXPOSE 8081

CMD ["/bin/auth"]