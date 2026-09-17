# ChaosSQL Enterprise: Self-Hosted & On-Premise Deployment Guide

This guide details how to deploy, configure, and operate the **ChaosSQL SaaS Control Plane & Concurrency Gate** in your private cloud, on-premise infrastructure, or air-gapped data centers.

---

## 1. Architectural Overview

ChaosSQL Enterprise is designed as a lightweight, high-performance concurrency observability engine:

- **ChaosSQL Server (`chaossql-server` / `chaossql server`):** Core API service written in Go. Handles execution ingestion, regression evaluation, baseline diffing, real-time alert dispatching (Discord, Slack, Datadog), and token authentication.
- **ChaosSQL Web Dashboard:** Modern React/Vite dashboard providing interactive causal trace inspection, SQL step timelines, Delta-Debugging visualization, and webhook management.
- **Storage Layer:** Embedded SQLite with WAL mode by default for zero-ops durability and single-binary portability; supports external volume mounting and automated snapshots.
- **Worker Execution:** ChaosSQL CLI runs in your CI/CD pipelines (GitHub Actions, GitLab CI, CircleCI) and securely publishes execution metadata to your private server.

```
┌────────────────────────────────────────────────────────┐
│                   CI / CD Pipelines                    │
│   (GitHub Actions / GitLab CI: chaossql run --spec)   │
└──────────────────────────┬─────────────────────────────┘
                           │ POST /v1/runs (Metadata only)
                           ▼
┌────────────────────────────────────────────────────────┐
│              ChaosSQL Control Plane Server             │
│        (Regressions Engine + Baselines Store)          │
└──────────────┬──────────────────────────┬──────────────┘
               │                          │
               ▼                          ▼
┌────────────────────────────┐ ┌─────────────────────────┐
│     Incident Webhooks      │ │   Internal Dashboard    │
│ (Slack / Discord / Alerts) │ │ (Interleaving Viewer)   │
└────────────────────────────┘ └─────────────────────────┘
```

---

## 2. Fast-Track: Single-Node with Docker Compose

Deploy the complete enterprise stack (PostgreSQL 16 test target + ChaosSQL Control Plane + Web Dashboard) in 30 seconds:

### Step 1: Clone and Configure Environment

```bash
git clone https://github.com/bregaldahq/chaossql.git
cd chaossql

cp .env.example .env
```

### Step 2: Launch the Stack

```bash
docker compose up -d
```

### Step 3: Verify Container Health

```bash
docker compose ps
```

Expected output:
```
NAME                         IMAGE               STATUS                    PORTS
chaossql-server             chaossql-server     Up (healthy) 8080->8080/tcp
chaossql-dashboard          chaossql-dashboard  Up (healthy) 3000->80/tcp
chaossql-postgres-target    postgres:16-alpine  Up (healthy) 5432->5432/tcp
```

### Step 4: Access the Services
- **Web Dashboard:** `http://localhost:3000/#/dashboard`
- **Control Plane API:** `http://localhost:8080/v1/health`
- **PostgreSQL Database:** `localhost:5432` (User: `postgres`, Pass: `postgres`, DB: `test`)

---

## 3. Kubernetes Deployment with Helm

ChaosSQL provides an official Helm chart located in `charts/chaossql-server/`.

### Step 1: Review Chart Values

Inspect `charts/chaossql-server/values.yaml`:

```yaml
replicaCount: 1

image:
  repository: ghcr.io/bregaldahq/chaossql-server
  tag: "1.5.0"
  pullPolicy: IfNotPresent

persistence:
  enabled: true
  size: 10Gi
  storageClass: "gp3" # or standard

secrets:
  adminToken: "your_strong_enterprise_secret_token"
  discordWebhookUrl: ""
```

### Step 2: Install via Helm

```bash
# From the repository root
helm upgrade --install chaossql-server ./charts/chaossql-server \
  --namespace chaossql \
  --create-namespace \
  --set secrets.adminToken="super_secret_admin_token"
```

### Step 3: Verify Pod and Service Status

```bash
kubectl get pods -n chaossql
kubectl get svc -n chaossql
```

### Step 4: Expose via Ingress (Optional)

Enable ingress in `values.yaml` or set via command line:

```bash
helm upgrade chaossql-server ./charts/chaossql-server \
  --namespace chaossql \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=chaossql.internal.mycompany.com
```

---

## 4. Standalone CLI Server Mode

For local developers, staging environments, or air-gapped workstations, the `chaossql` binary can run the server directly with zero external dependencies:

### Start the Server

```bash
# Start server on port 8080 using a local SQLite file
chaossql server start --port=8080 --db=/var/data/chaossql.db --token=my_secret_token
```

### Generate Organization CI Tokens

```bash
chaossql server create-token --db=/var/data/chaossql.db --org=payments-team --name="Payments CI Token"
```

Output:
```
=== ChaosSQL API Token Created ===
Organization: payments-team
Token Name:   Payments CI Token
API Token:    csql_8f9a2b1c4e7d0f3a6b8e2c5d
===================================
```

---

## 5. Enterprise Hardening & Best Practices

### Non-Root Execution
The production Docker container runs under UID `10001` (`chaossql`) with `allowPrivilegeEscalation: false` and all Linux capabilities dropped.

### Storage & Backups
SQLite database data resides in `/data/chaossql-cloud.db`. To take an online hot-backup without stopping the service:

```bash
# Using SQLite backup API or CLI
sqlite3 /data/chaossql-cloud.db ".backup '/backups/chaossql-backup-$(date +%Y%m%d%H%M%S).db'"
```

### Air-Gapped Environments
For offline environments without internet access:
1. Mirror the images:
   ```bash
   docker pull ghcr.io/bregaldahq/chaossql-server:1.5.0
   docker save ghcr.io/bregaldahq/chaossql-server:1.5.0 -o chaossql-server.tar
   ```
2. Load images into your private container registry (Harbor, AWS ECR, Artifactory).
3. Set `image.repository` in `values.yaml`.

---

## 6. CI/CD Pipeline Integration

Configure your CI pipelines to point to your self-hosted server:

### GitHub Actions

```yaml
name: Concurrency Gate

on:
  pull_request:
    branches: [ main ]

jobs:
  concurrency-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1.5.0
        with:
          spec-path: 'chaos.yaml'
          cloud-url: 'https://chaossql.internal.mycompany.com'
          cloud-token: ${{ secrets.CHAOSSQL_SELFHOSTED_TOKEN }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
          post-pr-comment: 'true'
```

### GitLab CI

```yaml
concurrency_test:
  stage: test
  image: ghcr.io/bregaldahq/chaossql:1.5.0
  script:
    - chaossql run chaos.yaml \
        --cloud-url="https://chaossql.internal.mycompany.com" \
        --cloud-token="$CHAOSSQL_SELFHOSTED_TOKEN"
```
