# Tyr Fintech

Tyr Fintech is a high-performance multi-currency digital wallet, B2B payment gateway, and hosted checkout platform. Built with a Go (Gin-Gonic) backend, it ensures strict transactional consistency, balanced double-entry ledger bookkeeping, and field-level encryption, paired with a modern React 19 glassmorphic single-page application (SPA) featuring real-time Server-Sent Events (SSE).

---

## Quick Recruiter Access & Live Demos

To explore the application without manual registration, use the pre-seeded credentials or the 1-click login buttons:

* **Email**: `demo@tyr.com`
* **Password**: `demo123456`
* **Pre-funded Balances**: 25,000.00 TRY, 1,500.00 USD, 1,000.00 EUR
* **Virtual Cards**: Active virtual Visa card with instant freeze/unfreeze controls.
* **Foodeli API Key**: `tyr_live_foodeli_secret_key_12345` (Pre-seeded B2B merchant key).

---

## Key Architectural Highlights

### 1. B2B Payment Gateway & Double-Entry Ledger (Stripe/Iyzico-style)
* **Direct Card Charges (`POST /api/v1/charges`)**: Accepts minor currency units (`int64`), verifies Luhn checksum, validates test card tokens, and prevents duplicate billing via HTTP `Idempotency-Key` caching.
* **Balanced Double-Entry Bookkeeping**: For every payment, balanced debits and credits are executed atomically in PostgreSQL (`pgx.Tx`):
  ```text
  Total Debits = Total Credits
  ACQUIRING_CLEARING (Debit) = MERCHANT_PAYABLE (Credit) + PLATFORM_FEE_REVENUE (Credit)
  ```

### 2. Hosted Checkout Engine ("Pay with TyrFintech")
* **Checkout Sessions (`POST /api/v1/checkout/sessions`)**: Generates secure checkout sessions with custom TTL, merchant fee split, and callback redirect URLs.
* **Hosted Payment UI (`/checkout?session_id=...`)**: Dedicated checkout screen where consumers log in, select currency wallets, approve payments, and redirect back to merchant callbacks (e.g., Foodeli).

### 3. Durable RabbitMQ Webhook Dispatcher & Dead-Letter Queue (DLQ)
* **Asynchronous Webhook Pipeline**: Webhook notifications are published with persistent delivery mode (`amqp.Persistent`) to RabbitMQ topic `webhook.merchant.event`.
* **Exponential Backoff Retries**: Automatically retries failed webhook deliveries up to 3 times with exponential backoff (2s, 4s, 6s).
* **Dead-Letter Queue (DLQ)**: Poisoned or permanently failing webhook messages are diverted to `merchant_webhooks_dlq` via RabbitMQ DLX (`fintech.events.dlx`) for audit and manual replay.

### 4. Enterprise Security & CORS Whitelisting
* **Field-Level AES-256-GCM Encryption**: Primary Account Numbers (PAN) and CVVs are encrypted at rest with unique 12-byte nonces before storage in PostgreSQL (`pkg/encryption`).
* **Origin Whitelisting**: Configurable CORS middleware restricting cross-origin requests to trusted production origins while permitting `localhost` during local development.
* **Row-Level Locking**: Pessimistic `SELECT ... FOR UPDATE` locks prevent double-spending and race conditions during wallet debits.

---

## Tech Stack

### Backend
* **Language & Runtime**: Go 1.26+
* **HTTP Framework**: Gin-Gonic
* **Database & Pooling**: PostgreSQL 16 via `pgx/v5` connection pool
* **Caching & Rate Limiting**: Redis 7 (sliding token bucket limiter + exchange rate cache wrapper)
* **Message Broker**: RabbitMQ 3 (Topic exchanges, durable queues, DLX/DLQ)
* **Security & Crypto**: AES-256-GCM, Bcrypt, JWT (HMAC-SHA256)
* **Observability**: Prometheus metrics exporter (`/metrics`)

### Frontend
* **Build Tool**: Vite
* **Framework**: React 19
* **Data Fetching & Cache**: TanStack Query (React Query v5)
* **Styling**: Tailwind CSS v4 (Glassmorphic dark UI)
* **Real-Time Stream**: Server-Sent Events (SSE) with auto-reconnect

### DevOps & Infrastructure
* **Containerization**: Docker Compose with multi-stage production builds
* **Port Isolation**: Isolated port mapping designed to run side-by-side with external microservices (e.g. Foodeli) without host conflicts.

---

## Port Matrix & Local Infrastructure

| Service | Host Port | Internal Port | URL / Access |
| :--- | :--- | :--- | :--- |
| **Frontend (Vite / Nginx)** | `3005` | `80` / `3005` | [http://localhost:3005](http://localhost:3005) |
| **Backend API (Go Gin)** | `8081` | `8081` | [http://localhost:8081](http://localhost:8081) |
| **PostgreSQL** | `55432` | `5432` | `localhost:55432` (`admin`/`secretpassword`) |
| **Redis** | `56379` | `6379` | `localhost:56379` |
| **RabbitMQ (AMQP)** | `5673` | `5672` | `localhost:5673` |
| **RabbitMQ Management** | `15673` | `15672` | [http://localhost:15673](http://localhost:15673) (`guest`/`guest`) |
| **Prometheus** | `9090` | `9090` | [http://localhost:9090](http://localhost:9090) |
| **Grafana** | `3001` | `3000` | [http://localhost:3001](http://localhost:3001) (`admin`/`admin`) |

---

## Project Structure

```
├── Makefile                     # Root developer task runner (make api, make web, etc.)
├── docker-compose.yml           # Isolated multi-container environment
├── progress.md                  # Implementation roadmap and production audit log
├── backend/
│   ├── cmd/api/main.go          # Application bootstrapper and dependency injector
│   ├── config/                  # Dynamic environment configuration (.env loader)
│   ├── internal/
│   │   ├── db/                  # PostgreSQL pool, schema seeders, and migration helpers
│   │   ├── dto/                 # Request/Response data transfer objects
│   │   ├── handlers/            # HTTP controllers (charges, checkout, wallets, auth)
│   │   ├── middleware/          # CORS whitelist, auth guards, Redis rate limiter
│   │   ├── models/              # Domain models (Merchant, Charge, Ledger, CheckoutSession)
│   │   ├── notifications/       # SSE streaming Hub
│   │   ├── queue/               # RabbitMQ publisher and exchange declarations
│   │   ├── repos/               # PostgreSQL repositories with ACID transactions
│   │   ├── services/            # Core business domain, card validator, checkout service
│   │   └── worker/              # RabbitMQ consumers (DLQ webhook dispatcher, event consumer)
│   ├── migrations/              # Versioned SQL migrations (000001 - 000003)
│   ├── pkg/                     # Utilities (AES-256-GCM encryption, JWT, PDF/CSV export)
│   └── scripts/                 # Automated end-to-end integration test suites
└── frontend/
    ├── src/
    │   ├── components/          # Reusable UI components (Wallets, Cards, Modals)
    │   ├── context/             # Authentication state context
    │   ├── hooks/               # TanStack Query and SSE stream hooks
    │   ├── lib/axios.js         # Axios HTTP client with credentials
    │   └── pages/
    │       ├── Dashboard.jsx    # User wallet and card management dashboard
    │       ├── Login.jsx        # User login and registration screen
    │       └── Checkout.jsx     # Hosted Checkout ("Pay with TyrFintech") UI
    ├── Dockerfile               # Multi-stage production Nginx container
    └── vite.config.js           # Vite dev configuration (port 3005)
```

---

## Getting Started

### 1. Start Infrastructure via Docker

Start PostgreSQL, Redis, and RabbitMQ:

```bash
make docker-up
# or: docker compose up -d
```

### 2. Run the Applications

You can run the backend and frontend locally with hot-reloading:

```bash
# Terminal 1: Backend API (port 8081)
make api

# Terminal 2: Frontend Vite Dev Server (port 3005)
make web
```

Or run all services (including backend and frontend) directly inside Docker:

```bash
docker compose up -d --build
```

---

## Automated Testing & Verification

### Unit & Service Tests

Run unit tests across services, encryption, and repositories:

```bash
make test
```

### End-to-End Gateway & Checkout Suites

TyrFintech includes automated shell test suites verifying end-to-end wire contracts:

```bash
# 1. Test B2B Direct Payment Gateway (Charges, Idempotency, 402/400 Declines, 401 Auth)
make test-gateway

# 2. Test Hosted Checkout Flow (Session creation, UI summary, Wallet payment, RabbitMQ webhook)
make test-checkout
```

---

## Production Deployment Checklist

When deploying to production:
1. Set `ENV=production` in container environment variables.
2. Provide cryptographically secure random values for `JWT_SECRET` and `CARD_ENCRYPTION_KEY` (32 bytes).
3. Whitelist production client domains using `ALLOWED_ORIGINS=https://checkout.yourdomain.com,https://merchant.example.com`.
4. Replace `MockExchangeService` with an external live Forex API.
5. In production card gateway flows, connect licensed Bank Virtual POS (VPOS) / 3D Secure v2 providers.

---

## License

MIT License. Developed for technical demonstration, portfolio evaluation, and distributed systems pair-programming.
