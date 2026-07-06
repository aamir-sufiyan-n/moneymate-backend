
FROM golang:1.26.4-alpine AS builder
WORKDIR /moneymate-backend


FROM builder AS dev


RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]


FROM builder AS prod-builder
COPY gateway/go.mod gateway/go.sum ./gateway/
WORKDIR /moneymate-backend/gateway
RUN go mod download

WORKDIR /moneymate-backend
COPY shared/ ./shared/
COPY db/ ./db/
COPY gateway/ ./gateway/

WORKDIR /moneymate-backend/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/main.go


FROM alpine:latest AS prod
WORKDIR /app/

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=prod-builder /server .

USER appuser


EXPOSE 8081

CMD ["./server"]