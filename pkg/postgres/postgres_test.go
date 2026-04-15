package postgres

import (
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

func TestNewQueryBuilderUsesDollar(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := `INSERT INTO "users" ("name") VALUES ($1)`
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 1 || args[0] != "John" {
		t.Errorf("got args %v, want [John]", args)
	}
}

func TestNewQueryBuilderSelectWhere(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Select("id").
		From("users").
		Where("id = ?", 1).
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := `SELECT "id" FROM "users" WHERE id = $1`
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("got args %v, want [1]", args)
	}
}

func TestNewQueryBuilderUpdate(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Update("users").
		Set("name", "Jane").
		Where("id = ?", 1).
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := `UPDATE "users" SET "name" = $1 WHERE id = $2`
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 2 {
		t.Errorf("got %d args, want 2", len(args))
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ dal.DBInterface = (*PostgresDB)(nil)
}

func TestExpressionOverrides(t *testing.T) {
	d := NewDialect()
	if got := d.ConcatExpr("a", "b"); got != "CONCAT(a, b)" {
		t.Errorf("ConcatExpr = %q, want CONCAT(a, b)", got)
	}
	if got := d.StringAggExpr("name", "', '"); got != "STRING_AGG(name, ', ')" {
		t.Errorf("StringAggExpr = %q, want STRING_AGG(name, ', ')", got)
	}
	if got := d.RandExpr(); got != "RANDOM()" {
		t.Errorf("RandExpr = %q, want RANDOM()", got)
	}
}

func TestTranslateSQL(t *testing.T) {
	d := NewDialect()
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ? AND name = ?")
	want := "UPDATE users SET x = $1 WHERE id = $2 AND name = $3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
