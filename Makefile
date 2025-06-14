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

actionlint:
	actionlint
	ghalint run
	pinact run
