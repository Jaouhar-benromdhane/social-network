FROM golang:1.23-bookworm AS builder

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations
RUN mkdir -p /app/data

ENV APP_PORT=8080
ENV DB_PATH=/app/data/social_network.db
ENV MIGRATIONS_PATH=file:///app/pkg/db/migrations/sqlite

EXPOSE 8080
CMD ["./server"]
