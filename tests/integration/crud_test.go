package integration

import (
	"context"
	"testing"
)

func TestInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Insert("users").
			Set("name", "Dave").
			Set("email", "dave@example.com").
			Set("active", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		var email string
		qb = td.builder()
		sq, sa, err := qb.Select("name", "email").
			From("users").
			Where("name = ?", "Dave").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		err = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&name, &email)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if name != "Dave" {
			t.Errorf("got name %q, want 'Dave'", name)
		}
		if email != "dave@example.com" {
			t.Errorf("got email %q, want 'dave@example.com'", email)
		}
	})
}

func TestSelect(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name", "email").
			From("users").
			Where("active = ?", true).
			OrderBy("name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

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

func TestUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Update("users").
			Set("email", "alice_new@example.com").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var email string
		qb = td.builder()
		sq, sa, err := qb.Select("email").
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		err = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&email)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if email != "alice_new@example.com" {
			t.Errorf("got email %q, want 'alice_new@example.com'", email)
		}
	})
}

func TestDelete(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		// Delete Charlie's orders first (FK constraint)
		qb := td.builder()
		rawSQL := "DELETE FROM orders WHERE user_id = (SELECT id FROM users WHERE name = 'Charlie')"
		td.dalDB.Exec(ctx, rawSQL)

		qb = td.builder()
		query, args, err := qb.Delete("users").
			Where("name = ?", "Charlie").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		qb = td.builder()
		sq, sa, err := qb.Select("COUNT(*)").
			From("users").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var count int
		td.dalDB.QueryRow(ctx, sq, sa...).Scan(&count)
		if count != 2 {
			t.Errorf("expected 2 users after delete, got %d", count)
		}
	})
}

func TestSelectWithLimitOffset(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			OrderBy("name").
			Limit(2).
			Offset(1).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		var names []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(names))
		}
		if names[0] != "Bob" {
			t.Errorf("got names[0] = %q, want 'Bob'", names[0])
		}
		if names[1] != "Charlie" {
			t.Errorf("got names[1] = %q, want 'Charlie'", names[1])
		}
	})
}
