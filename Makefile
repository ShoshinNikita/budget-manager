SHELL := /bin/bash

# Make all targets phony. Get list of all targets: 'cat Makefile | grep -P -o "^[\w-]+:" | rev | cut -c 2- | rev | sort | uniq'
.PHONY: build check default docker docker-build docker-clear docker-run export-config export-ldflags generate-docs lint run run-pg run-pg-test stop-pg stop-pg-test test test-integ test-unit

default: build run

# build builds a binary file
build: export-ldflags
	@ echo "Build Budget Manager..."
	@ go build -ldflags "${LDFLAGS}" -mod vendor -o bin/budget-manager cmd/budget-manager/main.go

# run runs built Budget Manager
run: export-config
	@ echo "Run Budget Manager..."
	@ ./bin/budget-manager

#
# Docker
#

docker: docker-build docker-run

# docker-build builds a Docker image
docker-build: TAG?=budget-manager:latest
docker-build: export-ldflags
	@ echo "Build Docker image for Budget Manager..."
	@ docker build -t ${TAG} --build-arg LDFLAGS="${LDFLAGS}" .

# docker-run runs both Budget Manager and PostgreSQL in containers
docker-run:
	@ echo "Run Budget Manager in Docker container..."
	@ docker-compose up --exit-code-from budget-manager

# docker-clear downs containers and removes volumes
docker-clear:
	@ docker-compose down -v || true

#
# Tests
#

test: test-integ

TEST_CMD=go test -v -mod=vendor ${TEST_FLAGS} \
	-cover -coverprofile=cover.out -coverpkg=github.com/ShoshinNikita/budget-manager/...\
	./cmd/... ./internal/... && \
	go tool cover -func=cover.out && rm cover.out

# test-unit runs only unit tests
test-unit:
	@ echo "Run unit tests..."
	${TEST_CMD}

# test-integ runs PostgreSQL in test mode and runs all tests
#
# Disable parallel tests for packages (with '-p 1') to avoid DB errors.
# Same situation: https://medium.com/@xcoulon/how-to-avoid-parallel-execution-of-tests-in-golang-763d32d88eec)
#
test-integ: TEST_FLAGS = -tags=integration -p=1 -count=1
test-integ:
	@ echo "Run integration tests..."
	@ $(MAKE) --no-print-directory run-pg-test
	${TEST_CMD}
	@ $(MAKE) --no-print-directory stop-pg-test

#
# PostgreSQL
#

PG_ENV=-e POSTGRES_USER=postgres -e POSTGRES_DB=postgres -e POSTGRES_HOST_AUTH_METHOD=trust
PG_CONAINER_NAME=budget-manager_pg

# run-pg runs develop PostgreSQL instance with mounted '_var/pg_data' directory
run-pg: stop-pg
	@ echo "Run develop PostgreSQL instance..."
	@ docker run --rm -d \
		--name ${PG_CONAINER_NAME} \
		-p "5432:5432" \
		-v $(shell pwd)/_var/pg_data:/var/lib/postgresql/data \
		${PG_ENV} \
		postgres:12-alpine -c "log_statement=all"

# stop-pg stops develop PostgreSQL instance
stop-pg:
	@ echo "Stop develop PostgreSQL instance..."
	@ docker stop ${PG_CONAINER_NAME} > /dev/null 2>&1 || true

# run-pg-test runs test PostgreSQL instance
run-pg-test: stop-pg stop-pg-test
	@ echo "Run test PostgreSQL instance..."
	@ docker run --rm -d \
		--name ${PG_CONAINER_NAME}-test \
		-p "5432:5432" \
		${PG_ENV} \
		postgres:12-alpine -c "log_statement=all"

# stop-pg-test stops test PostgreSQL instance
stop-pg-test:
	@ echo "Stop test PostgreSQL instance..."
	@ docker stop ${PG_CONAINER_NAME}-test > /dev/null 2>&1 || true

#
# Configuration
#

# export-ldflags exports LDFLAGS env variable. It is used during the build process to set version
# and git hash. It can be used as a dependency target
#
# For example, we have target 'build':
#
#  build: export-ldflags
#    go build -ldflags "${LDFLAGS}" main.go
#
# We can use it as 'make build VERSION=v1.0.0'. Then, next command will be executed:
#
#  go build -ldflags "-s -w -X 'main.version=v1.0.0' -X 'main.gitHash=some_hash'" main.go
#
export-ldflags: GIT_HASH=$(shell git log -1 --pretty="format:%h")
export-ldflags: VERSION?=unknown
export-ldflags:
	$(eval export LDFLAGS=-s -w -X 'main.version=${VERSION}' -X 'main.gitHash=${GIT_HASH}')
	@ echo Use this ldflags: ${LDFLAGS}

# export-config exports configuration env variables
export-config:
	$(eval ${config})

define config
	export DEBUG = true
	export LOGGER_MODE = develop
	export DB_TYPE = postgres
	export DB_PG_HOST = localhost
	export DB_PG_PORT = 5432
	export DB_PG_USER = postgres
	export DB_PG_DATABASE = postgres
	export SERVER_PORT = 8080
	export SERVER_SKIP_AUTH = true
	export SERVER_USE_EMBED = false
	# user:qwerty
	export SERVER_CREDENTIALS = user:$$$$apr1$$$$cpHMFyv.$$$$BSB0aaF3bOrTC2f3V2VYG/
	export SERVER_ENABLE_PROFILING = true
endef

#
# Other
#

# lint runs golangci-lint - https://github.com/golangci/golangci-lint
#
# Use go cache to speed up execution: https://github.com/golangci/golangci-lint/issues/1004
#
lint:
	@ echo "Run golangci-lint..."
	@ docker run --rm -it --network=none \
		-v $(shell go env GOCACHE):/cache/go \
		-e GOCACHE=/cache/go \
		-e GOLANGCI_LINT_CACHE=/cache/go \
		-v $(shell go env GOPATH)/pkg:/go/pkg \
		-v $(shell pwd):/app \
		-w /app \
		golangci/golangci-lint:v1.37-alpine golangci-lint run --config .golangci.yml

check: build lint test

# generate-docs generates Swagger API documentation with swag - https://github.com/swaggo/swag
generate-docs:
	@ echo "Clear Swagger API docs..."
	@ swag init --generalInfo cmd/budget-manager/main.go --output docs
	@ echo "Generate Swagger API docs..."
	@ rm ./docs/swagger.json ./docs/docs.go
