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

.PHONY: atlas-diff
atlas-diff:
	atlas schema diff --env dev

.PHONY: db-setup
db-setup: atlas-apply sqlc-generate

.PHONY: dev-deps
dev-deps:
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

actionlint:
	actionlint
	ghalint run
	pinact run
