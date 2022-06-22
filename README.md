# WIP: Budget Manager

**Budger Manager** is an easy-to-use, lightweight and self-hosted solution to track your finances

**TODO: add badges**
**TODO: add screenshot**

- [Features](#features)
- [Install](#install)
- [Configuration](#configuration)
- [Development](#development)
	- [Scripts](#scripts)
	- [API](#api)

***

## Features

- **Easy-to-use** - simple and intuitive UI

- **Lightweight** - **TODO**

- **Self-hosted** - you don't need to trust any proprietary software to store your financial information

## Install

You need [Docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/install/) (optional)

1. Create `docker-compose.yml` with the following content (you can find more settings in [Configuration](#configuration) section):

    ```yaml
    version: "2.4"

    services:
      budget-manager:
        image: ghcr.io/shoshinnikita/budget-manager:latest
        container_name: budget-manager
        environment:
          STORE_BOLT_PATH: ./var/budget-manager.db
          SERVER_AUTH_BASIC_CREDS: <your credentials> # more info in 'Configuration' section
        ports:
          - "8080:8080"
        volumes:
          - ./var:/srv/var
    ```

2. Run `docker-compose up -d`
3. Go to `http://localhost:8080`
4. Profit!

## Configuration

| Env Var                   | Default value | Description                                                                                                      |
| ------------------------- | ------------- | ---------------------------------------------------------------------------------------------------------------- |
| `STORE_BOLT_PATH`         | `localhost`   | Path to [bolt](https://github.com/etcd-io/bbolt) file                                                            |
| `SERVER_PORT`             | `8080`        |                                                                                                                  |
| `SERVER_USE_EMBED`        | `true`        | **TODO**                                                                                                         |
| `SERVER_AUTH_DISABLE`     | `false`       | Disable authentication                                                                                           |
| `SERVER_AUTH_BASIC_CREDS` |               | List of comma separated `login:password` pairs. Passwords must be hashed using BCrypt (`htpasswd -nB <user>`)    |
| `SERVER_ENABLE_PROFILING` | `false`       | Enable [pprof](https://blog.golang.org/pprof) handlers. You can find handler urls [here](internal/web/routes.go) |

## Development

### Scripts

```bash
# Run with dev config
make

# Run tests
make test

# Run linters
make lint

# Run both tests and linters
make check
```

### API

**TODO**

