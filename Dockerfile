#
# Build a binary file
#

FROM golang:1.18-alpine as backend-builder


WORKDIR /build/backend

# Copy dependencies (for better caching)
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy code
COPY main.go main.go
COPY cmd ./cmd
COPY internal ./internal

# Build
ARG LDFLAGS
RUN go build -ldflags "${LDFLAGS}" -o ./bin/budget-manager ./main.go


#
# Build the final image
#

FROM alpine:3.16

LABEL \
	org.opencontainers.image.url=https://github.com/users/ShoshinNikita/packages/container/package/budget-manager \
	org.opencontainers.image.source=https://github.com/ShoshinNikita/budget-manager

WORKDIR /srv

COPY --from=backend-builder build/backend/bin .

ENTRYPOINT [ "/srv/budget-manager" ]
