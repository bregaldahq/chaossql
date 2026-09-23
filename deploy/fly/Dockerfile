# syntax=docker/dockerfile:1
# Managed ChaosSQL Cloud image: API + dashboard on one origin, with continuous
# SQLite replication through Litestream. Build from the repository root:
#   docker build -f deploy/fly/Dockerfile -t chaossql-cloud .

FROM golang:1.25-alpine AS server
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/chaossql ./cmd/chaossql

FROM node:20-alpine AS site
WORKDIR /app
COPY site/package*.json ./
RUN npm ci
COPY site/ ./
RUN npm run build \
    && mkdir -p /out \
    && cp index.html /out/ \
    && cp -r assets brand /out/ \
    && cp -r public/wasm /out/wasm

FROM alpine:3.20 AS litestream
ARG LITESTREAM_VERSION=0.5.17
ARG LITESTREAM_SHA256=cfb371176d164437ae869f8351cfde49bd1804ae71c61923f75c9cba9c9c006d
RUN apk add --no-cache curl \
    && curl -fsSL -o /tmp/litestream.tar.gz \
       "https://github.com/benbjohnson/litestream/releases/download/v${LITESTREAM_VERSION}/litestream-${LITESTREAM_VERSION}-linux-x86_64.tar.gz" \
    && echo "${LITESTREAM_SHA256}  /tmp/litestream.tar.gz" | sha256sum -c - \
    && tar -xzf /tmp/litestream.tar.gz -C /usr/local/bin litestream

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget su-exec \
    && addgroup -g 10001 -S chaossql \
    && adduser -u 10001 -S chaossql -G chaossql -h /home/chaossql \
    && mkdir -p /data \
    && chown -R chaossql:chaossql /data

COPY --from=server /out/chaossql /usr/local/bin/chaossql
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream
COPY --from=site /out/ /srv/site/
COPY deploy/fly/litestream.yml /etc/litestream.yml
COPY deploy/fly/run.sh /usr/local/bin/chaossql-cloud-run
COPY deploy/fly/admin.sh /usr/local/bin/chaossql-admin
COPY deploy/fly/restore-check.sh /usr/local/bin/chaossql-restore-check
RUN chmod 0755 /usr/local/bin/chaossql-cloud-run /usr/local/bin/chaossql-admin /usr/local/bin/chaossql-restore-check

USER 10001:10001
WORKDIR /home/chaossql

ENV PORT=8080 \
    DB_PATH=/data/chaossql-cloud.db \
    STATIC_DIR=/srv/site \
    CHAOSSQL_ENFORCE_RETENTION=true

EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/v1/health || exit 1

ENTRYPOINT ["/bin/sh", "/usr/local/bin/chaossql-cloud-run"]
