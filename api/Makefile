include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_CONN_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_CONN_ADDR) down $(filter-out $@,$(MAKECMDGOALS))
