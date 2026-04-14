# GO-DAL API Reference

## Package `dal`

Core types, query builder, and logging.

### `Logger`

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

Compatible with `github.com/fortix/go-libs/logger` -- no adapter needed.

### `NoopLogger`

Discards all log messages. Used when no logger is provided.

```go
log := dal.NoopLoggerInstance()
```

### `DBInterface`

```go
type DBInterface interface {
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
    BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
    Close() error
}
```

### `BaseDB`

Shared implementation used by all drivers. Wraps `*sql.DB` with logging.

```go
bdb := dal.NewBaseDB(db *sql.DB, log Logger)
bdb.SetLogger(log Logger)  // change or disable logging
bdb.DB() *sql.DB           // access underlying *sql.DB
```

### `PlaceholderStyle`

```go
const (
    QuestionMark  PlaceholderStyle = iota  // ? (MySQL, SQLite)
    DollarNumber                            // $1, $2, ... (PostgreSQL)
    AtPNumber                               // @p1, @p2, ... (SQL Server)
)
```

### `LimitStyle`

```go
const (
    LimitOffsetStyle  LimitStyle = iota  // LIMIT X OFFSET Y (MySQL, PostgreSQL, SQLite)
    FetchNextStyle                        // OFFSET X ROWS FETCH NEXT Y ROWS ONLY (SQL Server)
)
```

### `Dialect` Interface

```go
type Dialect interface {
    BuildSelect(q *SelectQuery) (string, []interface{})
    BuildInsert(q *InsertQuery) (string, []interface{})
    BuildUpdate(q *UpdateQuery) (string, []interface{})
    BuildDelete(q *DeleteQuery) (string, []interface{})
}
```

### `BaseDialect`

Common SQL generation used by all built-in drivers. Embed and override for databases with additional quirks.

```go
d := &dal.BaseDialect{
    PlaceholderStyle: dal.QuestionMark,
    LimitStyle:       dal.LimitOffsetStyle,
}
```

### `QueryBuilder`

```go
qb := dal.NewQueryBuilder(dialect)  // requires a Dialect
```

| Method | Returns | Description |
|--------|---------|-------------|
| `Select(columns ...string)` | `*SelectQuery` | Start a SELECT query |
| `Insert(table string)` | `*InsertQuery` | Start an INSERT query |
| `Update(table string)` | `*UpdateQuery` | Start an UPDATE query |
| `Delete(table string)` | `*DeleteQuery` | Start a DELETE query |

### `SelectQuery`

| Method | Description |
|--------|-------------|
| `From(table)` | Set the table |
| `Where(condition, args...)` | Add WHERE condition (AND-combined) |
| `OrderBy(columns...)` | Set ORDER BY |
| `Limit(limit)` | Set LIMIT |
| `Offset(offset)` | Set OFFSET |
| `Build() (string, []interface{})` | Build the query |

### `InsertQuery`

| Method | Description |
|--------|-------------|
| `Set(key, value)` | Add column-value pair (order preserved) |
| `Build() (string, []interface{})` | Build the query |

### `UpdateQuery`

| Method | Description |
|--------|-------------|
| `Set(key, value)` | Add column-value pair (order preserved) |
| `Where(condition, args...)` | Add WHERE condition (AND-combined) |
| `Build() (string, []interface{})` | Build the query |

### `DeleteQuery`

| Method | Description |
|--------|-------------|
| `Where(condition, args...)` | Add WHERE condition (AND-combined) |
| `Build() (string, []interface{})` | Build the query |

## Driver Packages

Each provides `NewXxxDB(db *sql.DB, log ...dal.Logger)`, `NewDialect() dal.Dialect`, and `NewQueryBuilder() *dal.QueryBuilder`.

| Package | DB Type | Placeholder | Dialect |
|---------|---------|-------------|---------|
| `pkg/mysql` | `MySQLDB` | `?` | `LimitOffsetStyle` |
| `pkg/postgres` | `PostgresDB` | `$1, $2, ...` | `LimitOffsetStyle` |
| `pkg/sqlite` | `SQLiteDB` | `?` | `LimitOffsetStyle` |
| `pkg/mssql` | `MSSQLDB` | `@p1, @p2, ...` | `FetchNextStyle` |
