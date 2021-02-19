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
          DB_PG_PASSWORD: very_strong_password
          DB_PG_DATABASE: postgres
          SERVER_PORT: 8080
          SERVER_CREDENTIALS: your credentials # more info in 'Configuration' section
        ports:
          - "8080:8080"

      postgres:
        image: postgres:12-alpine
        container_name: budget-manager_postgres
        environment:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: very_strong_password
          POSTGRES_DB: postgres
        volumes:
          # Store data in ./var/pg_data directory
          - type: bind
            source: ./var/pg_data
            target: /var/lib/postgresql/data
        command: -c "log_statement=all"
    ```

2. Run `docker-compose up -d`
3. Go to `http://localhost:8080` (change the port if needed)
4. Profit!

## Configuration

| Env Var                   | Default value | Description                                                                                                                                                                                                                                                                                                         |
| ------------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DEBUG`                   | `false`       | Is Debug Mode on                                                                                                                                                                                                                                                                                                    |
| `LOGGER_MODE`             | `prod`        | Logger mode. Available options: `prod` (or `production`), `dev` (or `develop`).                                                                                                                                                                                                                                     |
| `LOGGER_LEVEL`            | `info`        | Min level of log messages. Available options: `debug`, `info`, `warn`, `error`, `fatal`<br><br>**Note:** level is always `debug` when Debug Mode is on                                                                                                                                                              |
| `DB_TYPE`                 | `postgres`    | Database type. Only `postgres` is available now                                                                                                                                                                                                                                                                     |
| `DB_PG_HOST`              | `localhost`   | Host for connection to the db                                                                                                                                                                                                                                                                                       |
| `DB_PG_PORT`              | `5432`        | Port for connection to the db                                                                                                                                                                                                                                                                                       |
| `DB_PG_USER`              | `postgres`    | Use for connection to the db                                                                                                                                                                                                                                                                                        |
| `DB_PG_PASSWORD`          |               | Password for connection to the db                                                                                                                                                                                                                                                                                   |
| `DB_PG_DATABASE`          | `postgres`    | Database for connection                                                                                                                                                                                                                                                                                             |
| `SERVER_PORT`             | `8080`        |                                                                                                                                                                                                                                                                                                                     |
| `SERVER_USE_EMBED`        | `true`        | Defines whether server should use embedded templates and static files<br><br>**Note:** `false` value won't work for Docker container                                                                                                                                                                                |
| `SERVER_CREDENTIALS`      |               | List of comma separated `login:password` pairs. Password must be encrypted with MD5 algorithm (you can use this command `openssl passwd -apr1 YOUR_PASSWORD`)<br><br>More info about password encryption: [Password Formats - Apache HTTP Server](https://httpd.apache.org/docs/2.4/misc/password_encryptions.html) |
| `SERVER_SKIP_AUTH`        | `false`       | Disables authentication                                                                                                                                                                                                                                                                                             |
| `SERVER_ENABLE_PROFILING` | `false`       | Enable [pprof](https://blog.golang.org/pprof) handlers. You can find handler urls [here](internal/web/routes.go)                                                                                                                                                                                                    |

## Development

### Commands

#### Run

```bash
# Run the app with installed Go and PostgreSQL in Docker container
make

# Or run both the app and PostgreSQL in Docker containers
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
- `/{year}-{month}` - Month info
- `/search/spends` - Search for Spends

#### API

You can find Swagger 2.0 Documentation [here](docs/swagger.yaml). Use [Swagger Editor](https://editor.swagger.io/) to view it
