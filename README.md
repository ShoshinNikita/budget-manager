# Budget Manager

- [Configuration](#configuration)
- [API](#api)
  - [Income](#income)
  - [Monthly Payment](#monthly-payment)
  - [Spend](#spend)
  - [Spend Type](#spend-type)

## Configuration

| Env Var        | Default value | Description                                                                                                                                             |
| -------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DEBUG`        | `false`       | Is Debug Mode on                                                                                                                                        |
| `LOGGER_LEVEL` | `info`        | Min level of log messages. Available options: `debug`, `info`, `warn`, `error`, `fatal`.<br><br>**Note:** level is always `debug` when Debug Mode is on |
| `DB_HOST`      | `localhost`   | Host for connection to the db                                                                                                                           |
| `DB_PORT`      | `5432`        | Port for connection to the db                                                                                                                           |
| `DB_USER`      | `postgres`    | Use for connection to the db                                                                                                                            |
| `DB_PASSWORD`  |               | Password for connection to the db                                                                                                                       |
| `DB_DATABASE`  | `postgres`    | Database for connection                                                                                                                                 |
| `SERVER_PORT`  | `:8080`       |                                                                                                                                                         |

## API

All endpoints return json response with `Content-Type: application/json` header.

Requests and responses can be found in [internal/web/models](internal/web/models/models.go) package

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

- `POST /api/spend-types` - add new Spend Type

  **Request:** `models.AddSpendTypeReq`  
  **Response:** `models.AddSpendTypeResp` or `models.Response`

- `PUT /api/spend-types` - edit existing Spend Type

  **Request:** `models.EditSpendTypeReq`  
  **Response:** `models.Response`

- `DELETE /api/spend-types` - remove Spend Type

  **Request:** `models.RemoveSpendTypeReq`  
  **Response:** `models.Response`
