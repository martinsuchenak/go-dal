package integration

import (
	"context"
	"testing"
)

func TestSelectDistinct(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("active").
			Distinct().
			From("users").
			OrderBy("active").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var values []int
		for rows.Next() {
			var active interface{}
			_ = rows.Scan(&active)
			values = append(values, 1)
		}

		if len(values) != 2 {
			t.Errorf("expected 2 distinct active values, got %d", len(values))
		}
	})
}

func TestSelectDistinctCount(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.Select("user_id").
			Distinct().
			From("orders").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var count int
		for rows.Next() {
			var userID int
			_ = rows.Scan(&userID)
			count++
		}

		if count != 2 {
			t.Errorf("expected 2 distinct user_ids in orders, got %d", count)
		}
	})
}

func TestSelectAll(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()
		qb := td.builder()

		query, args, err := qb.SelectAll().
			From("users").
			Where("active = ?", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := td.dalDB.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		count := 0
		for rows.Next() {
			count++
		}

		if count != 2 {
			t.Errorf("expected 2 active users, got %d", count)
		}
	})
}
