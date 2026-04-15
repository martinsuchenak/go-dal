# Contributing to GO-DAL

## Development Setup

### Prerequisites

- Go 1.26.2+
- Docker and Docker Compose (for integration tests)
- [Task](https://taskfile.dev/) (build automation)

### Install Task

```bash
brew install go-task
```

### Project Structure

```
go-dal/
├── pkg/
│   ├── dal/                    # Core: types, query builder, dialect, logging
│   │   ├── types.go            # DBInterface, query structs, error sentinels
│   │   ├── query_builder.go    # QueryBuilder, fluent methods, In() helper
│   │   ├── dialect.go          # Dialect interface, BaseDialect, QuoteIdentifier
│   │   └── logger.go           # Logger interface, NoopLogger, BaseDB, Tx
│   ├── mysql/
│   │   ├── mysql.go            # MySQLDB (embeds BaseDB), NewQueryBuilder
│   │   └── dialect.go          # NewDialect() with MySQL config
│   ├── postgres/
│   │   ├── postgres.go         # PostgresDB (embeds BaseDB), NewQueryBuilder
│   │   └── dialect.go          # NewDialect() with PostgreSQL config
│   ├── sqlite/
│   │   ├── sqlite.go           # SQLiteDB (embeds BaseDB), NewQueryBuilder
│   │   └── dialect.go          # NewDialect() with SQLite config
│   └── mssql/
│       ├── mssql.go            # MSSQLDB (embeds BaseDB), NewQueryBuilder
│       └── dialect.go          # NewDialect() with SQL Server config
├── tests/
│   └── integration/            # Integration tests (CRUD, JOINs, aggregation, transactions)
├── docs/                       # Documentation
├── docker-compose.test.yml     # MySQL 8.0, PostgreSQL 16, MSSQL 2022
├── Taskfile.yml                # Build tasks
└── go.mod
```

---

## Running Tests

### Unit Tests

No external dependencies required:

```bash
task test
# or: go test ./pkg/... -v
```

### Integration Tests

Start the Docker containers first:

```bash
task docker-up
# or: docker compose -f docker-compose.test.yml up -d
```

Run integration tests:

```bash
task test-integration
# or: go test ./tests/integration/ -v -timeout 120s
```

Stop containers:

```bash
task docker-down
```

### All Tests

```bash
task test && task docker-up && task test-integration
```

---

## Architecture

### How the Pieces Fit

1. **`DBInterface`** (`pkg/dal/types.go`) — contract for all drivers (Exec, Query, QueryRow, BeginTx, Ping, Close)
2. **`BaseDB`** (`pkg/dal/logger.go`) — shared implementation of `DBInterface` with structured logging
3. **Driver structs** (`pkg/mysql/mysql.go`, etc.) — embed `*dal.BaseDB`, get all methods promoted automatically
4. **`Dialect`** (`pkg/dal/dialect.go`) — interface for SQL generation, implemented by `BaseDialect`
5. **`QueryBuilder`** (`pkg/dal/query_builder.go`) — fluent API that delegates `Build()` to the dialect
6. **`Tx`** (`pkg/dal/logger.go`) — transaction wrapper with logging

### Data Flow

```
User code
  │
  ▼
QueryBuilder.Select("name").From("users").Where("id = ?", 1).Build()
  │
  ▼
SelectQuery.Build() ──► dialect.BuildSelect(q)
  │
  ▼
BaseDialect.BuildSelect() ──► "SELECT `name` FROM `users` WHERE id = ?", []interface{}{1}, nil
  │
  ▼
MySQLDB.Query(ctx, query, args...) ──► BaseDB.Query() ──► *sql.DB.QueryContext()
                                        │
                                        └── logs at Debug/Error levels
```

### Why Drivers Are So Small

Each driver is ~27 lines because:

- **`BaseDB` embedding** — Go's method promotion means `MySQLDB` automatically has Exec, Query, QueryRow, BeginTx, Ping, Close, SetLogger, DB. No forwarding methods needed.
- **`BaseDialect` configuration** — all SQL generation is in one place. Each driver just provides a function-field config.

---

## Adding a New Database Driver

Adding support for a new database (e.g., Oracle, CockroachDB, ClickHouse) requires:

### 1. Create the Driver Package

Create `pkg/yourdb/yourdb.go`:

```go
package yourdb

import (
    "database/sql"
    "github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time interface check
var _ dal.DBInterface = (*YourDB)(nil)

// YourDB wraps a *sql.DB with your-database-specific query building.
type YourDB struct {
    *dal.BaseDB
}

// NewYourDB creates a new YourDB. Pass nil for log to disable logging.
func NewYourDB(db *sql.DB, log dal.Logger) *YourDB {
    return &YourDB{BaseDB: dal.NewBaseDB(db, log)}
}
```

### 2. Create the Dialect

Create `pkg/yourdb/dialect.go`:

```go
package yourdb

import "github.com/martinsuchenak/go-dal/pkg/dal"

func NewDialect() dal.Dialect {
    d := &dal.BaseDialect{
        Placeholder: dal.QuestionMarkPlaceholder, // or DollarPlaceholder, AtPPlaceholder
        AppendLimit: dal.LimitOffset,             // or FetchNextLimit
        QuoteStyle:  dal.DoubleQuoteQuoting,      // or BacktickQuoting, BracketQuoting, NoQuoting
    }
    // Enable RETURNING if supported:
    // d.AppendReturning = d.WriteReturning   // PostgreSQL/SQLite style
    // d.PrependReturning = d.WriteOutput     // MSSQL style (before VALUES)
    return d
}
```

The function-field pattern (`Placeholder`, `AppendLimit`, `AppendReturning`, `PrependReturning`) is more extensible than enum-based approaches — you can provide a custom function for databases with unique placeholder, LIMIT, or RETURNING syntax without modifying core code.

### 3. Add NewQueryBuilder

Add to `pkg/yourdb/yourdb.go`:

```go
func NewQueryBuilder() *dal.QueryBuilder {
    return dal.NewQueryBuilder(NewDialect())
}
```

### 4. If BaseDialect Doesn't Cover It

If your database has quirks beyond what `BaseDialect` handles (e.g., different identifier quoting rules, special INSERT syntax), embed `BaseDialect` and override the specific method:

```go
type YourDialect struct {
    dal.BaseDialect
}

func (d *YourDialect) BuildInsert(q *dal.InsertQuery) (string, []interface{}, error) {
    // custom INSERT rendering
    // you can call d.BaseDialect.BuildInsert(q) for the default behavior
    // and then modify the result
}
```

Then use your custom dialect in `NewDialect()`:

```go
func NewDialect() dal.Dialect {
    d := &YourDialect{
        BaseDialect: dal.BaseDialect{
            Placeholder: dal.DollarPlaceholder,
            AppendLimit: dal.LimitOffset,
            QuoteStyle:  dal.DoubleQuoteQuoting,
        },
    }
    d.AppendReturning = d.WriteReturning
    return d
}
```

### 5. Add Tests

Create `pkg/yourdb/yourdb_test.go`:

```go
package yourdb

import (
    "testing"
    "github.com/martinsuchenak/go-dal/pkg/dal"
)

func TestNewQueryBuilderUsesCorrectPlaceholders(t *testing.T) {
    qb := NewQueryBuilder()
    query, args, err := qb.Insert("users").Set("name", "John").Build()
    if err != nil {
        t.Fatal(err)
    }

    expected := `INSERT INTO "users" ("name") VALUES (?)`
    if query != expected {
        t.Errorf("got %q, want %q", query, expected)
    }
    if len(args) != 1 || args[0] != "John" {
        t.Errorf("got args %v, want [John]", args)
    }
}

func TestInterfaceCompliance(t *testing.T) {
    var _ dal.DBInterface = (*YourDB)(nil)
}
```

### 6. Add to Docker Compose (for integration tests)

Add a service to `docker-compose.test.yml` for your database, then add connection setup in `tests/integration/helpers.go`.

---

## Coding Conventions

- **No comments on code** — GoDoc comments on exported types and methods only
- **Embed, don't forward** — use Go struct embedding for shared behavior (BaseDB, BaseDialect)
- **Compile-time checks** — `var _ Interface = (*Type)(nil)` in every driver
- **Fluent API** — all query builder methods return `*XxxQuery` for chaining
- **Quote-aware** — placeholder replacement must skip `?` inside string literals
- **Table-driven tests** — use subtests with `t.Run()` for multi-database test cases
- **Handle errors** — `Build()` and `In()` now return errors; always check them

## Available Taskfile Commands

| Command | Description |
|---------|-------------|
| `task test` | Run unit tests |
| `task test-integration` | Run integration tests (requires Docker) |
| `task docker-up` | Start test databases |
| `task docker-down` | Stop test databases |
