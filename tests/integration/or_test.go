package integration

import (
	"context"
	"testing"
)

func TestOrWhere(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("name = ?", "Alice").
			OrWhere("name = ?", "Charlie").
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
			_ = rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(names))
		}
		if names[0] != "Alice" {
			t.Errorf("got names[0] = %q, want 'Alice'", names[0])
		}
		if names[1] != "Charlie" {
			t.Errorf("got names[1] = %q, want 'Charlie'", names[1])
		}
	})
}

func TestOrWhereWithAnd(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("active = ?", true).
			Where("name = ?", "Alice").
			OrWhere("name = ?", "Charlie").
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
			_ = rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Fatalf("expected 2 rows (Alice active, OR Charlie regardless), got %d", len(names))
		}
	})
}

func TestOrWhereUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Update("users").
			Set("email", "shared@example.com").
			Where("name = ?", "Alice").
			OrWhere("name = ?", "Bob").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 2 {
			t.Errorf("expected 2 rows affected, got %d", rows)
		}

		qb = td.builder()
		sq, sa, err := qb.Select("COUNT(*)").
			From("users").
			Where("email = ?", "shared@example.com").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var count int
		_ = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&count)
		if count != 2 {
			t.Errorf("expected 2 users with shared email, got %d", count)
		}
	})
}
