# Praxis Explorer Stack

This repository packages the Praxis explorer backend, web UI, and local Postgres database into a standalone stack.

## Structure

- `backend/` – Go API service and ERC-8004 indexer (build with `Dockerfile`).
- `frontend/` – Next.js UI that talks to the explorer API.
- `docker-compose.yml` – Spins up Postgres, runs migrations, launches the API and UI.

## Getting Started

1. Export the required RPC endpoint for on-chain indexing (optional for local demo):
   ```bash
   export SEPOLIA_RPC="https://..."
   ```
2. Start the stack:
   ```bash
   docker compose up --build
   ```
3. The explorer API is available at http://localhost:8080 and the UI at http://localhost:3100.

The database data is persisted in the named Docker volume `explorer_db_data` and migrations are applied automatically on startup.
