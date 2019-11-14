DOCKER_COMPOSE=./scripts/docker/docker-compose.yml

# Run

run: run-docker

run-docker: clear
	docker-compose -f ${DOCKER_COMPOSE} up --build

run-local: clear-local
	# Run Postgres
	./scripts/local/postgres.sh
	echo "Wait fot DB..."
	sleep 5
	# Run Budget Manager
	./scripts/local/run.sh

# Clear

clear: clear-docker

clear-docker:
	# Stop and remove containers and volumes
	docker-compose -f ${DOCKER_COMPOSE} rm -v --stop --force

clear-local:
	docker stop budget_manager_postgres || true

# Tests

test: test-integ

test-unit:
	go test -mod vendor -v ./...

test-integ:
	# Run Postgres
	./scripts/local/postgres.sh test
	echo "Wait fot DB..."
	sleep 5
	# Run integration tests
	go test -mod vendor --tags=integration -v ./...
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
	golangci-lint run \
		--exclude="Error return value of \`tx.Rollback\` is not checked" \
		--max-issues-per-linter=0 \
		--max-same-issues=0
