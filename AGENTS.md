# AGENTS.md

Instructions for LLM-based coding agents working on this project.

## Project Summary

GO-DAL is a Go library (not an application) providing a database abstraction layer with a fluent query builder and driver wrappers for MySQL, PostgreSQL, SQLite, and SQL Server.

- **Module**: `github.com/martinsuchenak/go-dal`
- **Go version**: 1.26.2
- **No main.go, no cmd/ directory** — this is a library only

## Architecture

```
pkg/dal/          Core package: types, query builder, dialect interface, logging
pkg/mysql/        MySQL driver (embeds BaseDB, configures BaseDialect)
pkg/postgres/     PostgreSQL driver
pkg/sqlite/       SQLite driver
pkg/mssql/        SQL Server driver
tests/integration/ Integration tests using Docker containers
```

### Key Types

| Type | File | Purpose |
|------|------|---------|
| `DBInterface` | `pkg/dal/types.go` | Interface for all drivers (Exec, Query, QueryRow, BeginTx, Ping, Close) |
| `BaseDB` | `pkg/dal/logger.go` | Shared implementation with structured logging; drivers embed this |
| `Tx` | `pkg/dal/logger.go` | Transaction wrapper with logging |
| `Dialect` | `pkg/dal/dialect.go` | Interface for SQL generation |
| `BaseDialect` | `pkg/dal/dialect.go` | Common implementation configured by PlaceholderStyle + LimitStyle + QuoteStyle |
| `QueryBuilder` | `pkg/dal/query_builder.go` | Fluent API, delegates Build() to Dialect |
| `SelectQuery` | `pkg/dal/types.go` | Fluent SELECT builder |
| `InsertQuery` | `pkg/dal/types.go` | Fluent INSERT builder (single-row and batch) |
| `UpdateQuery` | `pkg/dal/types.go` | Fluent UPDATE builder |
| `DeleteQuery` | `pkg/dal/types.go` | Fluent DELETE builder |
| `Logger` | `pkg/dal/logger.go` | Structured logging interface (compatible with fortix/go-libs/logger) |

### How Drivers Work

Each driver is ~26 lines: a struct embedding `*dal.BaseDB` (promoting all DBInterface methods), a constructor accepting optional `dal.Logger`, and `NewQueryBuilder()` using the driver's dialect. No forwarding methods — Go embedding handles promotion.

### How Dialects Work

Query structs hold a `Dialect` reference. `Build()` delegates to `dialect.BuildXxx(q)`. `BaseDialect` is configured with three fields (`PlaceholderStyle`, `LimitStyle`, `QuoteStyle`) and handles all four CRUD operations. Drivers can embed `BaseDialect` and override individual methods for database-specific behavior.

## Build & Test Commands

```bash
# Unit tests (no dependencies)
go test ./pkg/... -v

# Integration tests (requires Docker containers running)
docker compose -f docker-compose.test.yml up -d
go test ./tests/integration/ -v -timeout 120s

# Vet
go vet ./...

# Build (library, no binary)
go build ./...
```

Or use Taskfile:

```bash
task test              # unit tests
task docker-up         # start test databases
task test-integration  # integration tests
task docker-down       # stop test databases
```

## Docker Containers

Integration tests use three Docker containers:

| Database | Port | Credentials |
|----------|------|-------------|
| MySQL 8.0 | 13306 | root:testpass |
| PostgreSQL 16 | 15432 | godal:testpass |
| MSSQL 2022 | 11433 | sa:TestPass123! |

SQLite uses in-memory databases (no Docker).

## Coding Conventions

- GoDoc comments on exported types and methods only — no inline code comments unless requested
- Use Go struct embedding for shared behavior (never write forwarding methods)
- Every driver must have `var _ dal.DBInterface = (*XxxDB)(nil)` for compile-time check
- Fluent API: all query builder methods return their query type pointer
- Quote-aware placeholder replacement: `?` inside single/double-quoted strings must be skipped
- Use `t.Run(subtest)` for multi-database integration tests
- Use ordered slices (not maps) for column/value pairs to ensure deterministic SQL

## Common Tasks

### Adding a feature to the query builder

1. Add field(s) to the query struct in `pkg/dal/types.go`
2. Add fluent method(s) in `pkg/dal/query_builder.go`
3. Update `BaseDialect.BuildXxx()` in `pkg/dal/dialect.go`
4. Add unit tests in `pkg/dal/query_builder_test.go`
5. Add integration tests if the feature interacts with real databases

### Adding a new database driver

1. Create `pkg/yourdb/yourdb.go` — struct embedding `*dal.BaseDB`, constructor, `NewQueryBuilder`
2. Create `pkg/yourdb/dialect.go` — `NewDialect()` returning configured `BaseDialect`
3. Add interface compliance test: `var _ dal.DBInterface = (*YourDB)(nil)`
4. Add unit tests for the driver package
5. Add Docker service to `docker-compose.test.yml`
6. Add connection setup and tests in `tests/integration/`

See [docs/contributing.md](docs/contributing.md) for the full guide.

## Portability Gotchas

- PostgreSQL uses `$1, $2...`, MSSQL uses `@p1, @p2...`, MySQL/SQLite use `?`
- MSSQL uses `OFFSET X ROWS FETCH NEXT Y ROWS ONLY` instead of `LIMIT X OFFSET Y`
- PostgreSQL booleans: `TRUE`/`FALSE` in SQL; MSSQL: `1`/`0`; SQLite: `1`/`0`
- MSSQL requires creating the database explicitly after container startup
- `modernc.org/sqlite` registers as `"sqlite"` not `"sqlite3"`
- Identifier quoting: MySQL = backticks, PostgreSQL/SQLite = double quotes, MSSQL = brackets

## File Map

| File | What to change for |
|------|-------------------|
| `pkg/dal/types.go` | New query struct fields, interface changes |
| `pkg/dal/query_builder.go` | New fluent methods, In() helper |
| `pkg/dal/dialect.go` | SQL generation, quoting, placeholder translation |
| `pkg/dal/logger.go` | Logging, Tx wrapper, BaseDB methods |
| `pkg/*/yourdb.go` | Driver constructor, NewQueryBuilder |
| `pkg/*/dialect.go` | Driver-specific dialect config |
| `tests/integration/helpers.go` | Database connection setup, schema, seed data |
| `tests/integration/*.go` | Integration test cases |
| `docker-compose.test.yml` | Test database containers |
