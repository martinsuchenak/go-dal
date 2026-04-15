package mysql

import (
	"database/sql"
	"testing"

	"github.com/martinsuchenak/xdal/pkg/xdal"

	_ "modernc.org/sqlite"
)

func TestNewQueryBuilderUsesQuestionMark(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := "INSERT INTO `users` (`name`) VALUES (?)"
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
	expected := "SELECT `id` FROM `users` WHERE id = ?"
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("got args %v, want [1]", args)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ xdal.DBInterface = (*MySQLDB)(nil)
}

func TestTranslateSQL(t *testing.T) {
	d := NewDialect()
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = ? WHERE id = ?"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpressionDefaults(t *testing.T) {
	d := NewDialect()
	if got := d.ConcatExpr("a", "b"); got != "CONCAT(a, b)" {
		t.Errorf("ConcatExpr = %q, want CONCAT(a, b)", got)
	}
	if got := d.LengthExpr("name"); got != "LENGTH(name)" {
		t.Errorf("LengthExpr = %q, want LENGTH(name)", got)
	}
	if got := d.CurrentTimestamp(); got != "NOW()" {
		t.Errorf("CurrentTimestamp = %q, want NOW()", got)
	}
	if got := d.BoolLiteral(true); got != "TRUE" {
		t.Errorf("BoolLiteral(true) = %q, want TRUE", got)
	}
	if got := d.StringAggExpr("name", "', '"); got != "GROUP_CONCAT(name SEPARATOR ', ')" {
		t.Errorf("StringAggExpr = %q, want GROUP_CONCAT(name SEPARATOR ', ')", got)
	}
	if got := d.RandExpr(); got != "RAND()" {
		t.Errorf("RandExpr = %q, want RAND()", got)
	}
}

func TestNewMySQLDB(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()

	mdb := NewMySQLDB(sqlDB, nil)
	if mdb == nil {
		t.Fatal("expected non-nil MySQLDB")
	}
	if mdb.DB() != sqlDB {
		t.Error("DB() should return the underlying *sql.DB")
	}
	if mdb.Dialect() == nil {
		t.Error("Dialect() should return non-nil")
	}

	qb := mdb.NewQueryBuilder()
	if qb == nil {
		t.Fatal("expected non-nil QueryBuilder")
	}
}
