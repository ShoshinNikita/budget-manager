.PHONY: build dev interactive

RUN=docker run -it \
		-v $(shell pwd):/home/app \
		--workdir /home/app \
		-p 8081:8081 \
		--user $(shell id -u):$(shell id -g) \
		$(shell docker build -q .) ${COMMAND}

interactive: COMMAND=bash
interactive:
	@ ${RUN}

dev: COMMAND=yarn dev
dev:
	@ ${RUN}

build: COMMAND=yarn build
build:
	@ ${RUN}
