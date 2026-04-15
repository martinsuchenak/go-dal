package mssql

import (
	"database/sql"
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"

	_ "modernc.org/sqlite"
)

func TestNewQueryBuilderUsesAtP(t *testing.T) {
	qb := NewQueryBuilder()
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Build()

	if err != nil {
		t.Fatal(err)
	}
	expected := "INSERT INTO [users] ([name]) VALUES (@p1)"
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
	expected := "SELECT [id] FROM [users] WHERE id = @p1"
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
	expected := "UPDATE [users] SET [name] = @p1 WHERE id = @p2"
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 2 {
		t.Errorf("got %d args, want 2", len(args))
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var _ dal.DBInterface = (*MSSQLDB)(nil)
}

func TestExpressionOverrides(t *testing.T) {
	d := NewDialect()
	if got := d.ConcatExpr("a", "b", "c"); got != "a + b + c" {
		t.Errorf("ConcatExpr = %q, want a + b + c", got)
	}
	if got := d.LengthExpr("name"); got != "LEN(name)" {
		t.Errorf("LengthExpr = %q, want LEN(name)", got)
	}
	if got := d.CurrentTimestamp(); got != "GETDATE()" {
		t.Errorf("CurrentTimestamp = %q, want GETDATE()", got)
	}
	if got := d.BoolLiteral(true); got != "1" {
		t.Errorf("BoolLiteral(true) = %q, want 1", got)
	}
	if got := d.BoolLiteral(false); got != "0" {
		t.Errorf("BoolLiteral(false) = %q, want 0", got)
	}
	if got := d.StringAggExpr("name", "', '"); got != "STRING_AGG(name, ', ')" {
		t.Errorf("StringAggExpr = %q, want STRING_AGG(name, ', ')", got)
	}
	if got := d.RandExpr(); got != "RAND()" {
		t.Errorf("RandExpr = %q, want RAND()", got)
	}
}

func TestTranslateSQL(t *testing.T) {
	d := NewDialect()
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = @p1 WHERE id = @p2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewMSSQLDB(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()

	mdb := NewMSSQLDB(sqlDB, nil)
	if mdb == nil {
		t.Fatal("expected non-nil MSSQLDB")
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
