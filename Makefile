.PHONY: api web dev test docker-up docker-down docker-logs migrate-up migrate-down test-gateway test-checkout

# Run backend API server (port 8081)
api:
	cd backend && go run cmd/api/main.go

# Run frontend Vite dev server (port 3005)
web:
	cd frontend && npm run dev

# Run unit tests across all backend packages
test:
	cd backend && go test -v ./internal/services/... ./internal/repos/... ./pkg/...

# Start isolated Docker infrastructure (Postgres 55432, Redis 56379, RabbitMQ 5673)
docker-up:
	docker compose up -d

# Stop Docker infrastructure
docker-down:
	docker compose down

# Follow Docker logs
docker-logs:
	docker compose logs -f

# Database migrations
migrate-up:
	cd backend && $(MAKE) migrate-up

migrate-down:
	cd backend && $(MAKE) migrate-down

# End-to-end integration test suites against Foodeli contracts
test-gateway:
	cd backend && ./scripts/test_gateway.sh

test-checkout:
	cd backend && ./scripts/test_hosted_checkout.sh
