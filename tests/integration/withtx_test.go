package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

func TestWithTxCommit(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		err := td.dalDB.WithTx(ctx, nil, func(tx *dal.Tx) error {
			qb := td.dalDB.NewQueryBuilder()
			query, args, err := qb.Insert("users").
				Set("name", "WithTxUser").
				Set("email", "withtx@example.com").
				Set("active", true).
				Build()
			if err != nil {
				return err
			}

			result, err := tx.Exec(ctx, query, args...)
			if err != nil {
				return err
			}
			rows, _ := result.RowsAffected()
			if rows != 1 {
				return fmt.Errorf("expected 1 row affected, got %d", rows)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WithTx commit failed: %v", err)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "WithTxUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after WithTx commit failed: %v", err)
		}
		if name != "WithTxUser" {
			t.Errorf("got name %q, want 'WithTxUser'", name)
		}
	})
}

func TestWithTxRollbackOnError(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		err := td.dalDB.WithTx(ctx, nil, func(tx *dal.Tx) error {
			qb := td.dalDB.NewQueryBuilder()
			query, args, err := qb.Insert("users").
				Set("name", "RollbackUser").
				Set("email", "rollback@example.com").
				Set("active", true).
				Build()
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query, args...)
			if err != nil {
				return err
			}
			return fmt.Errorf("intentional rollback")
		})
		if err == nil {
			t.Fatal("expected error from WithTx")
		}

		var count int
		err = td.dalDB.Select("COUNT(*)").
			From("users").
			Where("name = ?", "RollbackUser").
			QueryRow(ctx).Scan(&count)
		if err != nil {
			t.Fatalf("count after WithTx rollback failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 rows after rollback, got %d", count)
		}
	})
}

func TestWithTxDirectExecution(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		err := td.dalDB.WithTx(ctx, nil, func(tx *dal.Tx) error {
			qb := td.dalDB.NewQueryBuilder()
			query, args, err := qb.Insert("users").
				Set("name", "TxDirUser").
				Set("email", "txdir@example.com").
				Set("active", true).
				Build()
			if err != nil {
				return err
			}

			result, err := tx.Exec(ctx, query, args...)
			if err != nil {
				return err
			}
			rows, _ := result.RowsAffected()
			if rows != 1 {
				return fmt.Errorf("expected 1 row, got %d", rows)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WithTx direct exec failed: %v", err)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "TxDirUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after WithTx direct exec failed: %v", err)
		}
		if name != "TxDirUser" {
			t.Errorf("got name %q, want 'TxDirUser'", name)
		}
	})
}

func TestTxQueryMultiRow(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		tx, err := td.dalDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		query, args, err := td.builder().Select("name").
			From("users").
			OrderBy("name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("tx query failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			names = append(names, name)
		}

		if len(names) != 3 {
			t.Errorf("expected 3 users, got %d: %v", len(names), names)
		}
		if names[0] != "Alice" {
			t.Errorf("got names[0] = %q, want 'Alice'", names[0])
		}
	})
}

func TestTxQueryWithParams(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		tx, err := td.dalDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		query, args, err := td.builder().Select("name").
			From("users").
			Where("active = ?", true).
			OrderBy("name").
			Build()
		if err != nil {
			t.Fatal(err)
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			t.Fatalf("tx query with params failed: %v", err)
		}
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			names = append(names, name)
		}

		if len(names) != 2 {
			t.Errorf("expected 2 active users, got %d: %v", len(names), names)
		}
	})
}

func TestWithTxUpdateAndVerify(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		err := td.dalDB.WithTx(ctx, nil, func(tx *dal.Tx) error {
			query, args, err := td.builder().Update("users").
				Set("email", "alice_wtx@example.com").
				Where("name = ?", "Alice").
				Build()
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query, args...)
			return err
		})
		if err != nil {
			t.Fatalf("WithTx update failed: %v", err)
		}

		var email string
		err = td.dalDB.Select("email").
			From("users").
			Where("name = ?", "Alice").
			QueryRow(ctx).Scan(&email)
		if err != nil {
			t.Fatalf("select after WithTx update failed: %v", err)
		}
		if email != "alice_wtx@example.com" {
			t.Errorf("got email %q, want 'alice_wtx@example.com'", email)
		}
	})
}
