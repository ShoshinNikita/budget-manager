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


FROM alpine:3.9
ARG WORKDIR

WORKDIR /srv

# Copy 'static' directory
COPY static ./static
# Copy 'templates' directory
COPY templates ./templates
# Copy binaries
COPY --from=backend-builder ${WORKDIR}/bin .

EXPOSE 8080

ENTRYPOINT [ "/srv/budget-manager" ]
