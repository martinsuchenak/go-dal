package dal

import (
	"testing"
)

func defaultDialect() Dialect {
	return &BaseDialect{PlaceholderStyle: QuestionMark, LimitStyle: LimitOffsetStyle}
}

func dollarDialect() Dialect {
	return &BaseDialect{PlaceholderStyle: DollarNumber, LimitStyle: LimitOffsetStyle}
}

func atPDialect() Dialect {
	return &BaseDialect{PlaceholderStyle: AtPNumber, LimitStyle: FetchNextStyle}
}

func TestSelectBasic(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("id", "name").
		From("users").
		Where("age > ?", 18).
		OrderBy("name").
		Limit(10).
		Build()

	expected := "SELECT id, name FROM users WHERE age > ? ORDER BY name LIMIT 10"
	assertQuery(t, query, expected)
	assertArgs(t, args, 18)
}

func TestSelectStar(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select().From("users").Build()

	assertQuery(t, query, "SELECT * FROM users")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestSelectMultipleWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("id", "name").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertQuery(t, query, "SELECT id, name FROM users WHERE age > ? AND active = ?")
	assertArgs(t, args, 18, true)
}

func TestSelectOffset(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _ := qb.Select("id").From("users").Limit(10).Offset(20).Build()
	assertQuery(t, query, "SELECT id FROM users LIMIT 10 OFFSET 20")
}

func TestInsert(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Insert("users").
		Set("name", "John Doe").
		Set("email", "john@example.com").
		Build()

	assertQuery(t, query, "INSERT INTO users (name, email) VALUES (?, ?)")
	if len(args) != 2 {
		t.Fatalf("got %d args, want 2", len(args))
	}
	if args[0] != "John Doe" {
		t.Errorf("got args[0] = %v, want 'John Doe'", args[0])
	}
	if args[1] != "john@example.com" {
		t.Errorf("got args[1] = %v, want 'john@example.com'", args[1])
	}
}

func TestInsertEmpty(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Insert("users").Build()
	assertQuery(t, query, "")
	if args != nil {
		t.Errorf("got args %v, want nil", args)
	}
}

func TestUpdate(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Update("users").
		Set("email", "new@example.com").
		Where("id = ?", 123).
		Build()

	assertQuery(t, query, "UPDATE users SET email = ? WHERE id = ?")
	assertArgs(t, args, "new@example.com", 123)
}

func TestUpdateMultipleSet(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Update("users").
		Set("name", "Jane").
		Set("email", "jane@example.com").
		Where("id = ?", 1).
		Build()

	assertQuery(t, query, "UPDATE users SET name = ?, email = ? WHERE id = ?")
	if len(args) != 3 {
		t.Fatalf("got %d args, want 3", len(args))
	}
}

func TestUpdateEmpty(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Update("users").Build()
	assertQuery(t, query, "")
	if args != nil {
		t.Errorf("got args %v, want nil", args)
	}
}

func TestDelete(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Delete("users").Where("id = ?", 123).Build()
	assertQuery(t, query, "DELETE FROM users WHERE id = ?")
	assertArgs(t, args, 123)
}

func TestDeleteAll(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Delete("users").Build()
	assertQuery(t, query, "DELETE FROM users")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestDeleteMultipleWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Delete("users").
		Where("active = ?", false).
		Where("created_at < ?", "2020-01-01").
		Build()

	assertQuery(t, query, "DELETE FROM users WHERE active = ? AND created_at < ?")
	if len(args) != 2 {
		t.Errorf("got %d args, want 2", len(args))
	}
}

func TestDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())

	query, args := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE age > $1 AND active = $2")
	assertArgs(t, args, 18, true)
}

func TestDollarInsert(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Insert("users").
		Set("name", "John").
		Set("email", "john@example.com").
		Build()

	assertQuery(t, query, "INSERT INTO users (name, email) VALUES ($1, $2)")
	assertArgs(t, args, "John", "john@example.com")
}

func TestDollarUpdate(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Update("users").
		Set("name", "Jane").
		Where("id = ?", 1).
		Build()

	assertQuery(t, query, "UPDATE users SET name = $1 WHERE id = $2")
	assertArgs(t, args, "Jane", 1)
}

func TestDollarDelete(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Delete("users").
		Where("id = ?", 1).
		Build()

	assertQuery(t, query, "DELETE FROM users WHERE id = $1")
	assertArgs(t, args, 1)
}

func TestAtPPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())

	query, args := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE age > @p1 AND active = @p2")
	assertArgs(t, args, 18, true)
}

func TestAtPInsert(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args := qb.Insert("users").
		Set("name", "John").
		Set("email", "john@example.com").
		Build()

	assertQuery(t, query, "INSERT INTO users (name, email) VALUES (@p1, @p2)")
	assertArgs(t, args, "John", "john@example.com")
}

func TestAtPUpdate(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args := qb.Update("users").
		Set("name", "Jane").
		Set("email", "jane@ex.com").
		Where("id = ?", 1).
		Build()

	assertQuery(t, query, "UPDATE users SET name = @p1, email = @p2 WHERE id = @p3")
	if len(args) != 3 {
		t.Fatalf("got %d args, want 3", len(args))
	}
}

func TestAtPDelete(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args := qb.Delete("users").Where("id = ?", 1).Build()
	assertQuery(t, query, "DELETE FROM users WHERE id = @p1")
	assertArgs(t, args, 1)
}

func TestDollarWhereSkipsQuotedQuestionMarks(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Select("id").
		From("users").
		Where("name = '?' AND id = ?", 42).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE name = '?' AND id = $1")
	assertArgs(t, args, 42)
}

func TestDollarWhereDoubleQuoted(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Select("id").
		From("users").
		Where("col = \"?\" AND val = ?", 99).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE col = \"?\" AND val = $1")
	assertArgs(t, args, 99)
}

func TestDollarWhereEscapedQuotes(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Select("id").
		From("users").
		Where("name = 'it''s ?' AND val = ?", 7).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE name = 'it''s ?' AND val = $1")
	assertArgs(t, args, 7)
}

func TestDollarWhereMultipleParamsAcrossClauses(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("name LIKE '?' AND active = ?", true).
		Build()

	assertQuery(t, query, "SELECT id FROM users WHERE age > $1 AND name LIKE '?' AND active = $2")
	assertArgs(t, args, 18, true)
}

func TestDollarUpdateWithQuotedPlaceholder(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Update("users").
		Set("name", "Jane").
		Where("col = '?' AND id = ?", 1).
		Build()

	assertQuery(t, query, "UPDATE users SET name = $1 WHERE col = '?' AND id = $2")
	if len(args) != 2 {
		t.Fatalf("got %d args, want 2", len(args))
	}
}

func TestDollarDeleteWithQuotedPlaceholder(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Delete("users").
		Where("col = '?' AND id = ?", 1).
		Build()

	assertQuery(t, query, "DELETE FROM users WHERE col = '?' AND id = $1")
	assertArgs(t, args, 1)
}

func TestSelectJoin(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("u.id", "u.name", "o.total").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		Build()

	assertQuery(t, query, "SELECT u.id, u.name, o.total FROM users u INNER JOIN orders o ON o.user_id = u.id WHERE u.active = ?")
	assertArgs(t, args, true)
}

func TestSelectMultipleJoins(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _ := qb.Select("u.name", "o.total", "p.name").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Join("INNER JOIN products p ON p.id = o.product_id").
		Build()

	assertQuery(t, query, "SELECT u.name, o.total, p.name FROM users u INNER JOIN orders o ON o.user_id = u.id INNER JOIN products p ON p.id = o.product_id")
}

func TestSelectGroupBy(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("u.id", "COUNT(o.id) as order_count").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		GroupBy("u.id").
		Build()

	assertQuery(t, query, "SELECT u.id, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON o.user_id = u.id GROUP BY u.id")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestSelectGroupByHaving(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("u.id", "COUNT(o.id) as order_count").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		GroupBy("u.id").
		Having("COUNT(o.id) > ?", 2).
		Build()

	assertQuery(t, query, "SELECT u.id, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON o.user_id = u.id GROUP BY u.id HAVING COUNT(o.id) > ?")
	assertArgs(t, args, 2)
}

func TestSelectJoinWhereGroupByHavingOrderByLimit(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args := qb.Select("u.name", "SUM(o.amount) as total_spent").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		GroupBy("u.name").
		Having("SUM(o.amount) > ?", 100).
		OrderBy("total_spent DESC").
		Limit(10).
		Build()

	assertQuery(t, query, "SELECT u.name, SUM(o.amount) as total_spent FROM users u INNER JOIN orders o ON o.user_id = u.id WHERE u.active = ? GROUP BY u.name HAVING SUM(o.amount) > ? ORDER BY total_spent DESC LIMIT 10")
	assertArgs(t, args, true, 100)
}

func TestSelectGroupByHavingDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args := qb.Select("u.id", "COUNT(o.id)").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		GroupBy("u.id").
		Having("COUNT(o.id) > ?", 5).
		Build()

	assertQuery(t, query, "SELECT u.id, COUNT(o.id) FROM users u LEFT JOIN orders o ON o.user_id = u.id WHERE u.active = $1 GROUP BY u.id HAVING COUNT(o.id) > $2")
	assertArgs(t, args, true, 5)
}

func TestMSSQLLimitOffsetWithOrderBy(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _ := qb.Select("id", "name").
		From("users").
		OrderBy("name").
		Limit(10).
		Offset(20).
		Build()

	assertQuery(t, query, "SELECT id, name FROM users ORDER BY name OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY")
}

func TestMSSQLLimitOnly(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _ := qb.Select("id").
		From("users").
		OrderBy("id").
		Limit(5).
		Build()

	assertQuery(t, query, "SELECT id FROM users ORDER BY id OFFSET 0 ROWS FETCH NEXT 5 ROWS ONLY")
}

func TestMSSQLOffsetOnly(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _ := qb.Select("id").
		From("users").
		OrderBy("id").
		Offset(10).
		Build()

	assertQuery(t, query, "SELECT id FROM users ORDER BY id OFFSET 10 ROWS")
}

func TestMSSQLLimitOffsetWithoutOrderBy(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _ := qb.Select("id").
		From("users").
		Limit(10).
		Offset(5).
		Build()

	assertQuery(t, query, "SELECT id FROM users ORDER BY (SELECT NULL) OFFSET 5 ROWS FETCH NEXT 10 ROWS ONLY")
}

func assertQuery(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertArgs(t *testing.T, got []interface{}, want ...interface{}) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got args %v (len %d), want %v (len %d)", got, len(got), want, len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got args[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
