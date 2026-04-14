package sqlite

import (
	"testing"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

func TestNewQueryBuilderUsesQuestionMark(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Insert("users").
		Set("name", "John").
		Build()

	expected := "INSERT INTO users (name) VALUES (?)"
	if query != expected {
		t.Errorf("got %q, want %q", query, expected)
	}
	if len(args) != 1 || args[0] != "John" {
		t.Errorf("got args %v, want [John]", args)
	}
}

func TestNewQueryBuilderSelectWhere(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.Select("id").
		From("users").
		Where("id = ?", 1).
		Build()

	expected := "SELECT id FROM users WHERE id = ?"
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
