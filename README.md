# Cloudflare D1 Go Client ☁️ 

<p align="center">
<img src="https://raw.githubusercontent.com/crosleyzack/cloudflare-d1-go/refs/heads/master/.github/assets/gopher.png" width="250" alt="Cloudflare D1 Go"/>
</p>


<p align="center">
<a href="https://pkg.go.dev/github.com/crosleyzack/cloudflare-d1-go"><img src="https://pkg.go.dev/badge/github.com/crosleyzack/cloudflare-d1-go.svg" alt="Go Reference"></a>
<a href="https://goreportcard.com/report/github.com/crosleyzack/cloudflare-d1-go"><img src="https://goreportcard.com/badge/github.com/crosleyzack/cloudflare-d1-go" alt="Go Report Card"></a>
<img src="https://img.shields.io/github/go-mod/go-version/crosleyzack/cloudflare-d1-go" alt="Go Version">
<img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
</p>

This is a [fork](https://github.com/ashayas/cloudflare-d1-go) of [@ashayas](https://github.com/ashayas) original library. The original library is fantastic and should be considered above this one depending on your use case. This implementation is intended to be a lower level, thin wrapper over Cloudflare's D1 REST API for programs to implement higher level packages around. Changes include:

1. Use of templating for return values, allowing more fields of the response to be parsed by the library.
2. Implementation of the remaining methods from the D1 REST API, such as GetDB and UpdateDB
3. Removal of ConnectDB and addition of database ID to all methods. A map is added to correlate name to ID, but this simplifies working with multiple databases in my opinion.
4. Addition of a mock which uses a local sqlite database for ease of testing.

<hr>

- This is a lightweight Go client for the Cloudflare D1 database
- D1 is a cool serverless, zero-config, transactional SQL database built by [Cloudflare](https://www.cloudflare.com/) built for the edge and cost-effective

## Installation 📦

```bash
go get github.com/crosleyzack/cloudflare-d1-go
```

## Usage 💻

### Initialize the client 🔑

```go
client := cloudflare_d1_go.NewClient("account_id", "api_token")
```

### Query the database 🔍

```go
// Execute a SQL query with optional parameters
// query: SQL query string
// params: Array of parameter values to bind to the query (use ? placeholders in query)
client.QueryDB(ctx, "<database_id>", "SELECT * FROM users WHERE age > ?", []string{"18"})
```

### Create a table 📄

```go
client.QueryDB(ctx, "<database_id>", "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
```

### Remove a table 🗑️

```go
client.QueryDB(ctx, "<database_id>", "DROP TABLE users")
```

### List Of Methods

#### Database Management
- `NewClient(accountID, apiToken string) *Client` - Creates a new D1 client
- `CreateDB(ctx context.Context, name string) (*utils.APIResponse[D1Database], error)` - Create a new database in cloudflare
- `DeleteDB(ctx context.Context, dbID string) (*utils.APIResponse[DeleteResult], error)` - delete a database from cloudflare.
- `UpdateDB(ctx context.Context, dbID string, settings DBSettings) (*utils.APIResponse[D1Database], error)` - update the settings of a database in cloudflare.
- `GetDB(ctx context.Context, dbID string) (*utils.APIResponse[D1Database], error)` - Retrieve a database from cloudflare.
- `ListDB(ctx context.Context) (*utils.APIResponse[D1DatabaseList], error)` - list all databases in the cloudflare account.

#### Query Execution
- `QueryDB(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]QueryResult[any]], error)`
- `QueryDBRaw(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]QueryResult[any]], error)`

## Testing 
- Run `go test` to run the tests

## Contributing 🤝
Contributions are welcome! Please feel free to submit a Pull Request.

## License 📄
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support 💪
If you encounter any issues or have questions, please file an issue on GitHub.