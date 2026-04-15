# Examples

All examples use MySQL. The same code works with PostgreSQL, SQLite, or SQL Server — just swap the import and constructor:

```go
// MySQL
import "github.com/martinsuchenak/go-dal/pkg/mysql"
db := mysql.NewMySQLDB(sqlDB)
qb := mysql.NewQueryBuilder()

// PostgreSQL
import "github.com/martinsuchenak/go-dal/pkg/postgres"
db := postgres.NewPostgresDB(sqlDB)
qb := postgres.NewQueryBuilder()

// SQLite
import "github.com/martinsuchenak/go-dal/pkg/sqlite"
db := sqlite.NewSQLiteDB(sqlDB)
qb := sqlite.NewQueryBuilder()

// SQL Server
import "github.com/martinsuchenak/go-dal/pkg/mssql"
db := mssql.NewMSSQLDB(sqlDB)
qb := mssql.NewQueryBuilder()
```

---

## Setup

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
    "github.com/martinsuchenak/go-dal/pkg/mysql"
)

func main() {
    sqlDB, err := sql.Open("mysql", "root:pass@tcp(localhost:3306)/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer sqlDB.Close()

    db := mysql.NewMySQLDB(sqlDB)
    ctx := context.Background()

    // ... examples below use db and ctx
}
```

---

## SELECT Examples

### Basic select with WHERE

```go
qb := mysql.NewQueryBuilder()
query, args := qb.Select("id", "name", "email").
    From("users").
    Where("active = ?", true).
    Build()

rows, err := db.Query(ctx, query, args...)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name, email string
    rows.Scan(&id, &name, &email)
    fmt.Printf("%d: %s <%s>\n", id, name, email)
}
```

### Select all columns

```go
query, args := qb.SelectAll().From("users").Build()
```

### Select with multiple WHERE conditions

```go
query, args := qb.Select("name").
    From("users").
    Where("active = ?", true).
    Where("age > ?", 18).
    Where("country = ?", "US").
    Build()
// WHERE active = ? AND age > ? AND country = ?
```

### Select with OR

```go
query, args := qb.Select("name").
    From("users").
    Where("role = ?", "admin").
    OrWhere("role = ?", "moderator").
    Build()
// WHERE role = ? OR role = ?
```

### Select with IN clause

```go
query, args := qb.Select("name").
    From("users").
    Where("id IN (?)", dal.In(1, 2, 3, 4, 5)).
    Build()
// WHERE id IN (?, ?, ?, ?, ?)
```

### Select with JOIN

```go
query, args := qb.Select("u.name", "o.total_price").
    From("users u").
    Join("INNER JOIN orders o ON o.user_id = u.id").
    Where("o.total_price > ?", 100).
    Build()
```

### Select with three-table JOIN

```go
query, args := qb.Select("u.name", "p.name", "oi.quantity").
    From("order_items oi").
    Join("INNER JOIN orders o ON o.id = oi.order_id").
    Join("INNER JOIN users u ON u.id = o.user_id").
    Join("INNER JOIN products p ON p.id = oi.product_id").
    Build()
```

### Select with GROUP BY and HAVING

```go
query, args := qb.Select("country", "COUNT(*) as cnt").
    From("users").
    GroupBy("country").
    Having("COUNT(*) > ?", 5).
    Build()
```

### Select DISTINCT

```go
query, args := qb.Select("country").Distinct().From("users").Build()
// SELECT DISTINCT `country` FROM `users`
```

### Select with ORDER BY, LIMIT, OFFSET

```go
query, args := qb.Select("name").
    From("users").
    Where("active = ?", true).
    OrderBy("name ASC").
    Limit(10).
    Offset(20).
    Build()
// page 3 of 10-item pages
```

### Select single row

```go
query, args := qb.Select("name", "email").
    From("users").
    Where("id = ?", 42).
    Build()

var name, email string
err := db.QueryRow(ctx, query, args...).Scan(&name, &email)
if err == sql.ErrNoRows {
    fmt.Println("not found")
}
```

---

## INSERT Examples

### Single-row insert

```go
qb := mysql.NewQueryBuilder()
query, args := qb.Insert("users").
    Set("name", "Alice").
    Set("email", "alice@example.com").
    Set("active", true).
    Build()

result, err := db.Exec(ctx, query, args...)
if err != nil {
    log.Fatal(err)
}
id, _ := result.LastInsertId()
fmt.Println("inserted id:", id)
```

### Batch insert

```go
query, args := qb.Insert("users").
    Columns("name", "email").
    Values("Alice", "alice@example.com").
    Values("Bob", "bob@example.com").
    Values("Charlie", "charlie@example.com").
    Build()
// INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?), (?, ?)

result, err := db.Exec(ctx, query, args...)
```

### Insert with RETURNING (PostgreSQL)

```go
import "github.com/martinsuchenak/go-dal/pkg/postgres"

qb := postgres.NewQueryBuilder()
query, args := qb.Insert("users").
    Set("name", "Alice").
    Returning("id").
    Build()
// INSERT INTO "users" ("name") VALUES ($1) RETURNING id

var id int
db.QueryRow(ctx, query, args...).Scan(&id)
```

### Insert with OUTPUT (SQL Server)

```go
import "github.com/martinsuchenak/go-dal/pkg/mssql"

qb := mssql.NewQueryBuilder()
query, args := qb.Insert("users").
    Set("name", "Alice").
    Returning("id").
    Build()
// INSERT INTO [users] ([name]) VALUES (@p1) OUTPUT INSERTED.[id]

var id int
db.QueryRow(ctx, query, args...).Scan(&id)
```

---

## UPDATE Examples

### Basic update

```go
query, args := qb.Update("users").
    Set("email", "new@example.com").
    Where("id = ?", 42).
    Build()

result, err := db.Exec(ctx, query, args...)
rowsAffected, _ := result.RowsAffected()
```

### Update multiple columns

```go
query, args := qb.Update("users").
    Set("name", "Alice Smith").
    Set("email", "alice.smith@example.com").
    Set("updated_at", time.Now()).
    Where("id = ?", 1).
    Build()
```

### Update with OR in WHERE

```go
query, args := qb.Update("users").
    Set("active", false).
    Where("last_login < ?", "2023-01-01").
    OrWhere("status = ?", "banned").
    Build()
// UPDATE `users` SET `active` = ? WHERE last_login < ? OR status = ?
```

---

## DELETE Examples

### Delete with WHERE

```go
query, args := qb.Delete("users").
    Where("id = ?", 42).
    Build()

result, err := db.Exec(ctx, query, args...)
```

### Delete all

```go
query, args := qb.Delete("users").Build()
// DELETE FROM `users`
```

### Delete with OR

```go
query, args := qb.Delete("sessions").
    Where("expired_at < ?", time.Now()).
    OrWhere("user_id IS NULL").
    Build()
```

### Delete with IN clause

```go
query, args := qb.Delete("users").
    Where("id IN (?)", dal.In(1, 2, 3)).
    Build()
// DELETE FROM `users` WHERE id IN (?, ?, ?)
```

---

## Transaction Examples

### Commit

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

query, args := qb.Insert("orders").
    Set("user_id", 1).
    Set("total", 99.99).
    Build()

result, err := tx.Exec(ctx, query, args...)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

### Rollback on error

```go
tx, _ := db.BeginTx(ctx, nil)
defer tx.Rollback()

_, err := tx.Exec(ctx, "INSERT INTO orders (user_id, total) VALUES (?, ?)", 1, 99.99)
if err != nil {
    // Rollback is called by defer
    log.Fatal(err)
}

// Intentional rollback
tx.Rollback()
```

### Read within a transaction

```go
tx, _ := db.BeginTx(ctx, nil)
defer tx.Rollback()

tx.Exec(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1)
tx.Exec(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", 100, 2)

var balance float64
query, args := qb.Select("balance").From("accounts").Where("id = ?", 2).Build()
tx.QueryRow(ctx, query, args...).Scan(&balance)

tx.Commit()
```

---

## Logging Examples

### With a custom logger

```go
type myLogger struct{}

func (l *myLogger) Debug(msg string, kv ...any) {
    fmt.Println("[DEBUG]", msg, kv)
}
func (l *myLogger) Error(msg string, kv ...any) {
    fmt.Println("[ERROR]", msg, kv)
}
// ... implement Trace, Info, Warn, Fatal similarly

db := mysql.NewMySQLDB(sqlDB, &myLogger{})
```

### Using with fortix/go-libs/logger

```go
import "github.com/fortix/go-libs/logger"

myLogger := logger.New()
db := mysql.NewMySQLDB(sqlDB, myLogger)
// No adapter needed — the interfaces are structurally identical
```

### Change logger at runtime

```go
db.SetLogger(newLogger)
```

### Disable logging

```go
db.SetLogger(nil)
```
