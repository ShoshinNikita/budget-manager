# Budget Manager

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

## Configuration

| Env Var        | Default value | Description                                                                                                                                                              |
| -------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `DEBUG`        | `false`       | Is Debug Mode on                                                                                                                                                         |
| `LOGGER_MODE`  | `prod`        | Logger mode. Available options: `prod` (or `production`), `dev` (or `develop`). More info - [github.com/ShoshinNikita/go-clog](https://github.com/ShoshinNikita/go-clog) |
| `LOGGER_LEVEL` | `info`        | Min level of log messages. Available options: `debug`, `info`, `warn`, `error`, `fatal`.<br><br>**Note:** level is always `debug` when Debug Mode is on                  |
| `DB_HOST`      | `localhost`   | Host for connection to the db                                                                                                                                            |
| `DB_PORT`      | `5432`        | Port for connection to the db                                                                                                                                            |
| `DB_USER`      | `postgres`    | Use for connection to the db                                                                                                                                             |
| `DB_PASSWORD`  |               | Password for connection to the db                                                                                                                                        |
| `DB_DATABASE`  | `postgres`    | Database for connection                                                                                                                                                  |
| `SERVER_PORT`  | `8080`        |                                                                                                                                                                          |

## Development

### Run

You can run local version with `make run` (or `make run-docker`) with Docker. If you don't want to use Docker, you can build and run **Budget Manager** with `make run-local` (this target uses `go run` command)

After the launch you can use `api.rest` file to make basic API requests. More info about the `.rest` and `.http` files:

- [REST Client Extension for VS Code](https://github.com/Huachao/vscode-restclient)
- [HTTP client in IntelliJ IDEA code editor](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html)

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
