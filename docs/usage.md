# Usage Guide

## Creating a Database Connection

Each driver wraps a standard `*sql.DB` and optionally accepts a logger:

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/martinsuchenak/go-dal/pkg/mysql"
)

sqlDB, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/mydb")

// Silent (default)
db := mysql.NewMySQLDB(sqlDB)

// With logging
db := mysql.NewMySQLDB(sqlDB, myLogger)
```

All drivers follow the same pattern:

```go
mysql.NewMySQLDB(sqlDB, logger...)
postgres.NewPostgresDB(sqlDB, logger...)
sqlite.NewSQLiteDB(sqlDB, logger...)
mssql.NewMSSQLDB(sqlDB, logger...)
```

### Health Check

```go
if err := db.Ping(ctx); err != nil {
    log.Fatal("database unreachable")
}
```

### Accessing the Underlying *sql.DB

```go
sqlDB := db.DB()
```

---

## Query Builder

Each driver provides `NewQueryBuilder()` pre-configured with the correct dialect:

```go
qb := mysql.NewQueryBuilder()
```

You can also construct one directly if needed:

```go
import "github.com/martinsuchenak/go-dal/pkg/dal"

d := &dal.BaseDialect{
    PlaceholderStyle: dal.QuestionMark,
    LimitStyle:       dal.LimitOffsetStyle,
    QuoteStyle:       dal.BacktickQuoting,
}
qb := dal.NewQueryBuilder(d)
```

### SELECT

```go
query, args := qb.Select("id", "name", "email").
    From("users").
    Where("active = ?", true).
    OrderBy("name").
    Limit(10).
    Offset(20).
    Build()
```

For `SELECT *`:

```go
query, args := qb.SelectAll().From("users").Build()
```

For `SELECT DISTINCT`:

```go
query, args := qb.Select("name").Distinct().From("users").Build()
```

### WHERE Clauses

Multiple `Where` calls are combined with AND:

```go
qb.Where("active = ?", true).Where("age > ?", 18)
// WHERE active = ? AND age > ?
```

Use `OrWhere` for OR conditions:

```go
qb.Where("active = ?", true).OrWhere("role = ?", "admin")
// WHERE active = ? OR role = ?
```

Mix AND and OR:

```go
qb.Where("a = ?", 1).OrWhere("b = ?", 2).Where("c = ?", 3)
// WHERE a = ? OR b = ? AND c = ?
```

### IN Clause

Use `dal.In()` to expand a single placeholder into multiple values:

```go
qb.Where("id IN (?)", dal.In(1, 2, 3))
// WHERE id IN (?, ?, ?) with args [1, 2, 3]
```

### JOINs

```go
qb.Select("u.name", "o.total").
    From("users u").
    Join("INNER JOIN orders o ON o.user_id = u.id").
    Where("o.total > ?", 100).
    Build()
```

Multiple joins:

```go
qb.Join("INNER JOIN orders o ON o.user_id = u.id").
    Join("INNER JOIN products p ON p.id = o.product_id")
```

### GROUP BY and HAVING

```go
qb.Select("active", "COUNT(*) as cnt").
    From("users").
    GroupBy("active").
    Having("COUNT(*) > ?", 1).
    Build()
```

### INSERT

Single row using `Set`:

```go
query, args := qb.Insert("users").
    Set("name", "Alice").
    Set("email", "alice@example.com").
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?)
```

Multi-row using `Columns` and `Values`:

```go
query, args := qb.Insert("users").
    Columns("name", "email").
    Values("Alice", "alice@example.com").
    Values("Bob", "bob@example.com").
    Values("Charlie", "charlie@example.com").
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?), (?, ?)
```

### INSERT RETURNING

For PostgreSQL (`RETURNING`) and SQL Server (`OUTPUT INSERTED.*`):

```go
query, args := qb.Insert("users").
    Set("name", "Alice").
    Returning("id").
    Build()

// PostgreSQL: INSERT INTO "users" ("name") VALUES ($1) RETURNING id
// MSSQL:      INSERT INTO [users] ([name]) VALUES (@p1) OUTPUT INSERTED.[id]
// MySQL:      INSERT INTO `users` (`name`) VALUES (?)  (no-op, use LastInsertId)
```

### UPDATE

```go
query, args := qb.Update("users").
    Set("email", "new@example.com").
    Set("active", false).
    Where("id = ?", 42).
    Build()
```

### DELETE

```go
query, args := qb.Delete("users").
    Where("active = ?", false).
    OrWhere("last_login < ?", "2020-01-01").
    Build()
```

Delete all rows (no WHERE):

```go
query, args := qb.Delete("users").Build()
// DELETE FROM `users`
```

### Executing Queries

```go
// Query builder produces (string, []interface{})
query, args := qb.Select("name").From("users").Where("id = ?", 1).Build()

// Execute
rows, err := db.Query(ctx, query, args...)
defer rows.Close()

var name string
var email string
for rows.Next() {
    rows.Scan(&name, &email)
}

// Single row
err := db.QueryRow(ctx, query, args...).Scan(&name)

// Write operations
result, err := db.Exec(ctx, query, args...)
```

---

## Transactions

`BeginTx` returns a `*dal.Tx` wrapper with logging. Use `tx.Exec`, `tx.Query`, `tx.QueryRow` instead of the raw `*sql.Tx` methods:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

tx.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Alice")
tx.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Bob")

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

---

## Logging

### Logger Interface

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

This is structurally identical to `github.com/fortix/go-libs/logger` — pass it directly, no adapter needed.

### Setting a Logger

```go
// At creation
db := mysql.NewMySQLDB(sqlDB, myLogger)

// Change at runtime
db.SetLogger(anotherLogger)

// Disable logging
db.SetLogger(nil)
```

### What Gets Logged

| Operation | Level | Fields |
|-----------|-------|--------|
| Exec start | Debug | `query`, `args` |
| Exec done | Debug | `query`, `duration` |
| Exec error | Error | `query`, `error`, `duration` |
| Query start | Debug | `query`, `args` |
| Query done | Debug | `query`, `duration` |
| Query error | Error | `query`, `error`, `duration` |
| QueryRow | Debug | `query`, `args` |
| BeginTx | Debug | — |
| Commit | Debug | — |
| Rollback | Debug | — |
| Ping | Debug | — |
| Close | Debug | — |

Transaction operations are prefixed with `tx ` (e.g., `tx exec`, `tx query`, `tx commit`).

---

## Dialect System

### How It Works

The query builder delegates SQL generation to a `Dialect` interface:

```go
type Dialect interface {
    BuildSelect(q *SelectQuery) (string, []interface{})
    BuildInsert(q *InsertQuery) (string, []interface{})
    BuildUpdate(q *UpdateQuery) (string, []interface{})
    BuildDelete(q *DeleteQuery) (string, []interface{})
    QuoteIdentifier(name string) string
}
```

Each query struct holds a `Dialect` reference. When you call `Build()`, it forwards to `dialect.BuildXxx(q)`.

### BaseDialect

`BaseDialect` handles the common case via three configuration fields:

```go
type BaseDialect struct {
    PlaceholderStyle PlaceholderStyle  // QuestionMark, DollarNumber, AtPNumber
    LimitStyle       LimitStyle        // LimitOffsetStyle, FetchNextStyle
    QuoteStyle       QuoteStyle        // NoQuoting, BacktickQuoting, DoubleQuoteQuoting, BracketQuoting
}
```

### Placeholder Translation

Write queries using `?` regardless of the database. The dialect translates:

| Dialect | Input | Output |
|---------|-------|--------|
| MySQL / SQLite | `WHERE id = ?` | `WHERE id = ?` |
| PostgreSQL | `WHERE id = ?` | `WHERE id = $1` |
| SQL Server | `WHERE id = ?` | `WHERE id = @p1` |

Placeholder numbering is automatic and correct across multiple WHERE/HAVING clauses.

The translation is **quote-aware** — `?` inside single-quoted or double-quoted string literals is left untouched:

```go
qb.Where("name = '?' AND id = ?", 42)
// PostgreSQL: WHERE name = '?' AND id = $1
```

### Identifier Quoting

The dialect automatically quotes table and column names based on `QuoteStyle`:

| QuoteStyle | MySQL | PostgreSQL | MSSQL |
|------------|-------|------------|-------|
| BacktickQuoting | `` `users` `` | — | — |
| DoubleQuoteQuoting | — | `"users"` | — |
| BracketQuoting | — | — | `[users]` |

Quoting is skipped for expressions containing spaces, parentheses, commas, or `AS` (aliases, function calls, raw SQL).

### LIMIT/OFFSET Handling

| LimitStyle | Generated SQL |
|------------|---------------|
| LimitOffsetStyle | `LIMIT 10 OFFSET 20` |
| FetchNextStyle | `OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY` |

---

## Portability Notes

The query builder handles mechanical differences (placeholders, quoting, LIMIT syntax) but does not abstract SQL expressions. Be aware of:

| Expression | MySQL | PostgreSQL | SQLite | MSSQL |
|------------|-------|------------|--------|-------|
| String concat | `CONCAT(a, b)` | `a \|\| b` | `a \|\| b` | `a + b` |
| Current timestamp | `NOW()` | `NOW()` | `datetime('now')` | `GETDATE()` |
| Boolean literal | `TRUE`/`FALSE` | `TRUE`/`FALSE` | `1`/`0` | `1`/`0` |

For cross-database code, stick to standard SQL expressions or use the `Dialect` interface to add expression helpers.
