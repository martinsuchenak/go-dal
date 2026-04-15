package integration

import (
	"context"
	"fmt"
	"testing"
)

func TestDirectExecInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		result, err := td.dalDB.Insert("users").
			Set("name", "DirectUser").
			Set("email", "direct@example.com").
			Set("active", true).
			Exec(ctx)
		if err != nil {
			t.Fatalf("direct insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "DirectUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after direct insert failed: %v", err)
		}
		if name != "DirectUser" {
			t.Errorf("got name %q, want 'DirectUser'", name)
		}
	})
}

func TestDirectExecSelectQuery(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		rows, err := td.dalDB.Select("name", "email").
			From("users").
			Where("active = ?", true).
			OrderBy("name").
			Query(ctx)
		if err != nil {
			t.Fatalf("direct select query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name, email string
			if err := rows.Scan(&name, &email); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Fatalf("expected 2 active users, got %d", len(names))
		}
		if names[0] != "Alice" {
			t.Errorf("got names[0] = %q, want 'Alice'", names[0])
		}
		if names[1] != "Bob" {
			t.Errorf("got names[1] = %q, want 'Bob'", names[1])
		}
	})
}

func TestDirectExecSelectQueryRow(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		var name string
		err := td.dalDB.Select("name").
			From("users").
			Where("email = ?", "alice@example.com").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("direct select queryrow failed: %v", err)
		}
		if name != "Alice" {
			t.Errorf("got name %q, want 'Alice'", name)
		}
	})
}

func TestDirectExecUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		result, err := td.dalDB.Update("users").
			Set("email", "alice_direct@example.com").
			Where("name = ?", "Alice").
			Exec(ctx)
		if err != nil {
			t.Fatalf("direct update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var email string
		err = td.dalDB.Select("email").
			From("users").
			Where("name = ?", "Alice").
			QueryRow(ctx).Scan(&email)
		if err != nil {
			t.Fatalf("select after direct update failed: %v", err)
		}
		if email != "alice_direct@example.com" {
			t.Errorf("got email %q, want 'alice_direct@example.com'", email)
		}
	})
}

func TestDirectExecDelete(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id = (SELECT id FROM users WHERE name = 'Charlie')")

		result, err := td.dalDB.Delete("users").
			Where("name = ?", "Charlie").
			Exec(ctx)
		if err != nil {
			t.Fatalf("direct delete failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var count int
		err = td.dalDB.Select("COUNT(*)").
			From("users").
			QueryRow(ctx).Scan(&count)
		if err != nil {
			t.Fatalf("count after direct delete failed: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 users after delete, got %d", count)
		}
	})
}

func TestDirectExecInsertReturning(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		if td.name == "mysql" {
			result, err := td.dalDB.Insert("users").
				Set("name", "RetVal").
				Set("email", "ret@example.com").
				Set("active", true).
				Exec(ctx)
			if err != nil {
				t.Fatalf("insert failed: %v", err)
			}
			id, err := result.LastInsertId()
			if err != nil {
				t.Fatalf("last insert id failed: %v", err)
			}
			if id <= 0 {
				t.Errorf("expected positive id, got %d", id)
			}
			return
		}

		var id int
		row := td.dalDB.Insert("users").
			Set("name", "RetVal").
			Set("email", "ret@example.com").
			Set("active", true).
			Returning("id").
			QueryRow(ctx)
		if err := row.Scan(&id); err != nil {
			t.Fatalf("direct insert returning failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive id, got %d", id)
		}
	})
}

func TestDirectExecUpdateReturning(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			t.Skip("MySQL doesn't support RETURNING")
		}
		ctx := context.Background()

		var id int
		var name string
		row := td.dalDB.Update("users").
			Set("name", "UpdatedDirect").
			Where("id = ?", 1).
			Returning("id", "name").
			QueryRow(ctx)
		if err := row.Scan(&id, &name); err != nil {
			t.Fatalf("direct update returning failed: %v", err)
		}
		if id != 1 {
			t.Errorf("got id %d, want 1", id)
		}
		if name != "UpdatedDirect" {
			t.Errorf("got name %q, want 'UpdatedDirect'", name)
		}
	})
}

func TestDirectExecDeleteReturning(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			t.Skip("MySQL doesn't support RETURNING")
		}
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id = 3")
		_, _ = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('ToDel', 'del@test.com', %s)", td.dialect.BoolLiteral(false)))

		var id int
		var name string
		row := td.dalDB.Delete("users").
			Where("name = ?", "ToDel").
			Returning("id", "name").
			QueryRow(ctx)
		if err := row.Scan(&id, &name); err != nil {
			t.Fatalf("direct delete returning failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("got id %d, want positive", id)
		}
		if name != "ToDel" {
			t.Errorf("got name %q, want 'ToDel'", name)
		}
	})
}

func TestDirectExecNewQueryBuilder(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		qb := td.dalDB.NewQueryBuilder()
		if qb == nil {
			t.Fatal("expected non-nil QueryBuilder")
		}

		query, _, err := qb.Select("name").From("users").Build()
		if err != nil {
			t.Fatal(err)
		}
		if query == "" {
			t.Error("expected non-empty query")
		}
	})
}
