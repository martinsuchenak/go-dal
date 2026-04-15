package integration

import (
	"context"
	"testing"
)

func TestBatchInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Insert("users").
			Columns("name", "email", "active").
			Values("Dave", "dave@example.com", true).
			Values("Eve", "eve@example.com", true).
			Values("Frank", "frank@example.com", false).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("batch insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 3 {
			t.Errorf("expected 3 rows affected, got %d", rows)
		}

		qb = td.builder()
		sq, sa, err := qb.Select("COUNT(*)").From("users").Build()
		if err != nil {
			t.Fatal(err)
		}

		var count int
		td.dalDB.QueryRow(ctx, sq, sa...).Scan(&count)
		if count != 3 {
			t.Errorf("expected 3 users, got %d", count)
		}
	})
}

func TestBatchInsertThenSelect(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		insertQ, insertArgs, err := qb.Insert("users").
			Columns("name", "email", "active").
			Values("Alice", "alice@example.com", true).
			Values("Bob", "bob@example.com", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		_, err = td.dalDB.Exec(ctx, insertQ, insertArgs...)
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}

		qb = td.builder()
		selQ, selArgs, err := qb.Select("name").
			From("users").
			OrderBy("name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, selQ, selArgs...)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		defer rows.Close()

		var names []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Fatalf("expected 2 users, got %d", len(names))
		}
		if names[0] != "Alice" {
			t.Errorf("got names[0] = %q, want 'Alice'", names[0])
		}
		if names[1] != "Bob" {
			t.Errorf("got names[1] = %q, want 'Bob'", names[1])
		}
	})
}
