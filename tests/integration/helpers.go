package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/martinsuchenak/go-dal/pkg/dal"
	"github.com/martinsuchenak/go-dal/pkg/mssql"
	"github.com/martinsuchenak/go-dal/pkg/mysql"
	"github.com/martinsuchenak/go-dal/pkg/postgres"
	"github.com/martinsuchenak/go-dal/pkg/sqlite"
	_ "modernc.org/sqlite"
)

type testDB struct {
	name    string
	db      *sql.DB
	dalDB   dal.DBInterface
	builder func() *dal.QueryBuilder
	dialect dal.Dialect
}

func connectWithRetry(t *testing.T, driver, dsn string, maxWait time.Duration) *sql.DB {
	t.Helper()
	deadline := time.Now().Add(maxWait)
	for {
		db, err := sql.Open(driver, dsn)
		if err != nil {
			if time.Now().After(deadline) {
				t.Fatalf("connect failed: %v", err)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return db
		}
		_ = db.Close()
		if time.Now().After(deadline) {
			t.Fatalf("ping failed after %v: %v", maxWait, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func setupAllDBs(t *testing.T) []*testDB {
	t.Helper()
	var dbs []*testDB

	// SQLite (always available)
	sqliteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sqlite open: %v", err)
	}
	sqliteDalDB := sqlite.NewSQLiteDB(sqliteDB, nil)
	sqliteQB := sqlite.NewQueryBuilder()
	dbs = append(dbs, &testDB{
		name:    "sqlite",
		db:      sqliteDB,
		dalDB:   sqliteDalDB,
		builder: func() *dal.QueryBuilder { return sqlite.NewQueryBuilder() },
		dialect: sqliteQB.Dialect(),
	})

	// MySQL
	mysqlDB := connectWithRetry(t, "mysql", "root:testpass@tcp(localhost:13306)/godal_test?parseTime=true", 10*time.Second)
	mysqlDalDB := mysql.NewMySQLDB(mysqlDB, nil)
	mysqlQB := mysql.NewQueryBuilder()
	dbs = append(dbs, &testDB{
		name:    "mysql",
		db:      mysqlDB,
		dalDB:   mysqlDalDB,
		builder: func() *dal.QueryBuilder { return mysql.NewQueryBuilder() },
		dialect: mysqlQB.Dialect(),
	})

	// PostgreSQL
	pgDB := connectWithRetry(t, "postgres", "postgres://godal:testpass@localhost:15432/godal_test?sslmode=disable", 10*time.Second)
	pgDalDB := postgres.NewPostgresDB(pgDB, nil)
	pgQB := postgres.NewQueryBuilder()
	dbs = append(dbs, &testDB{
		name:    "postgres",
		db:      pgDB,
		dalDB:   pgDalDB,
		builder: func() *dal.QueryBuilder { return postgres.NewQueryBuilder() },
		dialect: pgQB.Dialect(),
	})

	// MSSQL - connect to master first, create godal_test, then reconnect
	mssqlMaster := connectWithRetry(t, "sqlserver", "sqlserver://sa:TestPass123!@localhost:11433?encrypt=disable", 15*time.Second)
	_, _ = mssqlMaster.Exec("IF DB_ID('godal_test') IS NULL CREATE DATABASE godal_test")
	_ = mssqlMaster.Close()

	mssqlDB := connectWithRetry(t, "sqlserver", "sqlserver://sa:TestPass123!@localhost:11433?database=godal_test&encrypt=disable", 10*time.Second)
	mssqlDalDB := mssql.NewMSSQLDB(mssqlDB, nil)
	mssqlQB := mssql.NewQueryBuilder()
	dbs = append(dbs, &testDB{
		name:    "mssql",
		db:      mssqlDB,
		dalDB:   mssqlDalDB,
		builder: func() *dal.QueryBuilder { return mssql.NewQueryBuilder() },
		dialect: mssqlQB.Dialect(),
	})

	return dbs
}

func createSchema(t *testing.T, td *testDB) {
	t.Helper()
	ctx := context.Background()

	switch td.name {
	case "sqlite":
		_, _ = td.db.Exec("DROP TABLE IF EXISTS order_items")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS orders")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS products")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS users")
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL NOT NULL
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			total_price REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`)

	case "mysql":
		_, _ = td.db.Exec("DROP TABLE IF EXISTS order_items")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS orders")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS products")
		_, _ = td.db.Exec("DROP TABLE IF EXISTS users")
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE
		) ENGINE=InnoDB`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL
		) ENGINE=InnoDB`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		) ENGINE=InnoDB`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE order_items (
			id INT AUTO_INCREMENT PRIMARY KEY,
			order_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		) ENGINE=InnoDB`)

	case "postgres":
		_, _ = td.db.Exec("DROP TABLE IF EXISTS order_items; DROP TABLE IF EXISTS orders; DROP TABLE IF EXISTS products; DROP TABLE IF EXISTS users")
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			product_id INT NOT NULL REFERENCES products(id),
			quantity INT NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE order_items (
			id SERIAL PRIMARY KEY,
			order_id INT NOT NULL REFERENCES orders(id),
			product_id INT NOT NULL REFERENCES products(id),
			quantity INT NOT NULL
		)`)

	case "mssql":
		_, _ = td.db.Exec("IF OBJECT_ID('order_items', 'U') IS NOT NULL DROP TABLE order_items")
		_, _ = td.db.Exec("IF OBJECT_ID('orders', 'U') IS NOT NULL DROP TABLE orders")
		_, _ = td.db.Exec("IF OBJECT_ID('products', 'U') IS NOT NULL DROP TABLE products")
		_, _ = td.db.Exec("IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users")
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE users (
			id INT IDENTITY(1,1) PRIMARY KEY,
			name NVARCHAR(255) NOT NULL,
			email NVARCHAR(255) NOT NULL,
			active BIT NOT NULL DEFAULT 1
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE products (
			id INT IDENTITY(1,1) PRIMARY KEY,
			name NVARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE orders (
			id INT IDENTITY(1,1) PRIMARY KEY,
			user_id INT NOT NULL FOREIGN KEY REFERENCES users(id),
			product_id INT NOT NULL FOREIGN KEY REFERENCES products(id),
			quantity INT NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			created_at DATETIME DEFAULT GETDATE()
		)`)
		_, _ = td.dalDB.Exec(ctx, `CREATE TABLE order_items (
			id INT IDENTITY(1,1) PRIMARY KEY,
			order_id INT NOT NULL FOREIGN KEY REFERENCES orders(id),
			product_id INT NOT NULL FOREIGN KEY REFERENCES products(id),
			quantity INT NOT NULL
		)`)
	}
}

func seedData(t *testing.T, td *testDB) {
	t.Helper()
	ctx := context.Background()

	boolTrue := td.dialect.BoolLiteral(true)
	boolFalse := td.dialect.BoolLiteral(false)

	// Users
	_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Alice', 'alice@example.com', %s)", boolTrue))
	_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Bob', 'bob@example.com', %s)", boolTrue))
	_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Charlie', 'charlie@example.com', %s)", boolFalse))

	// Products
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO products (name, price) VALUES ('Widget', 9.99)")
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO products (name, price) VALUES ('Gadget', 24.99)")
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO products (name, price) VALUES ('Doohickey', 4.99)")

	// Orders
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO orders (user_id, product_id, quantity, total_price) VALUES (1, 1, 2, 19.98)")
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO orders (user_id, product_id, quantity, total_price) VALUES (1, 2, 1, 24.99)")
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO orders (user_id, product_id, quantity, total_price) VALUES (2, 2, 3, 74.97)")
	_, _ = td.dalDB.Exec(ctx, "INSERT INTO orders (user_id, product_id, quantity, total_price) VALUES (2, 3, 5, 24.95)")
}

func runForEachDB(t *testing.T, fn func(t *testing.T, td *testDB)) {
	t.Helper()
	dbs := setupAllDBs(t)
	for _, td := range dbs {
		t.Run(td.name, func(t *testing.T) {
			createSchema(t, td)
			defer func() { _ = td.dalDB.Close() }()
			fn(t, td)
		})
	}
}

func runForEachDBWithSeed(t *testing.T, fn func(t *testing.T, td *testDB)) {
	t.Helper()
	dbs := setupAllDBs(t)
	for _, td := range dbs {
		t.Run(td.name, func(t *testing.T) {
			createSchema(t, td)
			seedData(t, td)
			defer func() { _ = td.dalDB.Close() }()
			fn(t, td)
		})
	}
}
