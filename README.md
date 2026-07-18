# Tyr Fintech

Tyr Fintech is a modern, high-performance, and secure multi-currency digital wallet and virtual card application. It features a robust **Go (Gin-gonic)** backend that guarantees transaction consistency and ACID compliance, paired with a custom **glassmorphic React SPA** frontend powered by **TanStack Query**.

---

## 🚀 Tech Stack

### Backend
* **Language**: Go 1.26+
* **Web Framework**: Gin-Gonic (high performance, routing, middlewares)
* **Database Driver**: PGX v5 (connection pooling, native Postgres integration)
* **Database**: PostgreSQL 16 (relational database with transaction isolation)
* **Migrations**: Golang-migrate (versioned database migrations)

### Frontend
* **Build System**: Vite (lightning-fast HMR and building)
* **Framework**: React 19 (functional components, contexts, custom hooks)
* **Server State Management**: TanStack Query (React Query v5 for caching and reactive UI invalidation)
* **Styling**: Tailwind CSS v4 (responsive utility-first layout, custom glassmorphism design system)
* **Icons**: Lucide React
* **API Client**: Axios (configured with credentials and global interceptors)

### DevOps & CI/CD
* **Containers**: Docker & Docker Compose (multi-container orchestrated setup)
* **CI**: GitHub Actions (automated testing pipeline for backend)

---

## 🔒 Key Design & Features

1. **Virtual Card Ecosystem**:
   * Issue virtual Visa/Mastercards directly linked to specific wallet balances (enforced **1 card limit per wallet currency**).
   * Freeze/Unfreeze cards instantly to restrict unauthorized usage.
   * View card spendings and process test merchant payments on demand.
   * View masked card numbers by default with secure unmasking (`CVV` and 16-digit card number) upon user request.
   * Interactive single-card sliding carousel UI.

2. **Transaction Integrity (ACID)**:
   * Implements **pessimistic row-level locking (`FOR UPDATE`)** in Go transactions when updating wallet balances.
   * Prevents **"Lost Update"** concurrency bugs during simultaneous transfers or card payments.

3. **Idempotency Protection**:
   * The `/transfer` endpoint accepts an `X-Idempotency-Key` header.
   * Prevents duplicate requests (e.g., due to network retries or double clicks) from executing multiple transfers.

4. **Dynamic Exchange Rates & Destination Lookup**:
   * Automatic 500ms debounced recipient verification when typing destination wallet numbers.
   * Live exchange rate calculations (`GET /exchange-rate`) displaying exact recipient amounts.

5. **JWT Auth via HttpOnly Cookies**:
   * Secure authentication with JSON Web Tokens (JWT) stored in HTTP-Only, Secure cookies to prevent XSS token theft.

6. **Executive PDF & CSV Statement Exports**:
   * Download sleek, executive account statements formatted as **PDF** or **CSV**.
   * Features branded header banners, transaction summaries, decimal currency formatting, and card merchant descriptions.

---

## 📂 Project Structure

```
├── .github/workflows/       # GitHub Actions CI pipelines
│   └── ci.yml               # Runs automated tests on every push/PR
├── backend/                 # Backend source code
│   ├── cmd/api/             # App entrypoint (main.go)
│   ├── internal/            # Core business logic
│   │   ├── db/              # Postgres connections and pool configuration
│   │   ├── dto/             # Request/Response Data Transfer Objects
│   │   ├── handlers/        # Gin controllers and router definitions
│   │   ├── middleware/      # Auth and CORS middlewares
│   │   ├── models/          # Relational struct models (Wallet, Card, Transaction)
│   │   ├── repos/           # Database access layer (SQL queries and transactions)
│   │   ├── services/        # Business logic (Cards, Exchange rates, Transfers)
│   │   └── worker/          # Asynchronous webhook queues/workers
│   ├── migrations/          # Up/Down SQL schema migrations
│   ├── pkg/                 # Common helpers (apperrors, JWT, response, PDF/CSV export)
│   ├── Dockerfile           # Multistage backend container build
│   └── docker-compose.yml   # Docker compose configuration for local dev DB
├── frontend/                # Frontend source code
│   ├── public/              # Static assets
│   ├── src/                 # React frontend source files
│   │   ├── components/      # UI components (CardsSection, TransferForm, SpendingsModal)
│   │   ├── context/         # Auth state providers
│   │   ├── hooks/           # TanStack Query custom hooks (useQueries.js)
│   │   ├── lib/             # Axios API config
│   │   └── pages/           # Pages (Dashboard, Login, Register)
│   ├── Dockerfile           # Multistage frontend build served via Nginx
│   └── nginx.conf           # Custom Nginx configuration
└── docker-compose.yml       # Orchestrated system (Frontend + Backend + DB)
```

---

## ⚙️ Development Setup

### Running with Docker (Recommended)
Build and spin up the entire application stack (PostgreSQL + Backend + Frontend) in one command:

1. Clone the repository and navigate to the project root.
2. Spin up the containers:
   ```bash
   docker compose up --build -d
   ```
3. Run the database migrations to set up tables:
   ```bash
   cd backend
   make migrate-up
   ```
4. Access the application:
   * **Frontend**: [http://localhost:3000](http://localhost:3000)
   * **Backend API**: [http://localhost:8080](http://localhost:8080)

---

### Running Locally (Manual Setup)

#### 1. Database Setup
1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Start the database service:
   ```bash
   docker compose up -d db
   ```
3. Run database migrations:
   ```bash
   make migrate-up
   ```

#### 2. Run the Backend API
1. Create a `backend/.env` file from the example:
   ```bash
   cp .env.example .env
   ```
2. Start the Go server:
   ```bash
   go run cmd/api/main.go
   ```
   *Backend serves requests on `http://localhost:8080`.*

#### 3. Run the Frontend
1. Open a new terminal and navigate to the frontend:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Run the Vite development server:
   ```bash
   npm run dev
   ```
   *Frontend is running on [http://localhost:3000](http://localhost:3000).*

---

## 🛠️ Running Tests
To run backend unit and service tests:
```bash
cd backend
go test -v ./...
```

---

## 📡 API Reference

### Auth
* **`POST /api/v1/auth/register`**: Registers a new user.
* **`POST /api/v1/auth/login`**: Authenticates user and sets HttpOnly JWT cookie.
* **`POST /api/v1/logout`** (Protected): Clears user session.

### Wallets
* **`GET /api/v1/wallets`** (Protected): Retrieves all wallets owned by the authenticated user.
* **`POST /api/v1/wallets`** (Protected): Activates/Creates a new wallet for a specified currency (`TRY`, `USD`, or `EUR`).
* **`GET /api/v1/wallets/verify/:walletID`** (Protected): Verifies wallet number existence and owner details.
* **`DELETE /api/v1/wallets/:walletID`** (Protected): Soft deletes the specified wallet.

### Virtual Cards
* **`GET /api/v1/cards`** (Protected): Retrieves all virtual cards owned by the user.
* **`POST /api/v1/cards`** (Protected): Issues a new virtual card linked to a wallet (Limit 1 per wallet currency).
* **`GET /api/v1/cards/:cardID/details`** (Protected): Retrieves unmasked 16-digit card number and CVV.
* **`GET /api/v1/cards/:cardID/transactions`** (Protected): Retrieves spendings history for the specified card.
* **`POST /api/v1/cards/:cardID/freeze`** (Protected): Freezes the card.
* **`POST /api/v1/cards/:cardID/unfreeze`** (Protected): Activates/unfreezes the card.
* **`DELETE /api/v1/cards/:cardID`** (Protected): Terminates/closes the card.
* **`POST /api/v1/cards/:cardID/process-payment`** (Protected): Processes a merchant transaction against the card.

### Transfers & History
* **`POST /api/v1/transfer`** (Protected): Initiates money transfer between wallets with idempotency checking.
* **`GET /api/v1/exchange-rate`** (Protected): Retrieves live conversion exchange rates.
* **`GET /api/v1/transactions/:walletID`** (Protected): Retrieves transaction logs for the wallet.
* **`GET /api/v1/transactions/:walletID/export`** (Protected): Exports transaction statements (`?format=pdf` or `?format=csv`).
