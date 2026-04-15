package integration

import (
	"context"
	"testing"
)

func TestConcat(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		concatExpr := qb.ConcatExpr("name", "' <'", "email", "'>'")

		query, args, err := qb.Select(concatExpr).
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var result string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			t.Fatalf("concat query failed: %v", err)
		}
		if result != "Alice <alice@example.com>" {
			t.Errorf("got %q, want 'Alice <alice@example.com>'", result)
		}
	})
}

func TestStringLength(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select(qb.LengthExpr("name")).
			From("users").
			Where("name = ?", "Bob").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var length int
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&length)
		if err != nil {
			t.Fatalf("length query failed: %v", err)
		}
		if length != 3 {
			t.Errorf("got length %d, want 3", length)
		}
	})
}

func TestCoalesce(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("COALESCE(NULL, name, 'default')").
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var result string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			t.Fatalf("coalesce query failed: %v", err)
		}
		if result != "Alice" {
			t.Errorf("got %q, want 'Alice'", result)
		}
	})
}

func TestUpperLower(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("UPPER(name)", "LOWER(email)").
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var upper, lower string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&upper, &lower)
		if err != nil {
			t.Fatalf("upper/lower query failed: %v", err)
		}
		if upper != "ALICE" {
			t.Errorf("got upper %q, want 'ALICE'", upper)
		}
		if lower != "alice@example.com" {
			t.Errorf("got lower %q, want 'alice@example.com'", lower)
		}
	})
}

func TestCurrentTimestamp(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select(qb.CurrentTimestamp()).
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var result interface{}
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			t.Fatalf("current timestamp query failed: %v", err)
		}
		if result == nil {
			t.Error("got nil timestamp")
		}
	})
}

func TestBoolLiteral(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		raw := "SELECT " + td.dialect.BoolLiteral(true)
		var result interface{}
		err := td.dalDB.QueryRow(ctx, raw).Scan(&result)
		if err != nil {
			t.Fatalf("bool literal query failed: %v", err)
		}
	})
}

func TestStringAgg(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		aggExpr := qb.StringAggExpr("name", "', '")

		query, args, err := qb.Select(aggExpr).
			From("users").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var result *string
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			t.Fatalf("string agg query failed: %v", err)
		}
		if result == nil {
			t.Error("got nil string agg result")
		}
	})
}

func TestRandExpr(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select(qb.RandExpr()).
			From("users").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var result interface{}
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&result)
		if err != nil {
			t.Fatalf("rand query failed: %v", err)
		}
		if result == nil {
			t.Error("got nil random result")
		}
	})
}
