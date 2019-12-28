DOCKER_COMPOSE=./scripts/docker/docker-compose.yml

# Run

run: run-docker

run-docker: clear
	docker-compose -f ${DOCKER_COMPOSE} up \
		--build \
		--force-recreate \
		--renew-anon-volumes \
		--exit-code-from budget_manager

run-local: clear
	# Run Postgres
	./scripts/local/postgres.sh
	# Run Budget Manager
	./scripts/local/run.sh

# Clear

clear: clear-local clear-docker

clear-docker:
	# Stop and remove containers and volumes
	docker-compose -f ${DOCKER_COMPOSE} down -v

clear-local:
	docker stop budget_manager_postgres || true

# Tests

test: test-integ

test-unit:
	go test -mod vendor -count 1 -v ./...

test-integ:
	# Run Postgres
	./scripts/local/postgres.sh test

	# Run integration tests. We disable parallel tests for packages (with '-p 1') to avoid DB errors (same situation: https://medium.com/@xcoulon/how-to-avoid-parallel-execution-of-tests-in-golang-763d32d88eec)
	go test -mod=vendor -count=1 -p=1 --tags=integration -v \
		-cover -coverprofile=cover.out -coverpkg=github.com/ShoshinNikita/budget_manager/... \
		./...
	go tool cover -func=cover.out
	rm cover.out

	# Stop and remove DB
	docker stop budget_manager_postgres

# Other

build:
	go build -mod vendor -o bin/budget_manager cmd/budget_manager/main.go

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
