include .env
export

.PHONY: openapi test test-one component-test component-test-one integration-test integration-test-one unit-test unit-test-one

openapi:
	oapi-codegen -generate types -o "ports/openapi_types.gen.go" -package "ports" "ports/openapi.yml"
	oapi-codegen -generate chi-server -o "ports/openapi_api.gen.go" -package "ports" "ports/openapi.yml"

test:
	go test ./...

test-one:
	go test -count=1 -run $(TEST) ./...

component-test:
	go test ./tests/...

component-test-one:
	go test -count=1 -run $(TEST) ./tests/...

integration-test:
	go test ./adapters/...

integration-test-one:
	go test -count=1 -run $(TEST) ./adapters/...

unit-test:
	go test ./domain/...

unit-test-one:
	go test -count=1 -run $(TEST) ./domain/...

start:
	go run cmd/start/main.go

demo:
	go run cmd/demo/main.go

migration-status:
	sql-migrate status -config=migrations/dbconfig.yml -env="development"

migration-new:
	sql-migrate new -config=migrations/dbconfig.yml -env="development"

migration-up:
	sql-migrate up -config=migrations/dbconfig.yml -env="development"