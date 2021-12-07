# Budget Manager [![Last version](https://img.shields.io/github/v/tag/ShoshinNikita/budget-manager?label=version&style=flat-square)](https://github.com/ShoshinNikita/budget-manager/releases/latest) [![GitHub Workflow Status](https://img.shields.io/github/workflow/status/ShoshinNikita/budget-manager/check%20code?label=CI&logo=github&style=flat-square)](https://github.com/ShoshinNikita/budget-manager/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/ShoshinNikita/budget-manager?style=flat-square)](https://goreportcard.com/report/github.com/ShoshinNikita/budget-manager)

**Budger Manager** is an easy-to-use, lightweight and self-hosted solution to track your finances

![Month Page](./docs/images/month_page_large.png)

It was inspired by [Poor-Man's Budgeting Spreadsheet](https://www.reddit.com/r/personalfinance/comments/2tymvf/poormans_budgeting_spreadsheet/) and [You have less money than you think (rus)](https://journal.tinkoff.ru/spreadsheet/). These projects have a fatal flaw: you can't add multiple spends in a single day. **Budger Manager** resolves that issue

**Features:**

- **Easy-to-use** - simple and intuitive UI

- **Lightweight** - backend is written in [Go](https://golang.org/), HTML is prepared with [Go templates](https://golang.org/pkg/text/template/). Vanilla JavaScript is used just to make frontend interactive. So, JS code is very primitive and lightweight: it won't devour all your CPU and RAM (even with Chrome 😉)

- **Self-hosted** - you don't need to trust any proprietary software to store your financial information

You can find more screenshots [here](./docs/images/README.md)

***

- [Install](#install)
- [Configuration](#configuration)
- [Development](#development)
  - [Commands](#commands)
  - [Tools](#tools)
  - [Endpoints](#endpoints)

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
          DB_TYPE: postgres
          DB_PG_HOST: postgres
          DB_PG_PORT: 5432
          DB_PG_USER: postgres
          DB_PG_PASSWORD: <db password>
          DB_PG_DATABASE: postgres
          SERVER_AUTH_BASIC_CREDS: <your credentials> # more info in 'Configuration' section
        ports:
          - "8080:8080"

      postgres:
        image: postgres:12-alpine
        container_name: budget-manager_postgres
        environment:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: <db password>
          POSTGRES_DB: postgres
        volumes:
          # Store data in ./var/pg_data directory
          - type: bind
            source: ./var/pg_data
            target: /var/lib/postgresql/data
    ```

2. Run `docker-compose up -d`
3. Go to `http://localhost:8080` (change the port if needed)
4. Profit!

## Configuration

| Env Var                   | Default value             | Description                                                                                                      |
| ------------------------- | ------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `LOGGER_MODE`             | `prod`                    | Logging format. `dev` or `prod`                                                                                  |
| `LOGGER_LEVEL`            | `info`                    | Logging level. `debug`, `info`, `warn`, `error`, or `fatal`                                                      |
| `DB_TYPE`                 | `postgres`                | Database type. `postgres` or `sqlite`                                                                            |
| `DB_PG_HOST`              | `localhost`               | PostgreSQL host                                                                                                  |
| `DB_PG_PORT`              | `5432`                    | PostgreSQL port                                                                                                  |
| `DB_PG_USER`              | `postgres`                | PostgreSQL username                                                                                              |
| `DB_PG_PASSWORD`          |                           | PostgreSQL password                                                                                              |
| `DB_PG_DATABASE`          | `postgres`                | PostgreSQL database                                                                                              |
| `DB_SQLITE_PATH`          | `./var/budget-manager.db` | Path to the SQLite database                                                                                      |
| `SERVER_PORT`             | `8080`                    |                                                                                                                  |
| `SERVER_USE_EMBED`        | `true`                    | Use the [embedded](https://pkg.go.dev/embed) templates and static files or read them from disk                   |
| `SERVER_AUTH_DISABLE`     | `false`                   | Disable authentication                                                                                           |
| `SERVER_AUTH_BASIC_CREDS` |                           | List of comma separated `login:password` pairs. Passwords must be hashed using BCrypt (`htpasswd -nB <user>`)    |
| `SERVER_ENABLE_PROFILING` | `false`                   | Enable [pprof](https://blog.golang.org/pprof) handlers. You can find handler urls [here](internal/web/routes.go) |

## Development

### Commands

#### Run

```bash
# Run the app with installed Go and PostgreSQL in a Docker container
make

# Or run both the app and PostgreSQL in a Docker containers
make docker
```

#### Test

```bash
make test
```

#### More

You can find more commands in [Makefile](./Makefile)

### Tools

#### Linter

[golangci-lint](https://github.com/golangci/golangci-lint) can be used to lint the code. Just run `make lint`. Config can be found [here](./.golangci.yml)

#### API documentation

[swag](https://github.com/swaggo/swag) is used to generate API documentation. You can find more information about API endpoints in section [API](#api)

### Endpoints

#### Pages

You can find screenshots of pages [here](./docs/images/README.md)

- `/months` - Last 12 months
- `/months/month?year={year}&month={month}` - Month info
- `/search/spends` - Search for Spends

#### API

You can find Swagger 2.0 Documentation [here](docs/swagger.yaml). Use [Swagger Editor](https://editor.swagger.io/) to view it
