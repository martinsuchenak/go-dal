package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

func TestWhereGroupIntegration(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			Where("active = ?", true).
			WhereGroup(func(g *xdal.WhereGroup) {
				g.Where("name = ?", "Alice").OrWhere("name = ?", "Bob")
			}).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			names = append(names, name)
		}
		if len(names) != 2 {
			t.Errorf("got %d rows, want 2: %v", len(names), names)
		}
	})
}

func TestWhereBetweenIntegration(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("name").
			From("products").
			WhereBetween("price", 4.0, 15.0).
			OrderBy("price").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			names = append(names, name)
		}
		if len(names) != 2 {
			t.Errorf("got %d products, want 2 (Doohickey + Widget): %v", len(names), names)
		}
	})
}

func TestWhereIsNotNullIntegration(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		_, err := td.dalDB.Exec(ctx, fmt.Sprintf("INSERT INTO users (name, email, active) VALUES ('Alice', 'alice@test.com', %s)", td.dialect.BoolLiteral(true)))
		if err != nil {
			t.Fatal(err)
		}

		qb := td.builder()

		query, args, err := qb.Select("name").
			From("users").
			WhereIsNotNull("email").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 1 {
			t.Errorf("expected 1 row with non-null email, got %d", count)
		}
	})
}

func TestOrHavingIntegration(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("user_id", "SUM(total_price) as total").
			From("orders").
			GroupBy("user_id").
			Having("SUM(total_price) > ?", 50).
			OrHaving("COUNT(*) > ?", 2).
			OrderBy("user_id").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		count := 0
		for rows.Next() {
			count++
		}
		if count == 0 {
			t.Error("expected at least one group matching HAVING clause")
		}
	})
}
