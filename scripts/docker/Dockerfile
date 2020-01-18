# Global workdir variable
ARG WORKDIR=/go/src/github.com/ShoshinNikita/budget-manager


#
# Build a binary file
#

FROM golang:1.13.5-alpine3.10 as backend-builder
ARG WORKDIR

ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR ${WORKDIR}

# Copy dependencies (for better caching)
COPY vendor ./vendor
COPY go.mod .
COPY go.sum .

# Copy code
COPY cmd ./cmd
COPY internal ./internal

# Build
RUN go build -mod vendor -o ./bin/budget-manager ./cmd/budget-manager/main.go


#
# Minify files
#

FROM ubuntu:18.04 as frontend-builder
ARG WORKDIR

WORKDIR ${WORKDIR}

# Install minify
RUN apt-get update && \
	apt-get install -y wget

RUN wget -O minify.tar.gz https://github.com/tdewolff/minify/releases/download/v2.7.2/minify_2.7.2_linux_amd64.tar.gz && \
	tar -xvzf minify.tar.gz -C /usr/local/bin

# Minify files
COPY templates ./templates
COPY static ./static

RUN minify --html-keep-default-attrvals -o templates/ templates && \
	minify -o static/css/ static/css


#
# Build the final image
#

FROM alpine:3.9
ARG WORKDIR

WORKDIR /srv

# Copy 'static' directory
COPY --from=frontend-builder ${WORKDIR}/static ./static
# Copy 'templates' directory
COPY --from=frontend-builder ${WORKDIR}/templates ./templates
# Copy binaries
COPY --from=backend-builder ${WORKDIR}/bin .

EXPOSE 8080

ENTRYPOINT [ "/srv/budget-manager" ]
