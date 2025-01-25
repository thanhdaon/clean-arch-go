include .env
export

.PHONY: gen_openapi
gen_openapi:
	oapi-codegen -generate types -o "ports/openapi_types.gen.go" -package "ports" "ports/openapi.yml"
	oapi-codegen -generate chi-server -o "ports/openapi_api.gen.go" -package "ports" "ports/openapi.yml"

start:
	go run cmd/main.go

