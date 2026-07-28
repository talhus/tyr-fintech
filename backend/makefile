POSTGRESQL_URL := postgresql://admin:secretpassword@localhost:55432/fintech?sslmode=disable

migrate-up:
	migrate -database ${POSTGRESQL_URL} -path ./migrations up

migrate-down:
	migrate -database ${POSTGRESQL_URL} -path ./migrations down