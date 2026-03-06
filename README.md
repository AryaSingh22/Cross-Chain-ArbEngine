# Cosmos ArbEngine Pro 

**Cross-Chain Arbitrage Monitor + IBC Relay Dashboard**
Built natively for the Cosmos / IBC ecosystem.
 
## Architecture

```
┌─────────────────────────────┐
│       React Frontend        │  Port 3000
│  Dashboard │ Relay │ Charts │
└──────────────┬──────────────┘
               │ REST / WebSocket
┌──────────────▼──────────────┐
│         Go Backend          │  Port 8080
│  Price Feeds │ Arb Engine   │
│  IBC Monitor │ Event Bus    │
│  REST API   │ WS Broadcast  │
└──────┬──────────────┬───────┘
       │              │
  TimescaleDB      Redis
   Port 5432      Port 6379
```

## Features

- **Real-time arbitrage detection** across 7 Cosmos chains (Osmosis, Injective, Neutron, Stride, Juno, Cosmos Hub, Akash)
- **17+ arbitrage paths** monitored for ATOM, OSMO, INJ, NTRN, and more
- **IBC relay health monitoring** with channel status matrix
- **Live WebSocket updates** with opportunity feed table
- **Dark-mode trading dashboard** with sortable tables, path diagrams, fee breakdowns
- **Analytics** with spread charts, volume charts, and CSV export
- **Alert configuration** for webhook, Telegram, and in-app notifications

## Quick Start

### Docker Compose (Recommended)

```bash
docker-compose up -d
```

Dashboard: `http://localhost:3000`
API: `http://localhost:8080/api/v1/opportunities`

### Local Development

**Backend** (requires Go 1.21+):
```bash
cd backend
go mod tidy
go run ./cmd/server
```

**Frontend** (requires Node 18+):
```bash
cd frontend
npm install
npm run dev
```

**Infrastructure**:
```bash
# Start TimescaleDB and Redis
docker-compose up -d timescaledb redis
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/opportunities` | Live arbitrage opportunities |
| GET | `/api/v1/opportunities/history` | Historical opportunities |
| GET | `/api/v1/opportunities/export` | CSV export |
| GET | `/api/v1/chains` | Chain status |
| GET | `/api/v1/chains/prices` | Current prices |
| GET | `/api/v1/relay/channels` | IBC channel health |
| GET | `/api/v1/relay/channels/:id/events` | Channel events |
| WS  | `/ws/opportunities` | Live opportunity stream |

## Tech Stack

- **Backend**: Go, Gin, pgx, gorilla/websocket, shopspring/decimal
- **Frontend**: React 18, TypeScript, Vite, Zustand, TanStack Query, Recharts, Tailwind CSS
- **Database**: TimescaleDB (PostgreSQL + time-series)
- **Cache**: Redis
- **Infra**: Docker Compose

## Config

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://arbengine:arbengine_dev@localhost:5432/arbengine?sslmode=disable` | TimescaleDB connection |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection |
| `SERVER_PORT` | `8080` | REST API port |
| `USE_MOCK_FEEDS` | `true` | Use simulated price data |
| `MIN_NET_PROFIT_USD` | `5.0` | Min profit threshold |
| `CORS_ORIGIN` | `http://localhost:3000` | CORS allowed origin |

## License

MIT
