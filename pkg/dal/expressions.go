package dal

import (
	"fmt"
	"strings"
)

// ConcatExpr returns a SQL string concatenation expression for the given parts.
// Default: CONCAT(a, b, c) — works for MySQL, PostgreSQL, SQLite (3.44+).
// MSSQL overrides to a + b + c.
func (d *BaseDialect) ConcatExpr(parts ...string) string {
	return "CONCAT(" + strings.Join(parts, ", ") + ")"
}

// LengthExpr returns a SQL string length expression for the given column.
// Default: LENGTH(col) — works for MySQL, PostgreSQL, SQLite.
// MSSQL overrides to LEN(col).
func (d *BaseDialect) LengthExpr(col string) string {
	return "LENGTH(" + col + ")"
}

// CurrentTimestamp returns a SQL expression for the current date and time.
// Default: NOW() — works for MySQL, PostgreSQL.
// SQLite overrides to datetime('now'), MSSQL overrides to GETDATE().
func (d *BaseDialect) CurrentTimestamp() string {
	return "NOW()"
}

// BoolLiteral returns a SQL boolean literal.
// Default: TRUE / FALSE — works for MySQL, PostgreSQL.
// SQLite and MSSQL override to 1 / 0.
func (d *BaseDialect) BoolLiteral(v bool) string {
	if v {
		return "TRUE"
	}
	return "FALSE"
}

// StringAggExpr returns a SQL string aggregation expression.
// Default: GROUP_CONCAT(col SEPARATOR sep) — MySQL syntax.
// SQLite overrides to GROUP_CONCAT(col, sep).
// PostgreSQL and MSSQL override to STRING_AGG(col, sep).
func (d *BaseDialect) StringAggExpr(col, sep string) string {
	return fmt.Sprintf("GROUP_CONCAT(%s SEPARATOR %s)", col, sep)
}

// RandExpr returns a SQL expression that generates a random value.
// Default: RAND() — returns float [0, 1) for MySQL, MSSQL.
// PostgreSQL and SQLite override to RANDOM().
func (d *BaseDialect) RandExpr() string {
	return "RAND()"
}

// --- QueryBuilder convenience wrappers ---

// ConcatExpr returns a SQL concatenation expression using this builder's dialect.
func (qb *QueryBuilder) ConcatExpr(parts ...string) string {
	return qb.dialect.ConcatExpr(parts...)
}

// LengthExpr returns a SQL length expression using this builder's dialect.
func (qb *QueryBuilder) LengthExpr(col string) string {
	return qb.dialect.LengthExpr(col)
}

// CurrentTimestamp returns a SQL current timestamp expression using this builder's dialect.
func (qb *QueryBuilder) CurrentTimestamp() string {
	return qb.dialect.CurrentTimestamp()
}

// BoolLiteral returns a SQL boolean literal using this builder's dialect.
func (qb *QueryBuilder) BoolLiteral(v bool) string {
	return qb.dialect.BoolLiteral(v)
}

// StringAggExpr returns a SQL string aggregation expression using this builder's dialect.
func (qb *QueryBuilder) StringAggExpr(col, sep string) string {
	return qb.dialect.StringAggExpr(col, sep)
}

// RandExpr returns a SQL random expression using this builder's dialect.
func (qb *QueryBuilder) RandExpr() string {
	return qb.dialect.RandExpr()
}

// TranslateSQL replaces ? placeholders in raw SQL with the dialect's format.
// Useful for SQL that can't be expressed through the query builder
// (e.g., column expressions, subqueries).
func (qb *QueryBuilder) TranslateSQL(query string) string {
	return qb.dialect.TranslateSQL(query)
}
