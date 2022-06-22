.PHONY: build dev interactive

RUN=docker run --rm -it \
		-v $(shell pwd):/home/app \
		--workdir /home/app \
		-p 8081:8081 \
		--user $(shell id -u):$(shell id -g) \
		$(shell docker build -q .) ${COMMAND}

interactive: COMMAND=bash
interactive:
	@ ${RUN}

dev: COMMAND=yarn run dev
dev:
	@ ${RUN}

build: COMMAND=yarn run build
build:
	@ ${RUN}

check: COMMAND=yarn run check
check:
	@ ${RUN}

format: COMMAND=yarn run format
format:
	@ ${RUN}
