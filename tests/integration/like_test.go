package integration

import (
	"context"
	"testing"
)

func TestLike(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("email LIKE ?", "%@example.com").
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
			var name string
			rows.Scan(&name)
			names = append(names, name)
		}

		if len(names) != 3 {
			t.Fatalf("expected 3 users with @example.com email, got %d", len(names))
		}
	})
}

func TestLikeStartsWith(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("name LIKE ?", "A%").
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

		if len(names) != 1 {
			t.Fatalf("expected 1 user starting with A, got %d", len(names))
		}
		if names[0] != "Alice" {
			t.Errorf("got %q, want 'Alice'", names[0])
		}
	})
}

func TestLikeContains(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("name LIKE ?", "%ob%").
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

		if len(names) != 1 {
			t.Fatalf("expected 1 user with 'ob' in name, got %d", len(names))
		}
		if names[0] != "Bob" {
			t.Errorf("got %q, want 'Bob'", names[0])
		}
	})
}

func TestLikeCombinedWithInClause(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("email LIKE ?", "%@example.com").
			Where("name LIKE ?", "A%").
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

		if len(names) != 1 {
			t.Fatalf("expected 1 user matching both LIKEs, got %d", len(names))
		}
		if names[0] != "Alice" {
			t.Errorf("got %q, want 'Alice'", names[0])
		}
	})
}
