package integration

import (
	"context"
	"testing"
)

func TestTranslateSQLSelect(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		translated := td.dialect.TranslateSQL("SELECT name FROM users WHERE active = ? AND name LIKE ?")

		rows, err := td.dalDB.Query(ctx, translated, true, "A%")
		if err != nil {
			t.Fatalf("translated SQL query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			names = append(names, name)
		}

		if len(names) != 1 || names[0] != "Alice" {
			t.Errorf("expected [Alice], got %v", names)
		}
	})
}

func TestTranslateSQLInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		translated := td.dialect.TranslateSQL("INSERT INTO users (name, email, active) VALUES (?, ?, ?)")

		result, err := td.dalDB.Exec(ctx, translated, "TransUser", "trans@example.com", true)
		if err != nil {
			t.Fatalf("translated SQL insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "TransUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after translated insert failed: %v", err)
		}
		if name != "TransUser" {
			t.Errorf("got name %q, want 'TransUser'", name)
		}
	})
}

func TestTranslateSQLUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		translated := td.dialect.TranslateSQL("UPDATE users SET email = ? WHERE name = ?")

		result, err := td.dalDB.Exec(ctx, translated, "alice_trans@example.com", "Alice")
		if err != nil {
			t.Fatalf("translated SQL update failed: %v", err)
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
			t.Fatalf("select after translated update failed: %v", err)
		}
		if email != "alice_trans@example.com" {
			t.Errorf("got email %q, want 'alice_trans@example.com'", email)
		}
	})
}

func TestTranslateSQLDelete(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id = (SELECT id FROM users WHERE name = 'Charlie')")

		translated := td.dialect.TranslateSQL("DELETE FROM users WHERE name = ?")

		result, err := td.dalDB.Exec(ctx, translated, "Charlie")
		if err != nil {
			t.Fatalf("translated SQL delete failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var count int
		err = td.dalDB.Select("COUNT(*)").
			From("users").
			Where("name = ?", "Charlie").
			QueryRow(ctx).Scan(&count)
		if err != nil {
			t.Fatalf("count after translated delete failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 Charlies after delete, got %d", count)
		}
	})
}

func TestTranslateSQLSkipsQuotedPlaceholders(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		translated := td.dialect.TranslateSQL("SELECT name FROM users WHERE name = '?' OR email = ?")

		var name string
		err := td.dalDB.QueryRow(ctx, translated, "alice@example.com").Scan(&name)
		if err != nil {
			t.Fatalf("translated SQL with quoted placeholder failed: %v", err)
		}
		if name != "Alice" {
			t.Errorf("got name %q, want 'Alice'", name)
		}
	})
}

func TestTranslateSQLViaQB(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		qb := td.dalDB.NewQueryBuilder()
		translated := qb.TranslateSQL("SELECT name FROM users WHERE id = ?")

		var name string
		err := td.dalDB.QueryRow(ctx, translated, 1).Scan(&name)
		if err != nil {
			t.Fatalf("QB.TranslateSQL query failed: %v", err)
		}
		if name != "Alice" {
			t.Errorf("got name %q, want 'Alice'", name)
		}
	})
}

func TestTranslateSQLMultipleParams(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		translated := td.dialect.TranslateSQL("SELECT name, email FROM users WHERE active = ? AND name LIKE ? ORDER BY name")

		rows, err := td.dalDB.Query(ctx, translated, true, "%o%")
		if err != nil {
			t.Fatalf("translated SQL multi-param query failed: %v", err)
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

		if len(names) != 1 || names[0] != "Bob" {
			t.Errorf("expected [Bob], got %v", names)
		}
	})
}
