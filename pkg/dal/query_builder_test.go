package dal

import (
	"testing"
)

func defaultDialect() Dialect {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder, AppendLimit: LimitOffset, BackslashEscapes: true}
	d.AppendReturning = d.WriteReturning
	return d
}

func dollarDialect() Dialect {
	d := &BaseDialect{Placeholder: DollarPlaceholder, AppendLimit: LimitOffset}
	d.AppendReturning = d.WriteReturning
	return d
}

func atPDialect() Dialect {
	d := &BaseDialect{Placeholder: AtPPlaceholder, AppendLimit: FetchNextLimit}
	d.PrependReturning = d.WriteOutput
	d.AppendReturning = d.WriteOutput
	return d
}

func TestSelectBasic(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id", "name").
		From("users").
		Where("age > ?", 18).
		OrderBy("name").
		Limit(10).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id, name FROM users WHERE age > ? ORDER BY name LIMIT 10")
	assertArgs(t, args, 18)
}

func TestSelectStar(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select().From("users").Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT * FROM users")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestSelectMultipleWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id", "name").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id, name FROM users WHERE age > ? AND active = ?")
	assertArgs(t, args, 18, true)
}

func TestSelectOffset(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _, err := qb.Select("id").From("users").Limit(10).Offset(20).Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users LIMIT 10 OFFSET 20")
}

func TestInsert(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").
		Set("name", "John Doe").
		Set("email", "john@example.com").
		Build()

	assertNoError(t, err)
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
	_, _, err := qb.Insert("users").Build()
	if err != ErrEmptyColumns {
		t.Errorf("got err %v, want ErrEmptyColumns", err)
	}
}

func TestUpdate(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Update("users").
		Set("email", "new@example.com").
		Where("id = ?", 123).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET email = ? WHERE id = ?")
	assertArgs(t, args, "new@example.com", 123)
}

func TestUpdateMultipleSet(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Update("users").
		Set("name", "Jane").
		Set("email", "jane@example.com").
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET name = ?, email = ? WHERE id = ?")
	if len(args) != 3 {
		t.Fatalf("got %d args, want 3", len(args))
	}
}

func TestUpdateEmpty(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Update("users").Build()
	if err != ErrEmptyColumns {
		t.Errorf("got err %v, want ErrEmptyColumns", err)
	}
}

func TestDelete(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Delete("users").Where("id = ?", 123).Build()
	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM users WHERE id = ?")
	assertArgs(t, args, 123)
}

func TestDeleteAll(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Delete("users").Build()
	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM users")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestDeleteMultipleWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Delete("users").
		Where("active = ?", false).
		Where("created_at < ?", "2020-01-01").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM users WHERE active = ? AND created_at < ?")
	if len(args) != 2 {
		t.Errorf("got %d args, want 2", len(args))
	}
}

func TestDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())

	query, args, err := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE age > $1 AND active = $2")
	assertArgs(t, args, 18, true)
}

func TestDollarInsert(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Set("email", "john@example.com").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `INSERT INTO users (name, email) VALUES ($1, $2)`)
	assertArgs(t, args, "John", "john@example.com")
}

func TestDollarUpdate(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Update("users").
		Set("name", "Jane").
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `UPDATE users SET name = $1 WHERE id = $2`)
	assertArgs(t, args, "Jane", 1)
}

func TestDollarDelete(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Delete("users").
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `DELETE FROM users WHERE id = $1`)
	assertArgs(t, args, 1)
}

func TestAtPPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())

	query, args, err := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("active = ?", true).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE age > @p1 AND active = @p2")
	assertArgs(t, args, 18, true)
}

func TestAtPInsert(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args, err := qb.Insert("users").
		Set("name", "John").
		Set("email", "john@example.com").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name, email) VALUES (@p1, @p2)")
	assertArgs(t, args, "John", "john@example.com")
}

func TestAtPUpdate(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args, err := qb.Update("users").
		Set("name", "Jane").
		Set("email", "jane@ex.com").
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET name = @p1, email = @p2 WHERE id = @p3")
	if len(args) != 3 {
		t.Fatalf("got %d args, want 3", len(args))
	}
}

func TestAtPDelete(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, args, err := qb.Delete("users").Where("id = ?", 1).Build()
	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM users WHERE id = @p1")
	assertArgs(t, args, 1)
}

func TestDollarWhereSkipsQuotedQuestionMarks(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("name = '?' AND id = ?", 42).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `SELECT id FROM users WHERE name = '?' AND id = $1`)
	assertArgs(t, args, 42)
}

func TestDollarWhereDoubleQuoted(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where(`col = "?" AND val = ?`, 99).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `SELECT id FROM users WHERE col = "?" AND val = $1`)
	assertArgs(t, args, 99)
}

func TestDollarWhereEscapedQuotes(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("name = 'it''s ?' AND val = ?", 7).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `SELECT id FROM users WHERE name = 'it''s ?' AND val = $1`)
	assertArgs(t, args, 7)
}

func TestDollarWhereMultipleParamsAcrossClauses(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("age > ?", 18).
		Where("name LIKE '?' AND active = ?", true).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `SELECT id FROM users WHERE age > $1 AND name LIKE '?' AND active = $2`)
	assertArgs(t, args, 18, true)
}

func TestDollarUpdateWithQuotedPlaceholder(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Update("users").
		Set("name", "Jane").
		Where("col = '?' AND id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `UPDATE users SET name = $1 WHERE col = '?' AND id = $2`)
	if len(args) != 2 {
		t.Fatalf("got %d args, want 2", len(args))
	}
}

func TestDollarDeleteWithQuotedPlaceholder(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Delete("users").
		Where("col = '?' AND id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `DELETE FROM users WHERE col = '?' AND id = $1`)
	assertArgs(t, args, 1)
}

func TestBackslashEscapesSkipsQuoted(t *testing.T) {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder, AppendLimit: LimitOffset, BackslashEscapes: true}
	qb := NewQueryBuilder(d)
	query, args, err := qb.Select("id").
		From("users").
		Where("name = 'it\\'s ?' AND val = ?", 7).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE name = 'it\\'s ?' AND val = ?")
	assertArgs(t, args, 7)
}

func TestSelectJoin(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("u.id", "u.name", "o.total").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.id, u.name, o.total FROM users u INNER JOIN orders o ON o.user_id = u.id WHERE u.active = ?")
	assertArgs(t, args, true)
}

func TestSelectMultipleJoins(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _, err := qb.Select("u.name", "o.total", "p.name").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Join("INNER JOIN products p ON p.id = o.product_id").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.name, o.total, p.name FROM users u INNER JOIN orders o ON o.user_id = u.id INNER JOIN products p ON p.id = o.product_id")
}

func TestSelectGroupBy(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("u.id", "COUNT(o.id) as order_count").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		GroupBy("u.id").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.id, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON o.user_id = u.id GROUP BY u.id")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestSelectGroupByHaving(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("u.id", "COUNT(o.id) as order_count").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		GroupBy("u.id").
		Having("COUNT(o.id) > ?", 2).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.id, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON o.user_id = u.id GROUP BY u.id HAVING COUNT(o.id) > ?")
	assertArgs(t, args, 2)
}

func TestSelectJoinWhereGroupByHavingOrderByLimit(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("u.name", "SUM(o.amount) as total_spent").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		GroupBy("u.name").
		Having("SUM(o.amount) > ?", 100).
		OrderBy("total_spent DESC").
		Limit(10).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.name, SUM(o.amount) as total_spent FROM users u INNER JOIN orders o ON o.user_id = u.id WHERE u.active = ? GROUP BY u.name HAVING SUM(o.amount) > ? ORDER BY total_spent DESC LIMIT 10")
	assertArgs(t, args, true, 100)
}

func TestSelectGroupByHavingDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("u.id", "COUNT(o.id)").
		From("users u").
		Join("LEFT JOIN orders o ON o.user_id = u.id").
		Where("u.active = ?", true).
		GroupBy("u.id").
		Having("COUNT(o.id) > ?", 5).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT u.id, COUNT(o.id) FROM users u LEFT JOIN orders o ON o.user_id = u.id WHERE u.active = $1 GROUP BY u.id HAVING COUNT(o.id) > $2")
	assertArgs(t, args, true, 5)
}

func TestMSSQLLimitOffsetWithOrderBy(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _, err := qb.Select("id", "name").
		From("users").
		OrderBy("name").
		Limit(10).
		Offset(20).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id, name FROM users ORDER BY name OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY")
}

func TestMSSQLLimitOnly(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _, err := qb.Select("id").
		From("users").
		OrderBy("id").
		Limit(5).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users ORDER BY id OFFSET 0 ROWS FETCH NEXT 5 ROWS ONLY")
}

func TestMSSQLOffsetOnly(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _, err := qb.Select("id").
		From("users").
		OrderBy("id").
		Offset(10).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users ORDER BY id OFFSET 10 ROWS")
}

func TestMSSQLLimitOffsetWithoutOrderBy(t *testing.T) {
	qb := NewQueryBuilder(atPDialect())
	query, _, err := qb.Select("id").
		From("users").
		Limit(10).
		Offset(5).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users ORDER BY (SELECT NULL) OFFSET 5 ROWS FETCH NEXT 10 ROWS ONLY")
}

func TestSelectAll(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.SelectAll().From("users").Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT * FROM users")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestSelectDistinct(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _, err := qb.Select("name").Distinct().From("users").Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT DISTINCT name FROM users")
}

func TestSelectDistinctAll(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, _, err := qb.SelectAll().Distinct().From("users").Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT DISTINCT * FROM users")
}

func TestOrWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("active = ?", true).
		OrWhere("role = ?", "admin").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE active = ? OR role = ?")
	assertArgs(t, args, true, "admin")
}

func TestOrWhereMultiple(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("a = ?", 1).
		OrWhere("b = ?", 2).
		Where("c = ?", 3).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE a = ? OR b = ? AND c = ?")
	assertArgs(t, args, 1, 2, 3)
}

func TestOrWhereWithDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("active = ?", true).
		OrWhere("role = ?", "admin").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE active = $1 OR role = $2")
	assertArgs(t, args, true, "admin")
}

func TestWhereGroup(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("active = ?", true).
		WhereGroup(func(g *WhereGroup) {
			g.Where("role = ?", "admin").OrWhere("role = ?", "moderator")
		}).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE active = ? AND (role = ? OR role = ?)")
	assertArgs(t, args, true, "admin", "moderator")
}

func TestOrWhereGroup(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		Where("active = ?", true).
		OrWhereGroup(func(g *WhereGroup) {
			g.Where("role = ?", "admin").Where("dept = ?", "IT")
		}).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE active = ? OR (role = ? AND dept = ?)")
	assertArgs(t, args, true, "admin", "IT")
}

func TestWhereGroupNested(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		WhereGroup(func(g *WhereGroup) {
			g.Where("a = ?", 1).OrWhere("b = ?", 2)
		}).
		Where("c = ?", 3).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE (a = ? OR b = ?) AND c = ?")
	assertArgs(t, args, 1, 2, 3)
}

func TestWhereBetween(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		WhereBetween("age", 18, 65).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE age BETWEEN ? AND ?")
	assertArgs(t, args, 18, 65)
}

func TestWhereIsNull(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		WhereIsNull("email").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE email IS NULL")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestWhereIsNotNull(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").
		From("users").
		WhereIsNotNull("email").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE email IS NOT NULL")
	if len(args) != 0 {
		t.Errorf("got args %v, want empty", args)
	}
}

func TestBatchInsert(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").
		Columns("name", "email").
		Values("Alice", "a@b.com").
		Values("Bob", "b@b.com").
		Values("Charlie", "c@b.com").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name, email) VALUES (?, ?), (?, ?), (?, ?)")
	if len(args) != 6 {
		t.Fatalf("got %d args, want 6", len(args))
	}
	assertArgs(t, args, "Alice", "a@b.com", "Bob", "b@b.com", "Charlie", "c@b.com")
}

func TestBatchInsertDollar(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Insert("users").
		Columns("name").
		Values("Alice").
		Values("Bob").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name) VALUES ($1), ($2)")
	assertArgs(t, args, "Alice", "Bob")
}

func TestBatchInsertRowLengthMismatch(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Insert("users").
		Columns("name", "email").
		Values("Alice", "a@b.com").
		Values("Bob").
		Build()

	if err != ErrBatchRowLength {
		t.Errorf("got err %v, want ErrBatchRowLength", err)
	}
}

func TestInsertReturningDollar(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Insert("users").
		Set("name", "Alice").
		Returning("id").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `INSERT INTO users (name) VALUES ($1) RETURNING id`)
	assertArgs(t, args, "Alice")
}

func newAtPDialectWithQuoting() *BaseDialect {
	d := &BaseDialect{Placeholder: AtPPlaceholder, AppendLimit: FetchNextLimit, QuoteStyle: BracketQuoting}
	d.PrependReturning = d.WriteOutput
	d.AppendReturning = d.WriteOutput
	d.AppendDeletedReturning = d.WriteDeletedOutput
	return d
}

func TestInsertReturningAtP(t *testing.T) {
	d := newAtPDialectWithQuoting()
	qb := NewQueryBuilder(d)
	query, args, err := qb.Insert("users").
		Set("name", "Alice").
		Returning("id", "name").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO [users] ([name]) OUTPUT INSERTED.[id], INSERTED.[name] VALUES (@p1)")
	assertArgs(t, args, "Alice")
}

func TestInsertReturningQuestionMark(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").
		Set("name", "Alice").
		Returning("id").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `INSERT INTO users (name) VALUES (?) RETURNING id`)
	assertArgs(t, args, "Alice")
}

func TestInClause(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	inVals, err := In(1, 2, 3)
	assertNoError(t, err)
	query, args, err := qb.Select("id").
		From("users").
		Where("id IN (?)", inVals).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE id IN (?, ?, ?)")
	assertArgs(t, args, 1, 2, 3)
}

func TestInClauseDollar(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	inVals, err := In(1, 2, 3)
	assertNoError(t, err)
	query, args, err := qb.Select("id").
		From("users").
		Where("id IN (?)", inVals).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE id IN ($1, $2, $3)")
	assertArgs(t, args, 1, 2, 3)
}

func TestInClauseWithOtherWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	inVals, err := In(1, 2)
	assertNoError(t, err)
	query, args, err := qb.Select("id").
		From("users").
		Where("active = ?", true).
		Where("id IN (?)", inVals).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE active = ? AND id IN (?, ?)")
	assertArgs(t, args, true, 1, 2)
}

func TestInEmptyReturnsError(t *testing.T) {
	_, err := In()
	if err != ErrEmptyInValues {
		t.Errorf("got err %v, want ErrEmptyInValues", err)
	}
}

func TestInTooManyReturnsError(t *testing.T) {
	vals := make([]interface{}, 1001)
	_, err := In(vals...)
	if err != ErrTooManyInValues {
		t.Errorf("got err %v, want ErrTooManyInValues", err)
	}
}

func TestUpdateOrWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Update("users").
		Set("active", false).
		Where("id = ?", 1).
		OrWhere("role = ?", "temp").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET active = ? WHERE id = ? OR role = ?")
	assertArgs(t, args, false, 1, "temp")
}

func TestDeleteOrWhere(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Delete("users").
		Where("id = ?", 1).
		OrWhere("id = ?", 2).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM users WHERE id = ? OR id = ?")
	assertArgs(t, args, 1, 2)
}

func TestSelectNoTable(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Select("id").Build()
	if err != ErrEmptyTable {
		t.Errorf("got err %v, want ErrEmptyTable", err)
	}
}

func TestDeleteNoTable(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Delete("").Build()
	if err != ErrEmptyTable {
		t.Errorf("got err %v, want ErrEmptyTable", err)
	}
}

func TestUpdateReturningDollar(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Update("users").
		Set("name", "Alice").
		Where("id = ?", 1).
		Returning("id", "name").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `UPDATE users SET name = $1 WHERE id = $2 RETURNING id, name`)
	assertArgs(t, args, "Alice", 1)
}

func TestUpdateReturningAtP(t *testing.T) {
	d := newAtPDialectWithQuoting()
	qb := NewQueryBuilder(d)
	query, args, err := qb.Update("users").
		Set("name", "Alice").
		Where("id = ?", 1).
		Returning("id", "name").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE [users] SET [name] = @p1 OUTPUT INSERTED.[id], INSERTED.[name] WHERE id = @p2")
	assertArgs(t, args, "Alice", 1)
}

func TestDeleteReturningDollar(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Delete("users").
		Where("id = ?", 1).
		Returning("id", "name").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, `DELETE FROM users WHERE id = $1 RETURNING id, name`)
	assertArgs(t, args, 1)
}

func TestDeleteReturningAtP(t *testing.T) {
	d := newAtPDialectWithQuoting()
	qb := NewQueryBuilder(d)
	query, args, err := qb.Delete("users").
		Where("id = ?", 1).
		Returning("id").
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "DELETE FROM [users] OUTPUT DELETED.[id] WHERE id = @p1")
	assertArgs(t, args, 1)
}

func TestQuoteSingleEscapes(t *testing.T) {
	tests := []struct {
		name       string
		quoteStyle QuoteStyle
		input      string
		want       string
	}{
		{"backtick", BacktickQuoting, "na`me", "`na``me`"},
		{"doublequote", DoubleQuoteQuoting, `na"me`, `"na""me"`},
		{"bracket", BracketQuoting, "na]me", "[na]]me]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BaseDialect{QuoteStyle: tt.quoteStyle}
			got := d.quoteSingle(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSafeIdentifier(t *testing.T) {
	if err := SafeIdentifier("users"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := SafeIdentifier("user_id"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := SafeIdentifier("users]; DROP TABLE users"); err == nil {
		t.Error("expected error for injection identifier")
	}
	if err := SafeIdentifier(""); err == nil {
		t.Error("expected error for empty identifier")
	}
	if err := SafeIdentifier("123abc"); err == nil {
		t.Error("expected error for identifier starting with digit")
	}
}

func TestOrHaving(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("country", "COUNT(*) as cnt").
		From("users").
		GroupBy("country").
		Having("COUNT(*) > ?", 10).
		OrHaving("AVG(score) < ?", 5).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT country, COUNT(*) as cnt FROM users GROUP BY country HAVING COUNT(*) > ? OR AVG(score) < ?")
	assertArgs(t, args, 10, 5)
}

func TestOrHavingWithDollarPlaceholders(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Select("country", "COUNT(*) as cnt").
		From("users").
		GroupBy("country").
		Having("COUNT(*) > ?", 10).
		OrHaving("AVG(score) < ?", 5).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "SELECT country, COUNT(*) as cnt FROM users GROUP BY country HAVING COUNT(*) > $1 OR AVG(score) < $2")
	assertArgs(t, args, 10, 5)
}

func TestErrReturningNotSupported(t *testing.T) {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder, AppendLimit: LimitOffset}
	qb := NewQueryBuilder(d)
	_, _, err := qb.Insert("users").Set("name", "Alice").Returning("id").Build()
	if err != ErrReturningNotSupported {
		t.Errorf("got err %v, want ErrReturningNotSupported", err)
	}
	_, _, err = qb.Update("users").Set("name", "Alice").Where("id = ?", 1).Returning("id").Build()
	if err != ErrReturningNotSupported {
		t.Errorf("got err %v, want ErrReturningNotSupported", err)
	}
	_, _, err = qb.Delete("users").Where("id = ?", 1).Returning("id").Build()
	if err != ErrReturningNotSupported {
		t.Errorf("got err %v, want ErrReturningNotSupported", err)
	}
}

func TestInsertEmptyTable(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Insert("").Set("name", "Alice").Build()
	if err != ErrEmptyTable {
		t.Errorf("got err %v, want ErrEmptyTable", err)
	}
}

func TestWhereIsNullWithQuoting(t *testing.T) {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder, AppendLimit: LimitOffset, QuoteStyle: BacktickQuoting}
	qb := NewQueryBuilder(d)
	query, _, err := qb.Select("id").From("users").WhereIsNull("email").Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT `id` FROM `users` WHERE `email` IS NULL")
}

func TestWhereBetweenWithQuoting(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder, AppendLimit: LimitOffset, QuoteStyle: DoubleQuoteQuoting}
	qb := NewQueryBuilder(d)
	query, args, err := qb.Select("id").From("users").WhereBetween("age", 18, 65).Build()
	assertNoError(t, err)
	assertQuery(t, query, `SELECT "id" FROM "users" WHERE "age" BETWEEN $1 AND $2`)
	assertArgs(t, args, 18, 65)
}

func TestInExactlyMaxValues(t *testing.T) {
	vals := make([]interface{}, MaxInValues)
	for i := range vals {
		vals[i] = i
	}
	inVals, err := In(vals...)
	if err != nil {
		t.Errorf("In() with exactly MaxInValues should succeed, got err: %v", err)
	}
	if len(inVals) != MaxInValues {
		t.Errorf("got %d values, want %d", len(inVals), MaxInValues)
	}
}

func TestInSingleValue(t *testing.T) {
	inVals, err := In(42)
	assertNoError(t, err)
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Select("id").From("users").Where("id IN (?)", inVals).Build()
	assertNoError(t, err)
	assertQuery(t, query, "SELECT id FROM users WHERE id IN (?)")
	assertArgs(t, args, 42)
}

func TestInsertSetMap(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetMap(map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (age, email, name) VALUES (?, ?, ?)")
	assertArgs(t, args, 30, "alice@example.com", "Alice")
}

func TestUpdateSetMap(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Update("users").
		SetMap(map[string]interface{}{
			"email": "new@example.com",
			"name":  "Alice",
		}).
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET email = ?, name = ? WHERE id = ?")
	assertArgs(t, args, "new@example.com", "Alice", 1)
}

func TestSetMapEmpty(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	_, _, err := qb.Insert("users").SetMap(map[string]interface{}{}).Build()
	if err != ErrEmptyColumns {
		t.Errorf("got err %v, want ErrEmptyColumns", err)
	}
}

func TestSetMapPreservesExistingSet(t *testing.T) {
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").
		Set("id", 1).
		SetMap(map[string]interface{}{
			"name":  "Alice",
			"email": "a@b.com",
		}).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (id, email, name) VALUES (?, ?, ?)")
	assertArgs(t, args, 1, "a@b.com", "Alice")
}

func TestInsertSetStruct(t *testing.T) {
	type User struct {
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"age"`
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)")
	assertArgs(t, args, "Alice", "alice@example.com", 30)
}

func TestUpdateSetStruct(t *testing.T) {
	type UpdateUser struct {
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Update("users").
		SetStruct(UpdateUser{Name: "Bob", Email: "bob@example.com"}).
		Where("id = ?", 1).
		Build()

	assertNoError(t, err)
	assertQuery(t, query, "UPDATE users SET name = ?, email = ? WHERE id = ?")
	assertArgs(t, args, "Bob", "bob@example.com", 1)
}

func TestSetStructPointerInput(t *testing.T) {
	type User struct {
		Name string `db:"name"`
	}

	u := &User{Name: "Alice"}
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(u).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name) VALUES (?)")
	assertArgs(t, args, "Alice")
}

func TestSetStructSkipsNilPointers(t *testing.T) {
	type User struct {
		Name  string  `db:"name"`
		Email *string `db:"email"`
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:  "Alice",
		Email: nil,
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name) VALUES (?)")
	assertArgs(t, args, "Alice")
}

func TestSetStructDereferencesNonNilPointers(t *testing.T) {
	type User struct {
		Name  string  `db:"name"`
		Email *string `db:"email"`
	}

	email := "alice@example.com"
	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:  "Alice",
		Email: &email,
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name, email) VALUES (?, ?)")
	assertArgs(t, args, "Alice", "alice@example.com")
}

func TestSetStructSkipsUnexported(t *testing.T) {
	type User struct {
		Name   string `db:"name"`
		secret string
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:   "Alice",
		secret: "hidden",
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name) VALUES (?)")
	assertArgs(t, args, "Alice")
}

func TestSetStructSkipsDashTag(t *testing.T) {
	type User struct {
		Name    string `db:"name"`
		Ignored string `db:"-"`
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:    "Alice",
		Ignored: "skip me",
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (name) VALUES (?)")
	assertArgs(t, args, "Alice")
}

func TestSetStructNoTagUsesSnakeCase(t *testing.T) {
	type User struct {
		FullName string
		Email    string
	}

	qb := NewQueryBuilder(defaultDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		FullName: "Alice Smith",
		Email:    "alice@example.com",
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, "INSERT INTO users (full_name, email) VALUES (?, ?)")
	assertArgs(t, args, "Alice Smith", "alice@example.com")
}

func TestSetStructDollarPlaceholders(t *testing.T) {
	type User struct {
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	qb := NewQueryBuilder(dollarDialect())
	query, args, err := qb.Insert("users").SetStruct(User{
		Name:  "Alice",
		Email: "alice@example.com",
	}).Build()

	assertNoError(t, err)
	assertQuery(t, query, `INSERT INTO users (name, email) VALUES ($1, $2)`)
	assertArgs(t, args, "Alice", "alice@example.com")
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
