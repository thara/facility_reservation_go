.PHONY: tsp
tsp:
	cd ./spec ; tsp compile .

.PHONY: ogen
ogen: tsp
	ogen -target oas -package oas --clean ./spec/tsp-output/schema/3.1.0/openapi.yaml

.PHONY: format
fmt:
	golangci-lint fmt -v

.PHONY: lint
lint:
	golangci-lint config verify
	golangci-lint run -v

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

.PHONY: atlas-status
atlas-status:
	atlas schema inspect --env dev

.PHONY: atlas-apply
atlas-apply:
	atlas schema apply --env dev

.PHONY: atlas-apply-test
atlas-apply-test:
	atlas schema apply --env test --auto-approve || true

.PHONY: atlas-diff
atlas-diff:
	atlas schema diff --env dev

.PHONY: db-up
db-up:
	docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@until docker compose exec postgres pg_isready -U postgres > /dev/null 2>&1; do sleep 1; done
	@echo "Database is ready!"

.PHONY: db-down
db-down:
	docker compose down

.PHONY: db-logs
db-logs:
	docker compose logs -f postgres

.PHONY: db-clean
db-clean:
	docker compose down -v

.PHONY: db-test-up
db-test-up:
	docker compose up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@until docker compose exec postgres-test pg_isready -U postgres > /dev/null 2>&1; do sleep 1; done
	@echo "Test database is ready!"

.PHONY: db-setup
db-setup: db-up atlas-apply sqlc-generate

.PHONY: test
test:
	go test ./... -v

.PHONY: test-short
test-short:
	go test ./... -v -short

.PHONY: test-integration
test-integration: db-test-up atlas-apply-test
	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable" go test ./... -v

.PHONY: dev-deps
dev-deps:
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

actionlint:
	actionlint
	ghalint run
	pinact run
