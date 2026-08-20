cd /workspaces/myVPN

# Create directories
mkdir -p backend bot postgres docs scripts deployments .github/workflows

# README.md
cat > README.md <<'EOF'
# VPN OS v1 (vpnos)

VPN OS v1 is a production-ready backend control plane for Xray-core VPN nodes. It manages VLESS + REALITY + XTLS Vision users dynamically via a durable PostgreSQL-backed job queue, background workers with retries and dead-letter handling, and a small REST API for control-plane operations.

See docs/ for architecture and deployment details.
EOF

# .env.example
cat > .env.example <<'EOF'
# PostgreSQL
DATABASE_URL=postgres://vpnos:vpnos_pass@postgres:5432/vpnos?sslmode=disable

# Backend
BACKEND_PORT=8080
ADMIN_API_KEY=supersecretapikey
LOG_LEVEL=info

# x-ui integration
XUI_BASE_URL=http://x-ui:2053
XUI_USERNAME=admin
XUI_PASSWORD=CHANGE_ME
XUI_INBOUND_ID=1

# Retry config
XRAY_RETRY_COUNT=3
XRAY_BACKOFF_BASE_MS=50
EOF

# CONTRIBUTING.md
cat > CONTRIBUTING.md <<'EOF'
# Contributing to vpnos

Branching policy:
- main: production-ready
- develop: integration branch
- feature/<name>: from develop
- hotfix/<name>: from main

Commit messages:
Use Conventional Commits: type(scope): subject

Pull request process:
- Target develop unless hotfix
- CI must pass
- At least one reviewer

Testing:
- Unit tests for new code
- Integration tests for cross-component behavior (Testcontainers)
- Run: go test ./... in backend

Security:
- Never commit secrets. Use .env.example as template.
- Use a secrets manager (Vault, cloud SM) in production.
EOF

# PROJECT_STATUS.md
cat > PROJECT_STATUS.md <<'EOF'
# Project Status — VPN OS v1

Completed:
- Backend Core (Go REST API)
- PostgreSQL integration
- User Management
- Xray-core integration (HandlerService add/remove)
- VLESS + REALITY + XTLS Vision URI generation
- Subscription Management
- Durable PostgreSQL Job Queue
- Xray AddUser/RemoveUser Worker with retry + DLQ
- Health Checks and Prometheus metrics
- Integration tests

Pending:
- Payment / Billing
- Telegram Bot
- Admin Dashboard
- Persistent Xray API client (Phase 2)
- Production security hardening (mTLS, Vault)
- Multi-node orchestration
EOF

# CHANGELOG.md
cat > CHANGELOG.md <<'EOF'
# Changelog

## Unreleased
- Initial canonical repository snapshot for VPN OS v1:
  - Backend, job queue, subscription system, Xray integration, tests, and docs scaffold.
EOF

# docker-compose.yml
cat > docker-compose.yml <<'EOF'
version: '3.8'
services:
  postgres:
    image: postgres:15
    restart: unless-stopped
    environment:
      POSTGRES_USER: vpnos
      POSTGRES_PASSWORD: vpnos_pass
      POSTGRES_DB: vpnos
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build:
      context: ./backend
    environment:
      - DATABASE_URL=postgres://vpnos:vpnos_pass@postgres:5432/vpnos?sslmode=disable
      - BACKEND_PORT=8080
      - ADMIN_API_KEY=CHANGE_ME
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    volumes:
      - ./backend:/src
    restart: unless-stopped

volumes:
  pgdata:
EOF

# CI workflow (GitHub Actions)
mkdir -p .github/workflows
cat > .github/workflows/ci.yml <<'EOF'
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: vpnos
          POSTGRES_PASSWORD: vpnos_pass
          POSTGRES_DB: vpnos
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install dependencies
        run: go mod download
        working-directory: backend

      - name: gofmt check
        run: |
          go list ./... | xargs -n1 gofmt -l | tee /tmp/gofmt.txt
          if [ -s /tmp/gofmt.txt ]; then cat /tmp/gofmt.txt; exit 1; fi
        working-directory: backend

      - name: go vet
        run: go vet ./...
        working-directory: backend

      - name: run tests
        env:
          DATABASE_URL: postgres://vpnos:vpnos_pass@127.0.0.1:5432/vpnos?sslmode=disable
        run: |
          for i in `seq 1 10`; do pg_isready -h 127.0.0.1 && break || sleep 1; done
          go test ./... -v
        working-directory: backend
EOF
