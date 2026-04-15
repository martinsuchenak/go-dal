package dal

import "testing"

func TestConcatExpr(t *testing.T) {
	d := &BaseDialect{}
	got := d.ConcatExpr("a", "b", "c")
	want := "CONCAT(a, b, c)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLengthExpr(t *testing.T) {
	d := &BaseDialect{}
	got := d.LengthExpr("name")
	want := "LENGTH(name)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCurrentTimestamp(t *testing.T) {
	d := &BaseDialect{}
	if got := d.CurrentTimestamp(); got != "NOW()" {
		t.Errorf("got %q, want NOW()", got)
	}
}

func TestBoolLiteral(t *testing.T) {
	d := &BaseDialect{}
	if got := d.BoolLiteral(true); got != "TRUE" {
		t.Errorf("BoolLiteral(true) = %q, want TRUE", got)
	}
	if got := d.BoolLiteral(false); got != "FALSE" {
		t.Errorf("BoolLiteral(false) = %q, want FALSE", got)
	}
}

func TestStringAggExpr(t *testing.T) {
	d := &BaseDialect{}
	got := d.StringAggExpr("name", "', '")
	want := "GROUP_CONCAT(name SEPARATOR ', ')"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRandExpr(t *testing.T) {
	d := &BaseDialect{}
	if got := d.RandExpr(); got != "RAND()" {
		t.Errorf("got %q, want RAND()", got)
	}
}

func TestTranslateSQLQuestionMark(t *testing.T) {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = ? WHERE id = ?"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLDollar(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ? AND name = ?")
	want := "UPDATE users SET x = $1 WHERE id = $2 AND name = $3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLAtP(t *testing.T) {
	d := &BaseDialect{Placeholder: AtPPlaceholder}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = @p1 WHERE id = @p2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLSkipsQuotedPlaceholders(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE name = '?' AND id = ?")
	want := "UPDATE users SET x = $1 WHERE name = '?' AND id = $2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLDoubleQuoted(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL(`UPDATE users SET x = ? WHERE col = "?" AND id = ?`)
	want := `UPDATE users SET x = $1 WHERE col = "?" AND id = $2`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLEscapedQuotes(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE name = 'it''s ?' AND id = ?")
	want := "UPDATE users SET x = $1 WHERE name = 'it''s ?' AND id = $2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLBackslashEscapes(t *testing.T) {
	d := &BaseDialect{Placeholder: QuestionMarkPlaceholder, BackslashEscapes: true}
	got := d.TranslateSQL("UPDATE users SET x = ? WHERE name = 'it\\'s ?' AND id = ?")
	want := "UPDATE users SET x = ? WHERE name = 'it\\'s ?' AND id = ?"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLNoPlaceholders(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL("SELECT NOW()")
	want := "SELECT NOW()"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLViaQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder(dollarDialect())
	got := qb.TranslateSQL("UPDATE users SET x = ? WHERE id = ?")
	want := "UPDATE users SET x = $1 WHERE id = $2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslateSQLSubquery(t *testing.T) {
	d := &BaseDialect{Placeholder: DollarPlaceholder}
	got := d.TranslateSQL("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? AND active = ?)")
	want := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND active = $2)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
