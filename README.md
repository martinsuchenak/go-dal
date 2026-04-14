# GO-DAL - GO Database Abstraction Layer

GO-DAL is a lightweight, interface-driven database abstraction layer for Go. Write database-agnostic SQL queries (Select, Insert, Update, Delete) across MySQL, PostgreSQL, SQLite, and SQL Server.

## Features

- Support for MySQL, PostgreSQL, SQLite, and SQL Server
- Interface-driven design for easy extensibility
- Fluent query builder with deterministic output
- Database driver wrapper for executing queries and scanning results
- Automatic placeholder conversion (`?`, `$1`, `@p1`) per database dialect
- Dialect interface for extensible SQL generation — add new databases without modifying shared code
- Quote-aware placeholder replacement (skips `?` inside string literals)
- Injectable structured logger (compatible with `github.com/fortix/go-libs/logger`)

## Installation

```bash
go get github.com/martinsuchenak/go-dal
```

## Quick Start

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
    "github.com/martinsuchenak/go-dal/pkg/dal"
    "github.com/martinsuchenak/go-dal/pkg/mysql"
)

func main() {
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    mysqlDB := mysql.NewMySQLDB(db)     // no logger, silent
    mysqlDB := mysql.NewMySQLDB(db, myLogger)  // with logger

    qb := mysql.NewQueryBuilder()

    query, args := qb.Select("id", "name", "email").
        From("users").
        Where("age > ?", 18).
        OrderBy("name").
        Limit(10).
        Build()

    rows, err := mysqlDB.Query(context.Background(), query, args...)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%d - %s <%s>\n", id, name, email)
    }
}
```

## Logging

All driver constructors accept an optional `dal.Logger` as the second argument. The interface is compatible with `github.com/fortix/go-libs/logger`:

```go
type Logger interface {
    Trace(msg string, keysAndValues ...any)
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
    Fatal(msg string, keysAndValues ...any)
}
```

### Usage

```go
// With no logger (silent, default)
db := mysql.NewMySQLDB(sqlDB)

// With a logger
db := mysql.NewMySQLDB(sqlDB, myLogger)

// Change logger at runtime
db.SetLogger(anotherLogger)

// Disable logging
db.SetLogger(nil)
```

### What gets logged

| Operation | Level | Fields |
|-----------|-------|--------|
| Exec start | Debug | `query`, `args` |
| Exec success | Debug | `query`, `duration` |
| Exec error | Error | `query`, `error`, `duration` |
| Query start | Debug | `query`, `args` |
| Query success | Debug | `query`, `duration` |
| Query error | Error | `query`, `error`, `duration` |
| QueryRow | Debug | `query`, `args` |
| BeginTx | Debug | — |
| BeginTx error | Error | `error` |
| Close | Debug | — |

### Using with fortix/go-libs/logger

```go
import "github.com/fortix/go-libs/logger"

myLogger := logger.New() // or any logger.Logger implementation
db := mysql.NewMySQLDB(sqlDB, myLogger)
```

The `dal.Logger` interface is structurally identical to `logger.Logger` -- no adapter needed.

## Query Builder

Each driver package provides `NewQueryBuilder()` pre-configured with the correct dialect. Under the hood, each driver creates a `Dialect` that controls SQL generation:

```go
// Using a driver package (recommended):
qb := mysql.NewQueryBuilder()

// Or create a QueryBuilder directly with a dialect:
d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
qb := dal.NewQueryBuilder(d)
```

### SELECT

```go
qb.Select("id", "name").
    From("users").
    Where("active = ?", true).
    OrderBy("name").
    Limit(10).
    Offset(20).
    Build()

qb.Select().From("users").Build()  // SELECT * FROM users
```

### INSERT

```go
qb.Insert("users").
    Set("name", "John Doe").
    Set("email", "john@example.com").
    Build()
```

### UPDATE

```go
qb.Update("users").
    Set("email", "new@example.com").
    Where("id = ?", 123).
    Build()
```

### DELETE

```go
qb.Delete("users").Where("id = ?", 123).Build()
qb.Delete("users").Build()  // DELETE FROM users
```

## Database Drivers

| Package | Placeholder Style | LIMIT/OFFSET | Constructor |
|---------|-------------------|--------------|-------------|
| `pkg/mysql` | `?` | `LIMIT X OFFSET Y` | `NewMySQLDB(db *sql.DB, log ...dal.Logger)` |
| `pkg/postgres` | `$1, $2, ...` | `LIMIT X OFFSET Y` | `NewPostgresDB(db *sql.DB, log ...dal.Logger)` |
| `pkg/sqlite` | `?` | `LIMIT X OFFSET Y` | `NewSQLiteDB(db *sql.DB, log ...dal.Logger)` |
| `pkg/mssql` | `@p1, @p2, ...` | `OFFSET X ROWS FETCH NEXT Y ROWS ONLY` | `NewMSSQLDB(db *sql.DB, log ...dal.Logger)` |

All drivers implement `dal.DBInterface` and support `SetLogger(dal.Logger)`.

## Dialect Interface

For databases not yet supported, implement the `dal.Dialect` interface:

```go
type Dialect interface {
    BuildSelect(q *SelectQuery) (string, []interface{})
    BuildInsert(q *InsertQuery) (string, []interface{})
    BuildUpdate(q *UpdateQuery) (string, []interface{})
    BuildDelete(q *DeleteQuery) (string, []interface{})
}
```

`BaseDialect` handles the common case — embed it and override only what differs:

```go
type MyDialect struct {
    dal.BaseDialect
}

func (d *MyDialect) BuildSelect(q *dal.SelectQuery) (string, []interface{}) {
    // custom SELECT rendering
}
```

## Development

GO Version: 1.26.2
Package: `github.com/martinsuchenak/go-dal`
