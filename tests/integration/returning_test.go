package integration

import (
	"context"
	"fmt"
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

func TestUpdateReturning(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			t.Skip("MySQL doesn't support RETURNING")
		}
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Update("users").
			Set("name", "Updated").
			Where("id = ?", 1).
			Returning("id", "name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var id int
		var name string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&id, &name)
		if err != nil {
			t.Fatalf("update returning failed: %v", err)
		}
		if id != 1 {
			t.Errorf("got id %d, want 1", id)
		}
		if name != "Updated" {
			t.Errorf("got name %q, want 'Updated'", name)
		}
	})
}

func TestDeleteReturning(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		if td.name == "mysql" {
			t.Skip("MySQL doesn't support RETURNING")
		}
		ctx := context.Background()

		_, err := td.dalDB.Exec(ctx, "DELETE FROM orders WHERE user_id = 3")
		if err != nil {
			t.Fatal(err)
		}

		_, err = td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('ToDelete', 'del@test.com', %s)", td.dialect.BoolLiteral(false)))
		if err != nil {
			t.Fatal(err)
		}

		qb := td.builder()
		query, args, err := qb.Delete("users").
			Where("name = ?", "ToDelete").
			Returning("id", "name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var id int
		var name string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&id, &name)
		if err != nil {
			t.Fatalf("delete returning failed: %v", err)
		}
		if id <= 0 {
			t.Errorf("got id %d, want positive", id)
		}
		if name != "ToDelete" {
			t.Errorf("got name %q, want 'ToDelete'", name)
		}
	})
}
