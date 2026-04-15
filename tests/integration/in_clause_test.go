package integration

import (
	"context"
	"testing"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

func TestInClause(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		inVals, err := xdal.In("Alice", "Charlie")
		if err != nil {
			t.Fatal(err)
		}

		query, args, err := qb.Select("name").
			From("users").
			Where("name IN (?)", inVals).
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

func TestInClauseWithOtherWhere(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		inVals, err := xdal.In("Alice", "Bob", "Charlie")
		if err != nil {
			t.Fatal(err)
		}

		query, args, err := qb.Select("name").
			From("users").
			Where("active = ?", true).
			Where("name IN (?)", inVals).
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
			t.Fatalf("expected 2 active users from IN list, got %d", len(names))
		}
	})
}

func TestInClauseDelete(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, _ = td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id IN (SELECT id FROM users WHERE name IN ('Alice', 'Bob', 'Charlie'))")

		qb := td.builder()
		inVals, err := xdal.In("Alice", "Bob")
		if err != nil {
			t.Fatal(err)
		}

		query, args, err := qb.Delete("users").
			Where("name IN (?)", inVals).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		result, err := td.dalDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 2 {
			t.Errorf("expected 2 rows affected, got %d", rows)
		}
	})
}
