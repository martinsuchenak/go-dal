package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

func TestWhereIsNull(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "UPDATE orders SET created_at = NULL WHERE id = 1")

		query, args, err := td.builder().Select("id").
			From("orders").
			WhereIsNull("created_at").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}

		if len(ids) != 1 || ids[0] != 1 {
			t.Errorf("expected [1], got %v", ids)
		}
	})
}

func TestOrWhereGroup(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		query, args, err := td.builder().Select("name").
			From("users").
			Where("active = ?", true).
			OrWhereGroup(func(g *xdal.WhereGroup) {
				g.Where("name = ?", "Charlie").Where("email = ?", "charlie@example.com")
			}).
			OrderBy("name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			names = append(names, name)
		}

		if len(names) != 3 {
			t.Errorf("expected 3 rows (2 active + Charlie via OR group), got %d: %v", len(names), names)
		}
	})
}

func TestDeleteOrWhere(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id IN (1, 3)")

		query, args, err := td.builder().Delete("users").
			Where("name = ?", "Alice").
			OrWhere("name = ?", "Charlie").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("delete with or where failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 2 {
			t.Errorf("expected 2 rows affected, got %d", rows)
		}

		var count int
		sq, sa, err := td.builder().Select("COUNT(*)").From("users").Build()
		if err != nil {
			t.Fatal(err)
		}
		_ = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&count)
		if count != 1 {
			t.Errorf("expected 1 user remaining (Bob), got %d", count)
		}
	})
}

func TestLimitOnly(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		query, args, err := td.builder().Select("name").
			From("users").
			OrderBy("name").
			Limit(1).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("limit-only query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			_ = rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 1 {
			t.Errorf("expected 1 row with limit, got %d: %v", len(names), names)
		}
		if names[0] != "Alice" {
			t.Errorf("got names[0] = %q, want 'Alice'", names[0])
		}
	})
}

func TestOffsetOnly(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" || td.name == "sqlite" {
			t.Skip("MySQL and SQLite require LIMIT with OFFSET")
		}
		ctx := context.Background()

		query, args, err := td.builder().Select("name").
			From("users").
			OrderBy("name").
			Offset(1).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("offset-only query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			_ = rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Errorf("expected 2 rows after offset 1, got %d: %v", len(names), names)
		}
		if names[0] != "Bob" {
			t.Errorf("got names[0] = %q, want 'Bob'", names[0])
		}
	})
}

func TestPing(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		if err := td.dalDB.Ping(ctx); err != nil {
			t.Fatalf("ping failed: %v", err)
		}
	})
}

func TestErrEmptyTable(t *testing.T) {
	d := &xdal.BaseDialect{Placeholder: xdal.QuestionMarkPlaceholder}
	qb := xdal.NewQueryBuilder(d)

	_, _, err := qb.Select("id").Build()
	if err != xdal.ErrEmptyTable {
		t.Fatalf("select no table: got err %v, want ErrEmptyTable", err)
	}

	_, _, err = qb.Insert("").Set("id", 1).Build()
	if err != xdal.ErrEmptyTable {
		t.Errorf("insert empty table: got err %v, want ErrEmptyTable", err)
	}

	_, _, err = qb.Update("").Set("name", "x").Build()
	if err != xdal.ErrEmptyTable {
		t.Errorf("update empty table: got err %v, want ErrEmptyTable", err)
	}

	_, _, err = qb.Delete("").Build()
	if err != xdal.ErrEmptyTable {
		t.Errorf("delete empty table: got err %v, want ErrEmptyTable", err)
	}
}

func TestErrEmptyColumns(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		_, _, err := td.builder().Insert("users").Build()
		if err != xdal.ErrEmptyColumns {
			t.Errorf("got err %v, want ErrEmptyColumns", err)
		}

		_, _, err = td.builder().Update("users").Where("id = ?", 1).Build()
		if err != xdal.ErrEmptyColumns {
			t.Errorf("got err %v, want ErrEmptyColumns", err)
		}
	})
}

func TestErrEmptyInValues(t *testing.T) {
	_, err := xdal.In()
	if err != xdal.ErrEmptyInValues {
		t.Errorf("got err %v, want ErrEmptyInValues", err)
	}
}

func TestErrTooManyInValues(t *testing.T) {
	vals := make([]interface{}, 1001)
	_, err := xdal.In(vals...)
	if err != xdal.ErrTooManyInValues {
		t.Errorf("got err %v, want ErrTooManyInValues", err)
	}
}

func TestErrReturningNotSupported(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			_, _, err := td.builder().Insert("users").
				Set("name", "Test").
				Returning("id").
				Build()
			if err != xdal.ErrReturningNotSupported {
				t.Errorf("got err %v, want ErrReturningNotSupported", err)
			}
		}
	})
}

func TestErrBatchRowLength(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		_, _, err := td.builder().Insert("users").
			Columns("name", "email").
			Values("Alice", "alice@test.com").
			Values("Bob").
			Build()
		if err != xdal.ErrBatchRowLength {
			t.Errorf("got err %v, want ErrBatchRowLength", err)
		}
	})
}

func TestLoggingWithMockLogger(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		if td.name != "sqlite" {
			t.Skip("logging test only runs on sqlite for speed")
		}

		log := &mockLogger{}
		ctx := context.Background()

		sqlDB := td.db
		bdb := xdal.NewBaseDB(sqlDB, td.dialect, log)
		bdb.SetLogArgs(true)

		_, err := bdb.Exec(ctx, "INSERT INTO users (name, email, active) VALUES (?, ?, ?)", "LogUser", "log@test.com", true)
		if err != nil {
			t.Fatal(err)
		}

		if e := log.find("debug", "query exec"); e == nil {
			t.Error("expected 'query exec' log entry")
		}
		if !log.hasKey("query exec", "args") {
			t.Error("expected 'args' key in log entry")
		}
		if !log.hasKey("query exec done", "duration") {
			t.Error("expected 'duration' key in 'query exec done'")
		}
	})
}

func TestLoggingSetLogArgsRedacted(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		if td.name != "sqlite" {
			t.Skip("logging test only runs on sqlite for speed")
		}

		log := &mockLogger{}
		ctx := context.Background()

		sqlDB := td.db
		bdb := xdal.NewBaseDB(sqlDB, td.dialect, log)
		bdb.SetLogArgs(false)

		_, err := bdb.Exec(ctx, "INSERT INTO users (name, email, active) VALUES (?, ?, ?)", "RedactUser", "redact@test.com", true)
		if err != nil {
			t.Fatal(err)
		}

		e := log.find("debug", "query exec")
		if e == nil {
			t.Fatal("expected 'query exec' log entry")
		}
		for i := 0; i < len(e.kv)-1; i += 2 {
			if e.kv[i] == "args" {
				if e.kv[i+1] != "<redacted>" {
					t.Errorf("got args %v, want '<redacted>'", e.kv[i+1])
				}
				return
			}
		}
		t.Error("expected 'args' key in log entry")
	})
}

func TestUpdateWithoutWhere(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('A', 'a@t.com', %s)", td.dialect.BoolLiteral(true)))
		_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('B', 'b@t.com', %s)", td.dialect.BoolLiteral(true)))

		result, err := td.dalDB.Update("users").
			Set("email", "mass@example.com").
			Exec(ctx)
		if err != nil {
			t.Fatalf("mass update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows < 2 {
			t.Errorf("expected at least 2 rows affected, got %d", rows)
		}
	})
}

func TestDeleteWithoutWhere(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Del1', 'd1@t.com', %s)", td.dialect.BoolLiteral(true)))
		_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Del2', 'd2@t.com', %s)", td.dialect.BoolLiteral(true)))

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders")
		_, _ = td.dalDB.Exec(ctx, "DELETE FROM order_items")

		_, err := td.dalDB.Delete("users").
			Exec(ctx)
		if err != nil {
			t.Fatalf("mass delete failed: %v", err)
		}

		var count int
		_ = td.dalDB.Select("COUNT(*)").From("users").QueryRow(ctx).Scan(&count)
		if count != 0 {
			t.Errorf("expected 0 users after mass delete, got %d", count)
		}
	})
}

type mockLogger struct {
	entries []logEntry
}

type logEntry struct {
	level string
	msg   string
	kv    []any
}

func (m *mockLogger) Trace(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"trace", msg, kv})
}
func (m *mockLogger) Debug(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"debug", msg, kv})
}
func (m *mockLogger) Info(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"info", msg, kv})
}
func (m *mockLogger) Warn(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"warn", msg, kv})
}
func (m *mockLogger) Error(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"error", msg, kv})
}
func (m *mockLogger) Fatal(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"fatal", msg, kv})
}

func (m *mockLogger) find(level, msg string) *logEntry {
	for _, e := range m.entries {
		if e.level == level && e.msg == msg {
			return &e
		}
	}
	return nil
}

func (m *mockLogger) hasKey(msg string, key string) bool {
	for _, e := range m.entries {
		if e.msg == msg {
			for i := 0; i < len(e.kv)-1; i += 2 {
				if e.kv[i] == key {
					return true
				}
			}
		}
	}
	return false
}
