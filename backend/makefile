POSTGRESQL_URL := postgresql://admin:secretpassword@localhost:55432/fintech?sslmode=disable

.PHONY: run test migrate-up migrate-down test-gateway test-checkout docker-up docker-down

run:
	go run cmd/api/main.go

test:
	go test -v ./internal/services/... ./internal/repos/... ./pkg/...

migrate-up:
	migrate -database "${POSTGRESQL_URL}" -path ./migrations up

migrate-down:
	migrate -database "${POSTGRESQL_URL}" -path ./migrations down

test-gateway:
	./scripts/test_gateway.sh

test-checkout:
	./scripts/test_hosted_checkout.sh

docker-up:
	docker compose -f ../docker-compose.yml up -d

docker-down:
	docker compose -f ../docker-compose.yml down