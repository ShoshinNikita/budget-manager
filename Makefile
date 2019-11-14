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
	go test -v ./...

test-integ:
	# Run Postgres
	./scripts/local/postgres.sh test
	echo "Wait fot DB..."
	sleep 5
	# Run integration tests
	go test --tags=integration -v ./...
	# Stop and remove DB
	docker stop budget_manager_postgres
