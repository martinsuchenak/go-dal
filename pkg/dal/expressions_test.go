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
