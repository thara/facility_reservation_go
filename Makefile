.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: tsp
tsp:
	cd ./spec ; tsp compile .

.PHONY: ogen
ogen: tsp
	ogen -target internal/api -package api --clean ./spec/tsp-output/schema/3.1.0/openapi.yaml

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

.PHONY: migrate-up
migrate-up:
	migrate -database "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable" -path _migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable" -path _migrations down


.PHONY: migrate-version
migrate-version:
	migrate -database "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable" -path _migrations version

.PHONY: schema-generate
schema-generate:
	./_scripts/generate-schema.sh

.PHONY: schema-check
schema-check:
	./_scripts/check-schema.sh

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


.PHONY: db-setup
db-setup: db-up migrate-up schema-generate sqlc-generate

.PHONY: test
test:
	go test ./... -v

.PHONY: test-short
test-short:
	go test ./... -v -short

.PHONY: test-integration
test-integration:
	@if [ -z "$$CI" ]; then \
		$(MAKE) db-up migrate-up; \
	fi
	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable" go test ./... -v

.PHONY: test-all
test_all: test-short test-integration

.PHONY: build_dev
build_dev: clean fmt lint sqlc-generate ogen test_all
	go build -o ./bin/api-server ./cmd/api-server/

.PHONY: dev-deps
dev-deps:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: actionlint
actionlint:
	actionlint
	ghalint run
	pinact run
