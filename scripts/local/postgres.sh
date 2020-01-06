#!/bin/bash

docker stop budget-manager_postgres

MOUNT=""
MODE=${1:-full}
case $MODE in
	"full")
		MOUNT_DIR="$(pwd)/var/pg_data"

		echo "Mode is 'full'"
		echo "Mount '${MOUNT_DIR}'"
		MOUNT="-v ${MOUNT_DIR}:/var/lib/postgresql/data"
		;;

	"test" | "testing")
		echo "Mode is 'test'"
		echo "Mount nothing for test"
		# Mount nothing for tests
		MOUNT=""
		;;

	*)
		echo "Invalid mode '${MODE}'"
		exit 1
		;;
esac

docker run --rm -d \
	--name budget-manager_postgres \
	-p "5432:5432" \
	${MOUNT} \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_DB=postgres \
	postgres -c "log_statement=all"
