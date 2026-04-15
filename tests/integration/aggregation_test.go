package integration

import (
	"context"
	"testing"
)

func TestCountAggregate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("COUNT(*)").
			From("users").
			Where("active = ?", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		var count int
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&count)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 active users, got %d", count)
		}
	})
}

func TestSumAggregate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		var query string
		var args []interface{}
		var err error
		if td.name == "postgres" {
			query, args, err = qb.Select("SUM(o.total_price)").
				From("orders o").
				Join("INNER JOIN users u ON u.id = o.user_id").
				Where("u.name = ?", "Alice").
				Build()
		} else {
			query, args, err = qb.Select("SUM(total_price)").
				From("orders").
				Where("user_id = (SELECT id FROM users WHERE name = ?)", "Alice").
				Build()
		}
		if err != nil {
			t.Fatal(err)
		}

		var total *float64
		err = td.dalDB.QueryRow(ctx, query, args...).Scan(&total)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if total == nil {
			t.Fatal("expected non-NULL total, got NULL")
		}
		expectedTotal := 44.97
		if diff := *total - expectedTotal; diff < -0.01 || diff > 0.01 {
			t.Errorf("expected total ~%.2f, got %.2f", expectedTotal, *total)
		}
	})
}

func TestGroupBy(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("u.name", "COUNT(o.id) as order_count").
			From("users u").
			Join("LEFT JOIN orders o ON o.user_id = u.id").
			GroupBy("u.name").
			OrderBy("u.name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		type row struct {
			name       string
			orderCount int
		}
		var results []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.name, &r.orderCount); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 rows, got %d", len(results))
		}

		if results[0].name != "Alice" || results[0].orderCount != 2 {
			t.Errorf("Alice: expected 2 orders, got name=%q count=%d", results[0].name, results[0].orderCount)
		}
		if results[1].name != "Bob" || results[1].orderCount != 2 {
			t.Errorf("Bob: expected 2 orders, got name=%q count=%d", results[1].name, results[1].orderCount)
		}
		if results[2].name != "Charlie" || results[2].orderCount != 0 {
			t.Errorf("Charlie: expected 0 orders, got name=%q count=%d", results[2].name, results[2].orderCount)
		}
	})
}

func TestGroupByWithHaving(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("u.name", "SUM(o.total_price) as total_spent").
			From("users u").
			Join("INNER JOIN orders o ON o.user_id = u.id").
			GroupBy("u.name").
			Having("SUM(o.total_price) > ?", 50).
			OrderBy("total_spent DESC").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		type row struct {
			name       string
			totalSpent float64
		}
		var results []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.name, &r.totalSpent); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}

		// Bob: 74.97 + 24.95 = 99.92 (> 50)
		// Alice: 19.98 + 24.99 = 44.97 (NOT > 50)
		if len(results) != 1 {
			t.Fatalf("expected 1 row (Bob only), got %d", len(results))
		}
		if results[0].name != "Bob" {
			t.Errorf("expected Bob, got %q", results[0].name)
		}
		expectedTotal := 99.92
		if diff := results[0].totalSpent - expectedTotal; diff < -0.01 || diff > 0.01 {
			t.Errorf("expected total ~%.2f, got %.2f", expectedTotal, results[0].totalSpent)
		}
	})
}

func TestAvgAggregate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		query := "SELECT AVG(price) FROM products"
		var avg float64
		err := td.dalDB.QueryRow(ctx, query).Scan(&avg)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		// (9.99 + 24.99 + 4.99) / 3 = 13.323...
		expectedAvg := 13.323333
		if diff := avg - expectedAvg; diff < -0.01 || diff > 0.01 {
			t.Errorf("expected avg ~%.2f, got %.2f", expectedAvg, avg)
		}
	})
}

func TestMinMaxAggregate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		var minPrice, maxPrice float64
		err := td.dalDB.QueryRow(ctx, "SELECT MIN(price), MAX(price) FROM products").Scan(&minPrice, &maxPrice)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if minPrice != 4.99 {
			t.Errorf("expected min 4.99, got %.2f", minPrice)
		}
		if maxPrice != 24.99 {
			t.Errorf("expected max 24.99, got %.2f", maxPrice)
		}
	})
}

func TestGroupByMultipleColumns(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("u.name", "p.name as product", "SUM(o.quantity) as total_qty").
			From("users u").
			Join("INNER JOIN orders o ON o.user_id = u.id").
			Join("INNER JOIN products p ON p.id = o.product_id").
			GroupBy("u.name", "p.name").
			OrderBy("u.name", "p.name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		type row struct {
			user     string
			product  string
			totalQty int
		}
		var results []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.user, &r.product, &r.totalQty); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}

		// Alice: Gadget(1), Widget(2)
		// Bob: Doohickey(5), Gadget(3)
		if len(results) != 4 {
			t.Fatalf("expected 4 rows, got %d", len(results))
		}

		if results[0].user != "Alice" || results[0].product != "Gadget" || results[0].totalQty != 1 {
			t.Errorf("row 0: got %+v", results[0])
		}
		if results[1].user != "Alice" || results[1].product != "Widget" || results[1].totalQty != 2 {
			t.Errorf("row 1: got %+v", results[1])
		}
		if results[2].user != "Bob" || results[2].product != "Doohickey" || results[2].totalQty != 5 {
			t.Errorf("row 2: got %+v", results[2])
		}
		if results[3].user != "Bob" || results[3].product != "Gadget" || results[3].totalQty != 3 {
			t.Errorf("row 3: got %+v", results[3])
		}
	})
}
