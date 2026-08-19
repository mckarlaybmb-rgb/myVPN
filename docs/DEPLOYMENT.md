# Deployment - VPN OS v1

## Development

From the repository root:

```bash
docker-compose up --build
```

The API listens on port `8080` and PostgreSQL on port `5432`. Compose waits for the PostgreSQL health check before starting the backend. The backend runs SQL migrations from `backend/migrations` during startup.

## Production Guidance

- Build and deploy the multi-stage backend image without mounting source code.
- Inject `DATABASE_URL` and other secrets through a secret manager.
- Use TLS and authenticated administrative access before exposing the API publicly.
- Restrict PostgreSQL and Xray administration to trusted network paths.
- Add readiness checks and backups before operating persistent production data.
