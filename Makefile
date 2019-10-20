DOCKER_COMPOSE=./scripts/docker/docker-compose.yml

run: run_docker
run_docker: clear
	docker-compose -f ${DOCKER_COMPOSE} up --build
run_local: clear
	# Run Postgres
	./scripts/local/postgres.sh
	./scripts/local/run.sh
clear:
	# Stop local

	# Stop compose
	docker-compose -f ${DOCKER_COMPOSE} stop
	docker-compose -f ${DOCKER_COMPOSE} rm