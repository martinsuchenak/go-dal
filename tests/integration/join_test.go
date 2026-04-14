package integration

import (
	"context"
	"fmt"
	"testing"
)

func TestInnerJoin(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args := qb.Select("u.name", "p.name", "o.quantity").
			From("users u").
			Join("INNER JOIN orders o ON o.user_id = u.id").
			Join("INNER JOIN products p ON p.id = o.product_id").
			OrderBy("u.name", "p.name").
			Build()

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		type row struct {
			userName    string
			productName string
			quantity    int
		}
		var results []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.userName, &r.productName, &r.quantity); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}

		if len(results) != 4 {
			t.Fatalf("expected 4 rows, got %d", len(results))
		}

		// Alice: Widget(qty=2), Gadget(qty=1)
		if results[0].userName != "Alice" || results[0].productName != "Gadget" || results[0].quantity != 1 {
			t.Errorf("row 0: got %+v", results[0])
		}
		if results[1].userName != "Alice" || results[1].productName != "Widget" || results[1].quantity != 2 {
			t.Errorf("row 1: got %+v", results[1])
		}

		// Bob: Doohickey(qty=5), Gadget(qty=3)
		if results[2].userName != "Bob" || results[2].productName != "Doohickey" || results[2].quantity != 5 {
			t.Errorf("row 2: got %+v", results[2])
		}
		if results[3].userName != "Bob" || results[3].productName != "Gadget" || results[3].quantity != 3 {
			t.Errorf("row 3: got %+v", results[3])
		}
	})
}

func TestLeftJoin(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		// Charlie has no orders -- LEFT JOIN should still return Charlie with NULLs
		query, args := qb.Select("u.name", "o.total_price").
			From("users u").
			Join("LEFT JOIN orders o ON o.user_id = u.id").
			OrderBy("u.name", "o.total_price").
			Build()

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		type row struct {
			name       string
			totalPrice *float64
		}
		var results []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.name, &r.totalPrice); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			results = append(results, r)
		}

		// Alice(2 orders) + Bob(2 orders) + Charlie(0 orders = 1 NULL row) = 5 rows
		if len(results) != 5 {
			t.Fatalf("expected 5 rows, got %d", len(results))
		}

		// Charlie is last (ordered by name), should have NULL total_price
		charlie := results[4]
		if charlie.name != "Charlie" {
			t.Errorf("expected last row to be Charlie, got %q", charlie.name)
		}
		if charlie.totalPrice != nil {
			t.Errorf("expected Charlie's total_price to be NULL, got %v", charlie.totalPrice)
		}
	})
}

func TestJoinWithWhere(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args := qb.Select("u.name", "o.total_price").
			From("users u").
			Join("INNER JOIN orders o ON o.user_id = u.id").
			Where("u.name = ?", "Bob").
			OrderBy("o.total_price").
			Build()

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			var name string
			var price float64
			if err := rows.Scan(&name, &price); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			if name != "Bob" {
				t.Errorf("expected Bob, got %q", name)
			}
			count++
		}

		if count != 2 {
			t.Errorf("expected 2 orders for Bob, got %d", count)
		}
	})
}

func TestThreeTableJoin(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		activeVal := "1"
		if td.name == "postgres" {
			activeVal = "TRUE"
		}

		query := fmt.Sprintf(`SELECT u.name, p.name, o.quantity, o.total_price
			FROM users u
			INNER JOIN orders o ON o.user_id = u.id
			INNER JOIN products p ON p.id = o.product_id
			WHERE u.active = %s
			ORDER BY u.name, p.name`, activeVal)

		rows, err := td.dalDB.Query(ctx, query)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			var userName, productName string
			var quantity int
			var totalPrice float64
			if err := rows.Scan(&userName, &productName, &quantity, &totalPrice); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			count++
		}

		// Only active users (Alice + Bob) have orders = 4 rows
		if count != 4 {
			t.Errorf("expected 4 rows for active users, got %d", count)
		}
	})
}
