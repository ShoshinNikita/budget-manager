include .env
export

SHELL := /bin/bash

# Make all targets phony. Get list of all targets: 'cat Makefile | grep -P -o "^[\w-]+:" | rev | cut -c 2- | rev | sort | uniq'
.PHONY: build check default docker docker-build docker-clear docker-run export-ldflags generate-docs lint run test test-integ test-unit

default: build run

# build builds a binary file
build: export-ldflags
	@ echo "Build Budget Manager..."
	@ go build -ldflags "${LDFLAGS}" -mod=vendor -o bin/budget-manager cmd/budget-manager/main.go

# run runs built Budget Manager
run:
	@ echo "Run Budget Manager..."
	@ ./bin/budget-manager

#
# Tests
#

test: test-integ

TEST_CMD=go test -v -mod=vendor ${TEST_FLAGS} \
	-cover -coverprofile=cover.out -coverpkg=github.com/ShoshinNikita/budget-manager/...\
	./cmd/... ./internal/... && \
	go tool cover -func=cover.out && rm cover.out

# test-unit runs unit tests
test-unit: TEST_FLAGS=-short
test-unit:
	@ echo "Run unit tests..."
	${TEST_CMD}

# test-integ runs both unit and integration tests
#
# Disable parallel tests for packages (with '-p 1') to avoid DB errors.
# Same situation: https://medium.com/@xcoulon/how-to-avoid-parallel-execution-of-tests-in-golang-763d32d88eec)
#
test-integ: TEST_FLAGS=-p=1
test-integ:
	@ echo "Run integration tests..."
	${TEST_CMD}

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

#
# Other
#

# lint runs golangci-lint - https://github.com/golangci/golangci-lint
#
# Use go cache to speed up execution: https://github.com/golangci/golangci-lint/issues/1004
#
lint:
	# TODO: enable later
ifdef ENABLE_LINTERS
	@ echo "Run golangci-lint..."
	@ docker run --rm -it --network=none \
		-v $(shell go env GOCACHE):/cache/go \
		-e GOCACHE=/cache/go \
		-e GOLANGCI_LINT_CACHE=/cache/go \
		-v $(shell go env GOPATH)/pkg:/go/pkg \
		-v $(shell pwd):/app \
		-w /app \
		golangci/golangci-lint:v1.42-alpine golangci-lint run --config .golangci.yml -v
endif

check: build lint test

# generate-docs generates Swagger API documentation with swag - https://github.com/swaggo/swag
generate-docs:
	@ echo "Clear Swagger API docs..."
	@ swag init --generalInfo cmd/budget-manager/main.go --output docs
	@ echo "Generate Swagger API docs..."
	@ rm ./docs/swagger.json ./docs/docs.go
