package integration

import (
	"context"
	"testing"
)

func TestTransactionCommit(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		tx, err := td.dalDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}

		qb := td.builder()
		query, args, err := qb.Insert("users").
			Set("name", "TxUser").
			Set("email", "tx@example.com").
			Set("active", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("insert in tx failed: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("commit failed: %v", err)
		}

		qb = td.builder()
		sq, sa, err := qb.Select("name").From("users").Where("name = ?", "TxUser").Build()
		if err != nil {
			t.Fatal(err)
		}
		var name string
		err = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&name)
		if err != nil {
			t.Fatalf("select after commit failed: %v", err)
		}
		if name != "TxUser" {
			t.Errorf("got name %q, want 'TxUser'", name)
		}
	})
}

func TestTransactionRollback(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		tx, err := td.dalDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}

		qb := td.builder()
		query, args, err := qb.Insert("users").
			Set("name", "RollbackUser").
			Set("email", "rollback@example.com").
			Set("active", true).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("insert in tx failed: %v", err)
		}

		if err := tx.Rollback(); err != nil {
			t.Fatalf("rollback failed: %v", err)
		}

		qb = td.builder()
		sq, sa, err := qb.Select("COUNT(*)").From("users").Where("name = ?", "RollbackUser").Build()
		if err != nil {
			t.Fatal(err)
		}
		var count int
		td.dalDB.QueryRow(ctx, sq, sa...).Scan(&count)
		if count != 0 {
			t.Errorf("expected 0 rows after rollback, got %d", count)
		}
	})
}

func TestTransactionWithUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		tx, err := td.dalDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}

		// Update Alice's email in transaction
		qb := td.builder()
		query, args, err := qb.Update("users").
			Set("email", "alice_tx@example.com").
			Where("name = ?", "Alice").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("update in tx failed: %v", err)
		}

		// Read within same transaction -- should see new value
		var email string
		qb = td.builder()
		sq, sa, err := qb.Select("email").From("users").Where("name = ?", "Alice").Build()
		if err != nil {
			t.Fatal(err)
		}
		err = tx.QueryRow(ctx, sq, sa...).Scan(&email)
		if err != nil {
			t.Fatalf("select in tx failed: %v", err)
		}
		if email != "alice_tx@example.com" {
			t.Errorf("in-tx read: got %q, want 'alice_tx@example.com'", email)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("commit failed: %v", err)
		}

		// Read after commit -- should see new value
		qb = td.builder()
		sq, sa, err = qb.Select("email").From("users").Where("name = ?", "Alice").Build()
		if err != nil {
			t.Fatal(err)
		}
		err = td.dalDB.QueryRow(ctx, sq, sa...).Scan(&email)
		if err != nil {
			t.Fatalf("select after commit failed: %v", err)
		}
		if email != "alice_tx@example.com" {
			t.Errorf("post-commit read: got %q, want 'alice_tx@example.com'", email)
		}
	})
}
