# Makefile for Expiry Tracker Backend

TEST_TARGET_DIR = $$(go list ./... | grep -v "/gen/" | grep -v "/mock/" )
COVER_PROFILE_FILE ?= cover.out
DB_SOURCE ?= postgresql://root:secret@localhost:5432/expiry_tracker_db?sslmode=disable
STEP ?= 0

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

## to install external tools
install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 && \
	go install github.com/cosmtrek/air@v1.49.0 && \
	go install github.com/golang/mock/mockgen@v1.6.0 && \
	go install golang.org/x/vuln/cmd/govulncheck@latest

## to install tools from tools.go
install-tools:
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

## to run linter
lint:
	golangci-lint run --config=.golangci.yml --timeout 10m0s

## to run vulnerability check
vuln:
	govulncheck ./...
	
## to build the app into go binary
build_app:
	go build -race -o expiry-tracker-be

## to build docker image with local ssh key
docker.build:
	./build/local/docker_build.sh

## to run the http server
run:
	go run main.go http_server

# to run the http server with live reload
run.dev:
	air server
## to generate defined task
generate:
	go generate ./...

generate.openapi:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ./api/oapi_codegen.yaml ./api/openapi3.yml

## to generate sql migration file, require NAME to run. e.g make migration NAME=test
migration:
	$(call check_defined, NAME)
	go run github.com/pressly/goose/v3/cmd/goose -dir="./db/migrations" create $(NAME) sql

## to generate go-based sql migration file, require NAME to run. e.g make migration.go NAME=test
migration.go:
	$(call check_defined, NAME)
	go run github.com/pressly/goose/v3/cmd/goose -dir="./db/migrations" create $(NAME) go

## to migrate all the sql migration files
migrate:
	go run main.go migrate

## to rollback the latest version of sql migration
migrate.rollback:
	go run main.go migrate --rollback

help:
	@printf "Available targets:\n\n"
	@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
		helpCommand = $$1; \
		helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
	gsub("\\\\", "", helpCommand); \
	gsub(":+$$", "", helpCommand); \
		printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
	@printf "\n"
