# Usage Guide

## Creating a Database Connection

Each driver wraps a standard `*sql.DB` and accepts a `xdal.Logger` (pass `nil` to disable logging):

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/martinsuchenak/xdal/pkg/mysql"
)

sqlDB, _ := sql.Open("mysql", "user:pass@tcp(localhost:3306)/mydb")

// Silent (no logging)
db := mysql.NewMySQLDB(sqlDB, nil)

// With logging
db := mysql.NewMySQLDB(sqlDB, myLogger)
```

All drivers follow the same pattern:

```go
mysql.NewMySQLDB(sqlDB, logger)
postgres.NewPostgresDB(sqlDB, logger)
sqlite.NewSQLiteDB(sqlDB, logger)
mssql.NewMSSQLDB(sqlDB, logger)
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

### Direct Execution

The `db` instance provides factory methods (`Select`, `Insert`, `Update`, `Delete`) that return query objects pre-wired for direct execution. No separate `QueryBuilder` needed:

```go
// INSERT — build + execute in one chain
result, err := db.Insert("users").
    Set("name", "Alice").
    Set("email", "alice@example.com").
    Exec(ctx)

// SELECT — build + query in one chain
rows, err := db.Select("id", "name").
    From("users").
    Where("active = ?", true).
    Query(ctx)

// SELECT single row
var name string
err := db.Select("name").
    From("users").
    Where("id = ?", 42).
    QueryRow(ctx).Scan(&name)

// UPDATE
result, err := db.Update("users").
    Set("email", "new@example.com").
    Where("id = ?", 42).
    Exec(ctx)

// DELETE
result, err := db.Delete("users").Where("id = ?", 42).Exec(ctx)

// INSERT with RETURNING (PostgreSQL)
var id int
err := db.Insert("users").
    Set("name", "Alice").
    Returning("id").
    QueryRow(ctx).Scan(&id)
```

You can still call `Build()` on any query for inspection, logging, or deferred execution:

```go
q := db.Select("name").From("users").Where("id = ?", 42)
query, args, err := q.Build()  // just build, don't execute
```

---

## Query Builder

Each driver provides `NewQueryBuilder()` pre-configured with the correct dialect:

```go
qb := mysql.NewQueryBuilder()
```

You can also construct one directly if needed:

```go
import "github.com/martinsuchenak/xdal/pkg/xdal"

d := &xdal.BaseDialect{
    Placeholder: xdal.QuestionMarkPlaceholder,
    AppendLimit: xdal.LimitOffset,
    QuoteStyle:  xdal.BacktickQuoting,
}
// Enable RETURNING (optional):
// d.AppendReturning = d.WriteReturning
qb := xdal.NewQueryBuilder(d)
```

### SELECT

```go
query, args, err := qb.Select("id", "name", "email").
    From("users").
    Where("active = ?", true).
    OrderBy("name").
    Limit(10).
    Offset(20).
    Build()
```

For `SELECT *`:

```go
query, args, err := qb.SelectAll().From("users").Build()
```

For `SELECT DISTINCT`:

```go
query, args, err := qb.Select("name").Distinct().From("users").Build()
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

### WHERE Groups

Use `WhereGroup` to create parenthesized groups of conditions:

```go
qb.Where("active = ?", true).
    WhereGroup(func(g *xdal.WhereGroup) {
        g.Where("role = ?", "admin").OrWhere("role = ?", "moderator")
    })
// WHERE active = ? AND (role = ? OR role = ?)
```

Use `OrWhereGroup` to combine a group with OR:

```go
qb.Where("a = ?", 1).
    OrWhereGroup(func(g *xdal.WhereGroup) {
        g.Where("b = ?", 2).Where("c = ?", 3)
    })
// WHERE a = ? OR (b = ? AND c = ?)
```

### WHERE Shortcuts

```go
// IS NULL
qb.WhereIsNull("deleted_at")
// WHERE deleted_at IS NULL

// IS NOT NULL
qb.WhereIsNotNull("email")
// WHERE email IS NOT NULL

// BETWEEN
qb.WhereBetween("age", 18, 65)
// WHERE age BETWEEN ? AND ?  with args [18, 65]
```

### IN Clause

Use `xdal.In()` to expand a single placeholder into multiple values:

```go
inVals, err := xdal.In(1, 2, 3)
if err != nil {
    return err
}
qb.Where("id IN (?)", inVals)
// WHERE id IN (?, ?, ?) with args [1, 2, 3]
```

`In()` returns an error if no values are provided or if the count exceeds 1000 (`xdal.MaxInValues`).

### JOINs

```go
query, args, err := qb.Select("u.name", "o.total").
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
query, args, err := qb.Select("active", "COUNT(*) as cnt").
    From("users").
    GroupBy("active").
    Having("COUNT(*) > ?", 1).
    Build()
```

Use `OrHaving` for OR conditions in the HAVING clause:

```go
qb.Having("COUNT(*) > ?", 10).OrHaving("AVG(score) < ?", 5)
// HAVING COUNT(*) > ? OR AVG(score) < ?
```

### INSERT

Single row using `Set`:

```go
query, args, err := qb.Insert("users").
    Set("name", "Alice").
    Set("email", "alice@example.com").
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?)
```

Single row using `SetMap` (keys are sorted for deterministic output):

```go
query, args, err := qb.Insert("users").
    SetMap(map[string]interface{}{
        "name":  "Alice",
        "email": "alice@example.com",
    }).
    Build()
// INSERT INTO `users` (`email`, `name`) VALUES (?, ?)
```

Single row using `SetStruct` (reads `db` struct tags, falls back to snake_case):

```go
type User struct {
    Name  string `db:"name"`
    Email string `db:"email"`
}

query, args, err := qb.Insert("users").
    SetStruct(User{Name: "Alice", Email: "alice@example.com"}).
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?)
```

Multi-row using `Columns` and `Values`:

```go
query, args, err := qb.Insert("users").
    Columns("name", "email").
    Values("Alice", "alice@example.com").
    Values("Bob", "bob@example.com").
    Values("Charlie", "charlie@example.com").
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?), (?, ?)
```

### INSERT RETURNING

For PostgreSQL (`RETURNING`), SQLite (`RETURNING`), and SQL Server (`OUTPUT INSERTED.*`):

```go
query, args, err := qb.Insert("users").
    Set("name", "Alice").
    Returning("id").
    Build()

// PostgreSQL: INSERT INTO "users" ("name") VALUES ($1) RETURNING "id"
// SQLite:     INSERT INTO "users" ("name") VALUES (?) RETURNING "id"
// MSSQL:      INSERT INTO [users] ([name]) OUTPUT INSERTED.[id] VALUES (@p1)
// MySQL:      returns ErrReturningNotSupported (use LastInsertId instead)
```

### UPDATE

```go
query, args, err := qb.Update("users").
    Set("email", "new@example.com").
    Set("active", false).
    Where("id = ?", 42).
    Build()
```

With `SetMap`:

```go
query, args, err := qb.Update("users").
    SetMap(map[string]interface{}{
        "email":  "new@example.com",
        "active": false,
    }).
    Where("id = ?", 42).
    Build()
```

With `SetStruct`:

```go
type UserUpdate struct {
    Email  string `db:"email"`
    Active bool   `db:"active"`
}

query, args, err := qb.Update("users").
    SetStruct(UserUpdate{Email: "new@example.com", Active: false}).
    Where("id = ?", 42).
    Build()
```

### UPDATE RETURNING

For PostgreSQL, SQLite, and SQL Server, use `Returning()` on UPDATE:

```go
query, args, err := qb.Update("users").
    Set("email", "new@example.com").
    Where("id = ?", 42).
    Returning("id", "email").
    Build()

// PostgreSQL: UPDATE "users" SET "email" = $1 WHERE "id" = $2 RETURNING "id", "email"
// MSSQL:      UPDATE [users] SET [email] = @p1 OUTPUT INSERTED.[id], INSERTED.[email] WHERE [id] = @p2
```

### DELETE

```go
query, args, err := qb.Delete("users").
    Where("active = ?", false).
    OrWhere("last_login < ?", "2020-01-01").
    Build()
```

Delete all rows (no WHERE):

```go
query, args, err := qb.Delete("users").Build()
// DELETE FROM `users`
```

### DELETE RETURNING

For PostgreSQL, SQLite, and SQL Server, use `Returning()` on DELETE:

```go
query, args, err := qb.Delete("users").
    Where("active = ?", false).
    Returning("id", "name").
    Build()

// PostgreSQL: DELETE FROM "users" WHERE "active" = $1 RETURNING "id", "name"
// MSSQL:      DELETE FROM [users] OUTPUT DELETED.[id], DELETED.[name] WHERE [active] = @p1
```

### SetStruct Tag Conventions

`SetStruct` reads struct fields using these rules:

| Tag | Behavior |
|-----|----------|
| `db:"column_name"` | Use `column_name` as the column |
| `db:"-"` | Skip this field |
| *(no tag)* | Use Go field name converted to `snake_case` |
| *(unexported)* | Always skipped |

Pointer fields with `nil` values are skipped (useful for optional UPDATE columns). Non-nil pointers are dereferenced automatically.

```go
type User struct {
    ID        string  `db:"id"`         // column: id
    FullName  string                     // column: full_name (snake_case)
    Email     string  `db:"email"`      // column: email
    Password  string  `db:"-"`          // skipped
    secret    string                    // skipped (unexported)
    LockedAt  *string `db:"locked_at"`  // skipped when nil, included when set
}
```

`SetMap` and `SetStruct` can be mixed with `Set` calls — they all append to the same column/value list.

### Executing Queries

```go
// Query builder produces (string, []interface{}, error)
query, args, err := qb.Select("name").From("users").Where("id = ?", 1).Build()
if err != nil {
    return err
}

// Execute
rows, err := db.Query(ctx, query, args...)
defer rows.Close()

var name string
var email string
for rows.Next() {
    rows.Scan(&name, &email)
}

// Single row
err = db.QueryRow(ctx, query, args...).Scan(&name)

// Write operations
result, err := db.Exec(ctx, query, args...)
```

### Using DBExecutor

Both `BaseDB` and `Tx` satisfy the `DBExecutor` interface, so you can write functions that work with either:

```go
func getUser(ctx context.Context, db xdal.DBExecutor, id int) (string, error) {
    qb := mysql.NewQueryBuilder()
    query, args, err := qb.Select("name").From("users").Where("id = ?", id).Build()
    if err != nil {
        return "", err
    }
    var name string
    err = db.QueryRow(ctx, query, args...).Scan(&name)
    return name, err
}
```

---

## Transactions

### Manual Transactions

`BeginTx` returns a `*xdal.Tx` wrapper with logging. Use `tx.Exec`, `tx.Query`, `tx.QueryRow` instead of the raw `*sql.Tx` methods:

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

### WithTx Helper

`WithTx` handles begin/commit/rollback automatically:

```go
err := db.WithTx(ctx, nil, func(tx *xdal.Tx) error {
    if _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Alice"); err != nil {
        return err
    }
    if _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES (?)", "Bob"); err != nil {
        return err
    }
    return nil
})
```

If the callback returns an error, the transaction is rolled back. If it returns `nil`, the transaction is committed.

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

### Controlling Argument Logging

By default, query arguments are **redacted** in log output (shown as `<redacted>`). To include actual values:

```go
db.SetLogArgs(true)
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
    BuildSelect(q *SelectQuery) (string, []interface{}, error)
    BuildInsert(q *InsertQuery) (string, []interface{}, error)
    BuildUpdate(q *UpdateQuery) (string, []interface{}, error)
    BuildDelete(q *DeleteQuery) (string, []interface{}, error)
    QuoteIdentifier(name string) string
    SupportsReturning() bool
    TranslateSQL(query string) string
    ConcatExpr(parts ...string) string
    LengthExpr(col string) string
    CurrentTimestamp() string
    BoolLiteral(v bool) string
    StringAggExpr(col, sep string) string
    RandExpr() string
}
```

Each query struct holds a `Dialect` reference. When you call `Build()`, it forwards to `dialect.BuildXxx(q)`.

### BaseDialect

`BaseDialect` handles the common case via configurable function fields and style flags:

```go
type BaseDialect struct {
    Placeholder       func(idx int) string                                           // e.g. QuestionMarkPlaceholder
    AppendLimit       func(b *strings.Builder, orderBy []string, limit, offset *int64) // e.g. LimitOffset
    AppendReturning   func(b *strings.Builder, columns []string)                     // e.g. d.WriteReturning
    PrependReturning  func(b *strings.Builder, columns []string)                     // e.g. d.WriteOutput
    QuoteStyle        QuoteStyle                                                     // BacktickQuoting, DoubleQuoteQuoting, etc.
    BackslashEscapes  bool                                                           // true for MySQL
}
```

Set RETURNING hooks after construction (can't reference methods in struct literal):

```go
// PostgreSQL/SQLite
d := &xdal.BaseDialect{...}
d.AppendReturning = d.WriteReturning

// MSSQL
d := &xdal.BaseDialect{...}
d.PrependReturning = d.WriteOutput   // INSERT: OUTPUT before VALUES
d.AppendReturning = d.WriteOutput    // UPDATE/DELETE: OUTPUT after WHERE

// MySQL — don't set any returning hooks
```

### Placeholder Functions

Write queries using `?` regardless of the database. The dialect translates:

| Function | Dialect | Input | Output |
|----------|---------|-------|--------|
| `QuestionMarkPlaceholder` | MySQL / SQLite | `WHERE id = ?` | `WHERE id = ?` |
| `DollarPlaceholder` | PostgreSQL | `WHERE id = ?` | `WHERE id = $1` |
| `AtPPlaceholder` | SQL Server | `WHERE id = ?` | `WHERE id = @p1` |

Placeholder numbering is automatic and correct across multiple WHERE/HAVING clauses.

The translation is **quote-aware** — `?` inside single-quoted or double-quoted string literals is left untouched:

```go
qb.Where("name = '?' AND id = ?", 42)
// PostgreSQL: WHERE name = '?' AND id = $1
```

### TranslateSQL

For SQL that can't be expressed through the query builder (e.g., column expressions like `col=col+1`, subqueries like `SELECT EXISTS(...)`), use `TranslateSQL` to translate `?` placeholders to the dialect's format:

```go
// Column expression — can't use Set() for col=col+1
query := qb.TranslateSQL("UPDATE users SET failed_login_attempts=failed_login_attempts+1, updated_at=" + qb.CurrentTimestamp() + " WHERE id = ?")
result, err := db.Exec(ctx, query, userID)

// Subquery
query := qb.TranslateSQL("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? AND active = ?)")
var exists bool
db.QueryRow(ctx, query, email, true).Scan(&exists)
```

| Dialect | Input | Output |
|---------|-------|--------|
| MySQL / SQLite | `WHERE id = ?` | `WHERE id = ?` |
| PostgreSQL | `WHERE id = ?` | `WHERE id = $1` |
| SQL Server | `WHERE id = ?` | `WHERE id = @p1` |

Like the query builder, `TranslateSQL` is quote-aware and skips `?` inside string literals. It's available on both `Dialect` and `QueryBuilder`:

```go
dialect.TranslateSQL(query)  // on the dialect directly
qb.TranslateSQL(query)       // convenience wrapper
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

| Function | Generated SQL |
|----------|---------------|
| `LimitOffset` | `LIMIT 10 OFFSET 20` |
| `FetchNextLimit` | `OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY` |

### RETURNING/OUTPUT Handling

| Hook | Dialects | INSERT | UPDATE | DELETE |
|------|----------|--------|--------|--------|
| none | MySQL | not supported | not supported | not supported |
| `AppendReturning = WriteReturning` | PostgreSQL, SQLite | `... VALUES (...) RETURNING col` | `... WHERE ... RETURNING col` | `... WHERE ... RETURNING col` |
| `PrependReturning = WriteOutput` | SQL Server | `... OUTPUT INSERTED.col VALUES (...)` | — | — |
| `AppendReturning = WriteOutput` | SQL Server | — | `... OUTPUT INSERTED.col WHERE ...` | — |
| `AppendDeletedReturning = WriteDeletedOutput` | SQL Server | — | — | `... OUTPUT DELETED.col WHERE ...` |

### SafeIdentifier

Use `xdal.SafeIdentifier()` to validate that a table or column name is safe (letters, digits, underscores, dots only):

```go
if err := xdal.SafeIdentifier(userInput); err != nil {
    return fmt.Errorf("invalid identifier: %w", err)
}
```

---

## Portability Helpers

SQL expression syntax varies across databases. Expression helpers are methods on the `Dialect` interface and are exposed through `QueryBuilder` convenience wrappers — no driver-specific imports needed:

```go
qb := mysql.NewQueryBuilder()

qb.Select(qb.ConcatExpr("first_name", "' '", "last_name"))
// MySQL/PostgreSQL/SQLite: CONCAT(first_name, ' ', last_name)
// MSSQL:                   first_name + ' ' + last_name

qb.Select(qb.LengthExpr("name"))
// MySQL/PostgreSQL/SQLite: LENGTH(name)
// MSSQL:                   LEN(name)

qb.Select(qb.CurrentTimestamp())
// MySQL/PostgreSQL: NOW()
// SQLite:           datetime('now')
// MSSQL:            GETDATE()

qb.Where("active = " + qb.BoolLiteral(true))
// MySQL/PostgreSQL: active = TRUE
// SQLite/MSSQL:     active = 1

qb.Select(qb.StringAggExpr("name", "', '")).GroupBy("team")
// MySQL:      GROUP_CONCAT(name SEPARATOR ', ')
// SQLite:     GROUP_CONCAT(name, ', ')
// PostgreSQL: STRING_AGG(name, ', ')
// MSSQL:      STRING_AGG(name, ', ')

qb.Select(qb.RandExpr())
// MySQL/MSSQL:      RAND()
// PostgreSQL/SQLite: RANDOM()
```

Each query builder is pre-configured with the correct dialect, so `qb.XxxExpr()` always generates the right SQL for your database.

### Still Portable (No Helper Needed)

These expressions work identically across all four databases:

- `COALESCE(a, b)`
- `UPPER(s)`, `LOWER(s)`
- `SUBSTRING(s, start, length)`
- `CAST(x AS type)`
- `COUNT(*)`, `SUM(x)`, `AVG(x)`, `MIN(x)`, `MAX(x)`

### Still Database-Specific

These vary too much for simple helpers and require database-specific SQL:

- Date/time arithmetic (`INTERVAL`, `DATEADD`, etc.)
- JSON operations
- Full-text search
- Window functions (syntax varies in edge cases)
