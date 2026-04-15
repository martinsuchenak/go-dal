package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/martinsuchenak/go-dal/pkg/sqlite"

	_ "modernc.org/sqlite"
)

func Example_sqliteCRUD() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dal := sqlite.NewSQLiteDB(db, nil)

	ctx := context.Background()
	dal.Exec(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")

	qb := sqlite.NewQueryBuilder()

	// Insert
	insertQ, insertArgs, err := qb.Insert("users").
		Set("name", "Alice").
		Set("email", "alice@example.com").
		Build()
	if err != nil {
		log.Fatal(err)
	}
	dal.Exec(ctx, insertQ, insertArgs...)

	// Select
	qb = sqlite.NewQueryBuilder()
	selQ, selArgs, err := qb.Select("name", "email").
		From("users").
		Where("name = ?", "Alice").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	var name, email string
	dal.QueryRow(ctx, selQ, selArgs...).Scan(&name, &email)
	fmt.Println(name, email)

	// Update
	qb = sqlite.NewQueryBuilder()
	updQ, updArgs, err := qb.Update("users").
		Set("email", "alice_new@example.com").
		Where("name = ?", "Alice").
		Build()
	if err != nil {
		log.Fatal(err)
	}
	dal.Exec(ctx, updQ, updArgs...)

	// Delete
	qb = sqlite.NewQueryBuilder()
	delQ, delArgs, err := qb.Delete("users").
		Where("name = ?", "Alice").
		Build()
	if err != nil {
		log.Fatal(err)
	}
	dal.Exec(ctx, delQ, delArgs...)

	// Output: Alice alice@example.com
}
