DOCKER_COMPOSE=./scripts/docker/docker-compose.yml

run: run-local

# Build

build:
	go build -mod vendor -o bin/budget-manager cmd/budget-manager/main.go

build-docker:
	docker-compose -f ${DOCKER_COMPOSE} build

# Run

run-local: stop
	# Run Postgres
	./scripts/local/postgres.sh
	# Run Budget Manager
	./scripts/local/run.sh

run-docker: stop
	docker-compose -f ${DOCKER_COMPOSE} up \
		--build \
		--force-recreate \
		--renew-anon-volumes \
		--exit-code-from budget-manager

# Clear

stop: stop-local stop-docker

stop-local:
	docker stop budget-manager_postgres || true

stop-docker:
	# Stop and remove containers and volumes
	docker-compose -f ${DOCKER_COMPOSE} down -v || true

# Tests

test: test-integ

test-unit:
	go test -mod vendor -count 1 -v ./...

test-integ: stop
	# Run Postgres
	./scripts/local/postgres.sh test

	# Run integration tests. We disable parallel tests for packages (with '-p 1') to avoid DB errors (same situation: https://medium.com/@xcoulon/how-to-avoid-parallel-execution-of-tests-in-golang-763d32d88eec)
	go test -mod=vendor -count=1 -p=1 --tags=integration -v \
		-cover -coverprofile=cover.out -coverpkg=github.com/ShoshinNikita/budget-manager/... \
		./...
	go tool cover -func=cover.out
	rm cover.out

	# Stop and remove DB
	docker stop budget-manager_postgres

# Other

lint:
	# golangci-lint - https://github.com/golangci/golangci-lint
	#
	# golangci-lint can be installed with:
	#   curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.20.0
	#
	# More installation options: https://github.com/golangci/golangci-lint#binary-release
	#
	golangci-lint run --config .golangci.yml

check: lint test

bench:
	./tools/bench.sh
