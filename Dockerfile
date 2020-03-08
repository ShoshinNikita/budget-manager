# Global workdir variable
ARG WORKDIR=/go/src/github.com/ShoshinNikita/budget-manager


#
# Build a binary file
#

FROM golang:1.14-alpine3.11 as backend-builder
ARG WORKDIR

ENV CGO_ENABLED=0

WORKDIR ${WORKDIR}

# Copy dependencies (for better caching)
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy code
COPY cmd ./cmd
COPY internal ./internal

# Build
ARG LDFLAGS
RUN go build -ldflags "${LDFLAGS}" -mod vendor -o ./bin/budget-manager ./cmd/budget-manager/main.go


#
# Minify files
#

FROM ubuntu:18.04 as frontend-builder
ARG WORKDIR

WORKDIR ${WORKDIR}

# Install minify
ADD https://github.com/tdewolff/minify/releases/download/v2.7.2/minify_2.7.2_linux_amd64.tar.gz minify.tar.gz
RUN tar -xvzf minify.tar.gz -C /usr/local/bin && rm minify.tar.gz

# Minify files
COPY templates ./templates
COPY static ./static

RUN minify --html-keep-default-attrvals -o templates/ templates && \
	minify -o static/css/ static/css


#
# Build the final image
#

FROM alpine:3.11
ARG WORKDIR

WORKDIR /srv

# Copy 'static' directory
COPY --from=frontend-builder ${WORKDIR}/static ./static
# Copy 'templates' directory
COPY --from=frontend-builder ${WORKDIR}/templates ./templates
# Copy binaries
COPY --from=backend-builder ${WORKDIR}/bin .

ENTRYPOINT [ "/srv/budget-manager" ]
