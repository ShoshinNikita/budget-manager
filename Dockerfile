#
# Minify templates and static files
#

FROM ubuntu:18.04 as frontend-builder

WORKDIR /build/frontend

# Install minify
ADD https://github.com/tdewolff/minify/releases/download/v2.7.2/minify_2.7.2_linux_amd64.tar.gz minify.tar.gz
RUN tar -xvzf minify.tar.gz -C /usr/local/bin && rm minify.tar.gz

# Minify files
COPY templates ./templates
COPY static ./static

RUN minify --html-keep-default-attrvals -o templates/ templates && \
	minify -o static/css/ static/css && \
	minify -o static/js/ static/js

# Minify has poor support of Go template syntax. For example, it converts all attributes to lower case.
# It causes template execution errors. For example, <html lang="{{ .Lang }}"> is converted to <html lang="{{ .lang }}">
#
# There's an issue (https://github.com/tdewolff/minify/issues/35) for this problem opened in 2015, but it is still opened.
# So, we have to find a workaround. For example, we can manually fix Go templates in attributes (they can be found with
# this regexp: \{\{ .*? \}\}>
#
RUN sed -i "s/tohtmlattr/toHTMLAttr/g" templates/month.html templates/months.html


#
# Build a binary file
#

FROM golang:1.16-alpine as backend-builder

ENV CGO_ENABLED=0

WORKDIR /build/backend

# Copy dependencies (for better caching)
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy code
COPY cmd ./cmd
COPY internal ./internal

# Copy minified templates and static files
COPY --from=frontend-builder build/frontend/static ./static
COPY --from=frontend-builder build/frontend/templates ./templates

# Build
ARG LDFLAGS
RUN go build -ldflags "${LDFLAGS}" -o ./bin/budget-manager ./cmd/budget-manager/main.go


#
# Build the final image
#

FROM alpine:3.11

LABEL \
	org.opencontainers.image.url=https://github.com/users/ShoshinNikita/packages/container/package/budget-manager \
	org.opencontainers.image.source=https://github.com/ShoshinNikita/budget-manager

WORKDIR /srv

COPY --from=backend-builder build/backend/bin .

ENTRYPOINT [ "/srv/budget-manager" ]
