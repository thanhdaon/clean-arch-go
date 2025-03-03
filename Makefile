include .env
export

.PHONY: gen_openapi
openapi:
	oapi-codegen -generate types -o "ports/openapi_types.gen.go" -package "ports" "ports/openapi.yml"
	oapi-codegen -generate chi-server -o "ports/openapi_api.gen.go" -package "ports" "ports/openapi.yml"

test:
	go test ./...

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

