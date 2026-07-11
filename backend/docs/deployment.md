# Deployment Guide

## Targets

- Local PC: run backend and frontend directly.
- Home server/VPS/DigitalOcean/Oracle Cloud/AWS: run Docker Compose behind a TLS reverse proxy.

## Required Secrets

- `JWT_SECRET`
- `DATABASE_URL`
- `GROQ_API_KEY`
- `SMC_API_KEY`
- `SMC_API_SECRET`
- `SMC_CLIENT_ID`
- `SMC_PASSWORD`

## Health Checks

- `GET /api/v1/health`
- `GET /api/v1/version`
- `GET /api/v1/system/metrics`

## Database

Run SQL files in `backend/migrations` before switching from in-memory development mode to PostgreSQL-backed production repositories.
