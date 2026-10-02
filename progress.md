# TyrFintech Production Deployment & Architecture Progress

This document tracks system progress, architectural decisions, and the essential checklist required before taking **TyrFintech** to production.

---

## 1. High-Concurrency & Webhook Resilience (RabbitMQ vs. In-Memory Channels)

### Current State (MVP / Local Dev)
- Currently, webhooks use Go buffered channels (`var MerchantWebHookQueue = make(chan *dto.MerchantWebhookEvent, 100)`).
- **The Bottleneck at Scale (e.g., 1,000 concurrent payments):**
  - **Memory Buffer Overflow:** If Foodeli sends 1,000 concurrent payments, the buffer of 100 fills up. The remaining goroutines block or consume memory.
  - **Zero Crash Resilience:** If the backend pod/container restarts while 500 webhooks are pending in the Go channel, those events are **lost forever**. Foodeli will never receive payment confirmation.
  - **No Retry / Backoff:** If Foodeli's server is temporarily down (HTTP 502/503), in-memory dispatch cannot easily implement exponential backoff across server restarts.

### Production Solution: RabbitMQ Durable Exchange & Worker
Before launching to production:
1. **Durable RabbitMQ Exchange:**
   - Publish webhook events to a RabbitMQ direct/topic exchange: `events.merchant.webhook` with `DeliveryMode = amqp091.Persistent` (written to disk).
2. **Dead Letter Queue (DLQ) & Exponential Backoff:**
   - Failed webhook delivery attempts should be routed to a retry exchange with TTL backoff (e.g., retry after 30s, 2m, 10m, 1h).
   - If Foodeli fails to acknowledge after 5 attempts, route to a DLQ (`webhook.failed.dlq`) for merchant alerting.
3. **HMAC Webhook Signatures:**
   - Sign outgoing webhooks with a merchant-specific secret:
     `Header: Tyr-Signature: t=<timestamp>,v1=<hmac_sha256(payload, secret)>`
   - Prevents replay attacks and allows Foodeli to cryptographically verify event authenticity.

---

## 2. Hardcoded URLs & Configuration Strategy

### Current Status
- Eliminated all hardcoded strings.
- `FRONTEND_URL`: Injected into `config.Config` and defaults to `http://localhost:3005`, overridable via `.env`.
- `PORT` / `API_PORT`: Configurable via `PORT` (for Heroku, Render, AWS ECS, GCP Cloud Run) or `API_PORT`, defaulting to `8081`.

### Multi-Project Port Isolation Matrix (Foodeli vs TyrFintech)
To ensure **Foodeli** and **TyrFintech** can run concurrently on a developer's workstation without `bind: address already in use` conflicts:

| Service | Foodeli Host Port | TyrFintech Host Port | Conflict Resolved |
| :--- | :--- | :--- | :--- |
| **PostgreSQL** | `5432` | `55432` | ✅ Isolated |
| **Backend API** | `8080` (Go Server) | `8081` (Gateway) | ✅ Isolated |
| **Frontend Web** | `3000` (Next.js) | `3005` (Vite SPA) | ✅ Isolated |
| **Redis** | `6379` | `56379` (Docker mapped) | ✅ Isolated |
| **RabbitMQ AMQP** | `5672` | `5673` (Docker mapped) | ✅ Isolated |
| **RabbitMQ Mgmt** | `15672` | `15673` (Docker mapped) | ✅ Isolated |
| **Grafana** | *(None)* | `3001` | ✅ Isolated |
| **Prometheus** | *(None)* | `9090` | ✅ Isolated |

### Production Environment Variables Checklist
| Variable | Production Example | Purpose |
| :--- | :--- | :--- |
| `FRONTEND_URL` | `https://checkout.tyrfintech.com` | Base URL used to build user redirect checkout sessions |
| `API_HOST` | `0.0.0.0` | Listen host interface |
| `PORT` | `8081` (or assigned by orchestrator) | HTTP server port |
| `DATABASE_URL` | `postgres://user:pass@aurora-pg.cluster...:5432/fintech?sslmode=verify-full` | Production PostgreSQL connection pool |
| `REDIS_URL` | `redis://user:pass@redis-cluster...:6379` | Distributed rate limiting and exchange rate caching |
| `RABBITMQ_URL` | `amqp://user:pass@rabbitmq-cluster...:5672/` | Durable message bus for asynchronous event processing |
| `JWT_SECRET` | 64-character high-entropy secret | User session token signing |
| `CARD_ENCRYPTION_KEY`| 32-byte AES-256-GCM encryption key | Encryption of stored card PAN and CVV at rest |

---

## 3. Database & Ledger Invariants

1. **Integer Minor Units:** All financial calculations strictly use `int64` minor currency units (kuruş, cents). Floating-point arithmetic (`float32`/`float64`) is strictly prohibited.
2. **Double-Entry Ledger Balancing:**
   $$\sum \text{Debits} = \sum \text{Credits}$$
   Every successful charge and checkout payment writes balanced entries inside an atomic database transaction (`pgx.Tx`).
3. **Row-Level Locking:** All wallet balance deductions and checkout session completions use `SELECT ... FOR UPDATE` to prevent race conditions and double-spending.

---

## 4. Completed Milestones

- [x] B2B Direct Payment Gateway (`POST /api/v1/charges`) with exact Foodeli wire contract.
- [x] Merchant API Key Authentication Middleware (`Authorization: Bearer tyr_*`).
- [x] Deterministic Card Simulation & Luhn Algorithm Validation.
- [x] Double-Entry Ledger Settlement (`ACQUIRING_CLEARING`, `MERCHANT_PAYABLE`, `PLATFORM_FEE_REVENUE`).
- [x] HTTP-level Idempotency Layer with cached JSON response replay.
- [x] Hosted Checkout Session Schema (`checkout_sessions` table with TTL & status).
- [x] Hosted Checkout Repository & Service (`CreateSession`, `GetSessionDetails`, `PaySession`).
- [x] Asynchronous Merchant Webhook Queue (`worker.MerchantWebHookQueue`).
- [x] Docker Multi-Service Port Isolation (Postgres 55432, Redis 56379, RabbitMQ 5673) with zero-panic fallback.
- [x] End-to-end integration verified: `./scripts/test_gateway.sh` and `./scripts/test_hosted_checkout.sh` passing 100%.

---

## 5. Production Readiness & Dynamic Configuration Audit

### A. Environment Variables & Dynamic Overrides (Resolved)
- `ENV`: Set to `production` in live deployments. When `production`, demo user seeding (`demo@tyr.com`) and default merchant keys are disabled.
- `DEFAULT_WEBHOOK_URL`: Replaced legacy hardcoded `webhook.site` link with an environment variable fallback.
- `FOODELI_LIVE_API_KEY` & `FOODELI_TEST_API_KEY`: Can be injected via env variables instead of defaulting to hardcoded keys.
- `FRONTEND_URL`: Injects the public domain of the checkout page (`https://checkout.tyrfintech.com`).
- `VITE_API_BASE_URL`: Frontend now consistently points to `http://localhost:8081/api/v1` in dev, overridden by build env in production (`https://api.tyrfintech.com/api/v1`).
- `REDIS_URL` / `REDIS_ADDR`: Backend dynamically accepts both variable formats.

### B. Critical Architecture Updates Implemented (Portfolio Highlights)
1. **Durable RabbitMQ Webhook Queue with Retry & DLQ:**
   - Persistent message delivery (`amqp.Persistent`) published to topic `webhook.merchant.event`.
   - Consumer with QoS prefetch (10 concurrent).
   - Exponential backoff retry mechanism (attempts 1 to 3 with backoff delays 2s, 4s, 6s).
   - Dead-Letter Exchange (`fintech.events.dlx`) and Queue (`merchant_webhooks_dlq`) for poison/failed messages after 3 attempts.
   - Dual-dispatch fallback (in-memory + RabbitMQ) to guarantee 100% test compatibility.
2. **CORS Origin Whitelisting:**
   - Dedicated [`CORSMiddleware`](file:///wsl.localhost/Ubuntu/home/iamtbay/projects/tyr-fintech/backend/internal/middleware/cors.go).
   - Allowed origins strictly configurable via `ALLOWED_ORIGINS` environment variable.
   - Localhost (`localhost:3005`, `localhost:3000`) permitted in development mode.
   - Unauthorized origins rejected with HTTP 403.
3. **Simulated Card & Mock FX Portfolio Rationale:**
   - Card validation uses deterministic test card simulator and Luhn checksum.
   - FX rates provide deterministic currency conversions for wallet multi-currency demonstrations without third-party API dependencies.
