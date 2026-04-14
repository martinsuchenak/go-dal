package dal_test

import (
	"fmt"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

func Example_queryBuilder() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Select("id", "name").
		From("users").
		Where("active = ?", true).
		Where("age > ?", 18).
		OrderBy("name").
		Limit(10).
		Offset(5).
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users WHERE active = ? AND age > ? ORDER BY name LIMIT 10 OFFSET 5
	// [true 18]
}

func Example_insertQuery() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Insert("users").
		Set("name", "John").
		Set("email", "john@example.com").
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// INSERT INTO users (name, email) VALUES (?, ?)
	// [John john@example.com]
}

func Example_updateQuery() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Update("users").
		Set("email", "new@example.com").
		Where("id = ?", 1).
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// UPDATE users SET email = ? WHERE id = ?
	// [new@example.com 1]
}

func Example_deleteQuery() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Delete("users").
		Where("active = ?", false).
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// DELETE FROM users WHERE active = ?
	// [false]
}

func Example_selectWithJoin() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.QuestionMark, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Select("u.name", "o.total").
		From("users u").
		Join("INNER JOIN orders o ON o.user_id = u.id").
		Where("o.total > ?", 100).
		OrderBy("o.total DESC").
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT u.name, o.total FROM users u INNER JOIN orders o ON o.user_id = u.id WHERE o.total > ? ORDER BY o.total DESC
	// [100]
}

func Example_postgresPlaceholders() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.DollarNumber, LimitStyle: dal.LimitOffsetStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Select("id").
		From("users").
		Where("active = ?", true).
		Where("age > ?", 18).
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT id FROM users WHERE active = $1 AND age > $2
	// [true 18]
}

func Example_mssqlPlaceholders() {
	d := &dal.BaseDialect{PlaceholderStyle: dal.AtPNumber, LimitStyle: dal.FetchNextStyle}
	qb := dal.NewQueryBuilder(d)

	query, args := qb.Select("id", "name").
		From("users").
		OrderBy("name").
		Limit(10).
		Offset(20).
		Build()

	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// SELECT id, name FROM users ORDER BY name OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY
	// []
}
