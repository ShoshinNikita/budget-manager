# Budget Manager

**Budger Manager** is an easy-to-use, lightweight and self-hosted solution to track your finances

![Month Page](./docs/images/month_page_large.png)

It was inspired by [Poor-Man's Budgeting Spreadsheet](https://www.reddit.com/r/personalfinance/comments/2tymvf/poormans_budgeting_spreadsheet/) and [You have less money than you think (rus)](https://journal.tinkoff.ru/spreadsheet/). These projects have a fatal flaw: you can't add multiple spends in a single day. This project resolves this issue

**Features:**

- **Easy-to-use** - simple and intuitive UI

- **Lightweight** - backend is written on [Go](https://golang.org/), HTML is rendered with [Go templates](https://golang.org/pkg/text/template/). Vanilla JavaScript is used just to make frontend interactive. So, JS code is very primitive and lightweight: it won't devour all your CPU and RAM (even with Chrome 😉)

- **Self-hosted** - you don't need to trust any proprietary software to store your financial information

You can find more screenshots [here](./docs/images/README.md)

***

- [Install](#install)
- [Configuration](#configuration)
- [Development](#development)
  - [Run](#run)
  - [Test](#test)
- [API](#api)
  - [General](#general)
  - [Income](#income)
  - [Monthly Payment](#monthly-payment)
  - [Spend](#spend)
  - [Spend Type](#spend-type)

## Install

You need [Docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/install/) (optional)

1. Create `docker-compose.yml` with the following content (you can find more setting in [Configuration](#configuration) section):

    ```yaml
    version: "2.4"

    services:
      budget-manager:
        image: docker.pkg.github.com/shoshinnikita/budget-manager/budget-manager:latest
        container_name: budget-manager
        environment:
          DB_TYPE: postgres
          DB_PG_HOST: postgres
          DB_PG_PORT: 5432
          DB_PG_USER: postgres
          DB_PG_PASSWORD: very_strong_password
          DB_PG_DATABASE: postgres
          SERVER_PORT: 8080
          SERVER_CREDENTIALS: your creadentials # more info in 'Configuration' section
        ports:
          - "8080:8080"

      postgres:
        image: postgres
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

| Env Var                  | Default value | Description                                                                                                                                                                                                                                                                                                          |
| ------------------------ | ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DEBUG`                  | `false`       | Is Debug Mode on                                                                                                                                                                                                                                                                                                     |
| `LOGGER_MODE`            | `prod`        | Logger mode. Available options: `prod` (or `production`), `dev` (or `develop`).                                                                                                                                                                                                                                      | **** |
| `LOGGER_LEVEL`           | `info`        | Min level of log messages. Available options: `debug`, `info`, `warn`, `error`, `fatal`.<br><br>**Note:** level is always `debug` when Debug Mode is on                                                                                                                                                              |
| `DB_TYPE`                | `postgres`    | Database type. Only `postgres` is available now                                                                                                                                                                                                                                                                      |
| `DB_PG_HOST`             | `localhost`   | Host for connection to the db                                                                                                                                                                                                                                                                                        |
| `DB_PG_PORT`             | `5432`        | Port for connection to the db                                                                                                                                                                                                                                                                                        |
| `DB_PG_USER`             | `postgres`    | Use for connection to the db                                                                                                                                                                                                                                                                                         |
| `DB_PG_PASSWORD`         |               | Password for connection to the db                                                                                                                                                                                                                                                                                    |
| `DB_PG_DATABASE`         | `postgres`    | Database for connection                                                                                                                                                                                                                                                                                              |
| `SERVER_PORT`            | `8080`        |                                                                                                                                                                                                                                                                                                                      |
| `SERVER_CACHE_TEMPLATES` | `true`        | Defines whether templates have to be loaded from disk every request. It is always `false` in Debug mode                                                                                                                                                                                                              |
| `SERVER_CREDENTIALS`     |               | List of comma separated `login:password` pairs. Password must be encrypted with MD5 algorithm (you can use this command `openssl passwd -apr1 YOUR_PASSWORD`).<br><br>More info about password encryption: [Password Formats - Apache HTTP Server](https://httpd.apache.org/docs/2.4/misc/password_encryptions.html) |
| `SERVER_SKIP_AUTH`       | `false`       | Disables authentication. Works only in Debug mode!                                                                                                                                                                                                                                                                   |

## Development

### Run

You can run a local version with `make run-docker` with Docker. If you don't want to use Docker, you can build and run **Budget Manager** with `make` (or `make run-local`)

After the launch you can use [`tools/api.rest`](tools/api.rest) file to make basic API requests. More info about the `.rest` and `.http` files:

- [REST Client Extension for VS Code](https://github.com/Huachao/vscode-restclient)
- [HTTP client in IntelliJ IDEA code editor](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html)

Also you can use [`tools/fill_db.go`](tools/fill_db.go) script to fill the DB. This script makes `POST` requests to create Incomes, Monthly Payments, Spends and Spend Types.

```bash
# Run Budget Manager
make
# Or
make run-docker

# Add test data
go run tools/fill_db.go
```

### Test

#### Unit tests

```bash
make test-unit
```

#### Integration tests

```bash
make test
# Or
make test-integ
```

## API

All endpoints return json response with `Content-Type: application/json` header.

Requests and responses can be found in [internal/web/models](internal/web/models/models.go) package

### General

- `GET /api/months` - get month

  **Request:** `models.GetMonthReq` or `models.GetMonthByYearAndMonthReq`  
  **Response:** `models.GetMonthResp` or `models.Response`

- `GET /api/days` - get day

  **Request:** `models.GetDayReq` or `models.GetDayByDate`  
  **Response:** `models.GetDayResp` or `models.Response`

### Income

- `POST /api/incomes` - add a new income

  **Request:** `models.AddIncomeReq`  
  **Response:** `models.AddIncomeResp` or `models.Response`

- `PUT /api/incomes` - edit existing income

  **Request:** `models.EditIncomeReq`  
  **Response:** `models.Response`

- `DELETE /api/incomes` - remove income

  **Request:** `models.RemoveIncomeReq`  
  **Response:** `models.Response`

### Monthly Payment

- `POST /api/monthly-payments` - add new Monthly Payment

  **Request:** `models.AddMonthlyPaymentReq`  
  **Response:** `models.AddMonthlyPaymentResp` or `models.Response`

- `PUT /api/monthly-payments` - edit existing Monthly Payment

  **Request:** `models.EditMonthlyPaymentReq`  
  **Response:** `models.Response`

- `DELETE /api/monthly-payments` - remove Monthly Payment

  **Request:** `models.DeleteMonthlyPaymentReq`  
  **Response:** `models.Response`

### Spend

- `POST /api/spends` - add new Spend

  **Request:** `models.AddSpendReq`  
  **Response:** `models.AddSpendResp` or `models.Response`

- `PUT /api/spends` - edit existing Spend

  **Request:** `models.EditSpendReq`  
  **Response:** `models.Response`

- `DELETE /api/spends` - remove Spend

  **Request:** `models.RemoveSpendReq`  
  **Response:** `models.Response`

### Spend Type

- `GET /api/spend-types` - get list of all Spend Types

  **Request**: -
  **Response**: `models.GetSpendTypesResp` or `models.Response`

- `POST /api/spend-types` - add new Spend Type

  **Request:** `models.AddSpendTypeReq`  
  **Response:** `models.AddSpendTypeResp` or `models.Response`

- `PUT /api/spend-types` - edit existing Spend Type

  **Request:** `models.EditSpendTypeReq`  
  **Response:** `models.Response`

- `DELETE /api/spend-types` - remove Spend Type

  **Request:** `models.RemoveSpendTypeReq`  
  **Response:** `models.Response`
