package mssql

import (
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"
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
