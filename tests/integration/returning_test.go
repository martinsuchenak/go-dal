package integration

import (
	"context"
	"testing"
)

func TestInsertReturning(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		switch td.name {
		case "mysql":
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

			id, err := result.LastInsertId()
			if err != nil {
				t.Fatalf("last insert id failed: %v", err)
			}
			if id <= 0 {
				t.Errorf("expected positive id, got %d", id)
			}

		default:
			query, args, err := qb.Insert("users").
				Set("name", "Dave").
				Set("email", "dave@example.com").
				Set("active", true).
				Returning("id").
				Build()
			if err != nil {
				t.Fatal(err)
			}

			var id int
			err = td.dalDB.QueryRow(ctx, query, args...).Scan(&id)
			if err != nil {
				t.Fatalf("insert returning failed: %v", err)
			}
			if id <= 0 {
				t.Errorf("expected positive id, got %d", id)
			}
		}
	})
}

func TestInsertReturningMultipleColumns(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			t.Skip("MySQL doesn't support RETURNING")
		}
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Insert("users").
			Set("name", "Eve").
			Set("email", "eve@example.com").
			Set("active", true).
			Returning("id", "name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var id int
		var name string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&id, &name)
		if err != nil {
			t.Fatalf("insert returning multiple columns failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive id, got %d", id)
		}
		if name != "Eve" {
			t.Errorf("got name %q, want 'Eve'", name)
		}
	})
}
