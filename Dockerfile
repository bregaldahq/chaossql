# syntax=docker/dockerfile:1

# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binaries
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=1.5.0" -o /bin/chaossql-server ./cmd/chaossql-server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=1.5.0" -o /bin/chaossql ./cmd/chaossql

# Stage 2: Production runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -g 10001 -S chaossql \
    && adduser -u 10001 -S chaossql -G chaossql -h /home/chaossql \
    && mkdir -p /data \
    && chown -R chaossql:chaossql /data

COPY --from=builder /bin/chaossql-server /usr/local/bin/chaossql-server
COPY --from=builder /bin/chaossql /usr/local/bin/chaossql

USER 10001:10001

WORKDIR /home/chaossql
VOLUME ["/data"]

ENV PORT=8080 \
    DB_PATH=/data/chaossql-cloud.db \
    CHAOSSQL_ADMIN_TOKEN=chaossql_prod_secret \
    PUBLIC_URL=http://localhost:8080

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/v1/health || exit 1

ENTRYPOINT ["/usr/local/bin/chaossql-server"]
CMD ["start"]
