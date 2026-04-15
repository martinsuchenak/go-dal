package sqlite

import (
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

func TestNewQueryBuilderUsesQuestionMark(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := `INSERT INTO "users" ("name") VALUES (?)`
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
	expected := `SELECT "id" FROM "users" WHERE id = ?`
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("got args %v, want [1]", args)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ dal.DBInterface = (*SQLiteDB)(nil)
}

func TestExpressionOverrides(t *testing.T) {
	d := NewDialect()
	if got := d.CurrentTimestamp(); got != "datetime('now')" {
		t.Errorf("CurrentTimestamp = %q, want datetime('now')", got)
	}
	if got := d.BoolLiteral(true); got != "1" {
		t.Errorf("BoolLiteral(true) = %q, want 1", got)
	}
	if got := d.BoolLiteral(false); got != "0" {
		t.Errorf("BoolLiteral(false) = %q, want 0", got)
	}
	if got := d.StringAggExpr("name", "', '"); got != "GROUP_CONCAT(name, ', ')" {
		t.Errorf("StringAggExpr = %q, want GROUP_CONCAT(name, ', ')", got)
	}
	if got := d.RandExpr(); got != "RANDOM()" {
		t.Errorf("RandExpr = %q, want RANDOM()", got)
	}
}

func TestTranslateSQL(t *testing.T) {
	d := NewDialect()
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = ? WHERE id = ?"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
