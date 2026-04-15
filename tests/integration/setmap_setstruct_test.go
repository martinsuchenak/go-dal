package integration

import (
	"context"
	"testing"
)

func TestSetMapInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		result, err := td.dalDB.Insert("users").
			SetMap(map[string]interface{}{
				"name":   "MapUser",
				"email":  "map@example.com",
				"active": true,
			}).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setmap insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		var email string
		err = td.dalDB.Select("name", "email").
			From("users").
			Where("name = ?", "MapUser").
			QueryRow(ctx).Scan(&name, &email)
		if err != nil {
			t.Fatalf("select after setmap insert failed: %v", err)
		}
		if name != "MapUser" {
			t.Errorf("got name %q, want 'MapUser'", name)
		}
		if email != "map@example.com" {
			t.Errorf("got email %q, want 'map@example.com'", email)
		}
	})
}

func TestSetMapInsertDeterministicOrder(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		qb := td.dalDB.NewQueryBuilder()

		query1, _, err := qb.Insert("users").
			SetMap(map[string]interface{}{
				"active": true,
				"email":  "a@b.com",
				"name":   "Zeta",
			}).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		qb = td.dalDB.NewQueryBuilder()
		query2, _, err := qb.Insert("users").
			SetMap(map[string]interface{}{
				"name":   "Zeta",
				"active": true,
				"email":  "a@b.com",
			}).
			Build()
		if err != nil {
			t.Fatal(err)
		}

		if query1 != query2 {
			t.Errorf("SetMap should produce deterministic SQL:\n  %s\n  %s", query1, query2)
		}
	})
}

func TestSetMapUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		result, err := td.dalDB.Update("users").
			SetMap(map[string]interface{}{
				"email": "alice_map@example.com",
			}).
			Where("name = ?", "Alice").
			Exec(ctx)
		if err != nil {
			t.Fatalf("setmap update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var email string
		err = td.dalDB.Select("email").
			From("users").
			Where("name = ?", "Alice").
			QueryRow(ctx).Scan(&email)
		if err != nil {
			t.Fatalf("select after setmap update failed: %v", err)
		}
		if email != "alice_map@example.com" {
			t.Errorf("got email %q, want 'alice_map@example.com'", email)
		}
	})
}

func TestSetMapMultipleColumns(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		result, err := td.dalDB.Update("users").
			SetMap(map[string]interface{}{
				"name":  "AliceUpdated",
				"email": "alice_upd@example.com",
			}).
			Where("name = ?", "Alice").
			Exec(ctx)
		if err != nil {
			t.Fatalf("setmap multi-column update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name, email string
		err = td.dalDB.Select("name", "email").
			From("users").
			Where("email = ?", "alice_upd@example.com").
			QueryRow(ctx).Scan(&name, &email)
		if err != nil {
			t.Fatalf("select after setmap multi-column update failed: %v", err)
		}
		if name != "AliceUpdated" {
			t.Errorf("got name %q, want 'AliceUpdated'", name)
		}
	})
}

type testUser struct {
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active bool   `db:"active"`
}

type testUserPartial struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

type testUserWithDash struct {
	Name     string `db:"name"`
	Email    string `db:"email"`
	Password string `db:"-"`
}

type testUserNoTags struct {
	Name  string
	Email string
}

type testUserWithNilPtr struct {
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active *bool  `db:"active"`
}

type testUserPtrActive struct {
	Name   string `db:"name"`
	Email  string `db:"email"`
	Active *bool  `db:"active"`
}

func TestSetStructInsert(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUser{
			Name:   "StructUser",
			Email:  "struct@example.com",
			Active: true,
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name, email string
		err = td.dalDB.Select("name", "email").
			From("users").
			Where("name = ?", "StructUser").
			QueryRow(ctx).Scan(&name, &email)
		if err != nil {
			t.Fatalf("select after setstruct insert failed: %v", err)
		}
		if name != "StructUser" {
			t.Errorf("got name %q, want 'StructUser'", name)
		}
		if email != "struct@example.com" {
			t.Errorf("got email %q, want 'struct@example.com'", email)
		}
	})
}

func TestSetStructInsertPointer(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := &testUser{
			Name:   "PtrUser",
			Email:  "ptr@example.com",
			Active: true,
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct pointer insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "PtrUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after setstruct pointer insert failed: %v", err)
		}
		if name != "PtrUser" {
			t.Errorf("got name %q, want 'PtrUser'", name)
		}
	})
}

func TestSetStructInsertSkipsDashTag(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUserWithDash{
			Name:     "DashUser",
			Email:    "dash@example.com",
			Password: "secret123",
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Set("active", true).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct dash-tag insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "DashUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after setstruct dash-tag insert failed: %v", err)
		}
		if name != "DashUser" {
			t.Errorf("got name %q, want 'DashUser'", name)
		}
	})
}

func TestSetStructInsertNoTagUsesSnakeCase(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUserNoTags{
			Name:  "SnakeUser",
			Email: "snake@example.com",
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Set("active", true).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct snake-case insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "SnakeUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after setstruct snake-case insert failed: %v", err)
		}
		if name != "SnakeUser" {
			t.Errorf("got name %q, want 'SnakeUser'", name)
		}
	})
}

func TestSetStructInsertSkipsNilPointers(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUserWithNilPtr{
			Name:   "NilPtrUser",
			Email:  "nilptr@example.com",
			Active: nil,
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Set("active", true).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct nil-ptr insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "NilPtrUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after setstruct nil-ptr insert failed: %v", err)
		}
		if name != "NilPtrUser" {
			t.Errorf("got name %q, want 'NilPtrUser'", name)
		}
	})
}

func TestSetStructInsertDereferencesNonNilPointers(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		active := true
		u := testUserPtrActive{
			Name:   "PtrValUser",
			Email:  "ptrval@example.com",
			Active: &active,
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct ptr-val insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		err = td.dalDB.Select("name").
			From("users").
			Where("name = ?", "PtrValUser").
			QueryRow(ctx).Scan(&name)
		if err != nil {
			t.Fatalf("select after setstruct ptr-val insert failed: %v", err)
		}
		if name != "PtrValUser" {
			t.Errorf("got name %q, want 'PtrValUser'", name)
		}
	})
}

func TestSetStructUpdate(t *testing.T) {
	runForEachDBWithSeed(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUserPartial{
			Name:  "AliceStruct",
			Email: "alice_struct@example.com",
		}

		result, err := td.dalDB.Update("users").
			SetStruct(u).
			Where("name = ?", "Alice").
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct update failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name, email string
		err = td.dalDB.Select("name", "email").
			From("users").
			Where("email = ?", "alice_struct@example.com").
			QueryRow(ctx).Scan(&name, &email)
		if err != nil {
			t.Fatalf("select after setstruct update failed: %v", err)
		}
		if name != "AliceStruct" {
			t.Errorf("got name %q, want 'AliceStruct'", name)
		}
		if email != "alice_struct@example.com" {
			t.Errorf("got email %q, want 'alice_struct@example.com'", email)
		}
	})
}

func TestSetStructCombinedWithSet(t *testing.T) {
	runForEachDB(t, func(t *testing.T, td *testDB) {
		ctx := context.Background()

		u := testUserPartial{
			Name:  "ComboUser",
			Email: "combo@example.com",
		}

		result, err := td.dalDB.Insert("users").
			SetStruct(u).
			Set("active", true).
			Exec(ctx)
		if err != nil {
			t.Fatalf("setstruct+set insert failed: %v", err)
		}

		rows, _ := result.RowsAffected()
		if rows != 1 {
			t.Errorf("expected 1 row affected, got %d", rows)
		}

		var name string
		var active bool
		err = td.dalDB.Select("name", "active").
			From("users").
			Where("name = ?", "ComboUser").
			QueryRow(ctx).Scan(&name, &active)
		if err != nil {
			t.Fatalf("select after setstruct+set insert failed: %v", err)
		}
		if name != "ComboUser" {
			t.Errorf("got name %q, want 'ComboUser'", name)
		}
		if !active {
			t.Error("expected active=true")
		}
	})
}
