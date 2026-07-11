# TradePilot AI

TradePilot AI is an AI-assisted algorithmic trading platform with a Go backend, Chi router, SMC broker SDK, Groq integration, and a React frontend.

## Local Development

```bash
cd backend
go test ./...
go run ./cmd/server
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

## Production

Use `docker-compose.yml` for local server, VPS, home server, DigitalOcean, AWS, or Oracle Cloud deployments. Configure secrets through environment variables, especially `JWT_SECRET`, `DATABASE_URL`, `GROQ_API_KEY`, and broker credentials.

## Architecture

Business logic lives in `backend/internal/core`. HTTP handlers live in `backend/internal/api`. Broker-specific code remains in `backend/internal/broker`. Database details are isolated in `backend/internal/database`. AI and notifications use provider interfaces so implementations can be swapped without changing trading logic.
