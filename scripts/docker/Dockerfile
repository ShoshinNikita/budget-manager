FROM golang:1.12.9-alpine as builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/ShoshinNikita/budget_manager

# Copy dependencies (for better caching)
COPY vendor ./vendor
COPY go.mod .
COPY go.sum .

# Copy code
COPY cmd ./cmd
COPY internal ./internal

RUN go build -mod vendor -o ./bin/budget_manager ./cmd/budget_manager/main.go


FROM alpine:3.9

WORKDIR /srv

# Copy binaries
COPY --from=builder /go/src/github.com/ShoshinNikita/budget_manager/bin .
# Copy 'templates' directory
COPY templates ./templates

EXPOSE 8080

ENTRYPOINT [ "/srv/budget_manager" ]
