package dal

import (
	"errors"
	"fmt"
	"strings"
)

// LimitStyle determines how LIMIT/OFFSET is rendered in SELECT queries.
type LimitStyle int

const (
	// LimitOffsetStyle uses "LIMIT X OFFSET Y" syntax (MySQL, PostgreSQL, SQLite).
	LimitOffsetStyle LimitStyle = iota
	// FetchNextStyle uses "OFFSET X ROWS FETCH NEXT Y ROWS ONLY" syntax (SQL Server).
	FetchNextStyle
)

// QuoteStyle determines how identifiers are quoted.
type QuoteStyle int

const (
	// NoQuoting leaves identifiers as-is.
	NoQuoting QuoteStyle = iota
	// BacktickQuoting uses `name` (MySQL).
	BacktickQuoting
	// DoubleQuoteQuoting uses "name" (PostgreSQL, SQLite).
	DoubleQuoteQuoting
	// BracketQuoting uses [name] (MSSQL).
	BracketQuoting
)

// Dialect abstracts database-specific SQL generation. Each supported database
// provides its own Dialect implementation.
type Dialect interface {
	BuildSelect(q *SelectQuery) (string, []interface{}, error)
	BuildInsert(q *InsertQuery) (string, []interface{}, error)
	BuildUpdate(q *UpdateQuery) (string, []interface{}, error)
	BuildDelete(q *DeleteQuery) (string, []interface{}, error)
	QuoteIdentifier(name string) string
	SupportsReturning() bool
}

// BaseDialect provides a common SQL generation implementation that covers
// MySQL, PostgreSQL, SQLite, and SQL Server. Embed and override methods for
// databases with additional quirks.
type BaseDialect struct {
	PlaceholderStyle PlaceholderStyle
	LimitStyle       LimitStyle
	QuoteStyle       QuoteStyle
	BackslashEscapes bool
}

// SupportsReturning returns true for dialects that support RETURNING/OUTPUT clauses.
func (d *BaseDialect) SupportsReturning() bool {
	return d.PlaceholderStyle == DollarNumber || d.PlaceholderStyle == AtPNumber || d.PlaceholderStyle == QuestionMark
}

func (d *BaseDialect) placeholder(idx int) string {
	switch d.PlaceholderStyle {
	case DollarNumber:
		return fmt.Sprintf("$%d", idx)
	case AtPNumber:
		return fmt.Sprintf("@p%d", idx)
	default:
		return "?"
	}
}

// replaceAndCount replaces unquoted '?' characters with numbered placeholders
// starting at startIdx, returning the replaced string and the number of
// placeholders that were substituted.
func (d *BaseDialect) replaceAndCount(sql string, startIdx int) (string, int) {
	if d.PlaceholderStyle == QuestionMark {
		return sql, countUnquoted(sql, d.BackslashEscapes)
	}

	var b strings.Builder
	b.Grow(len(sql) + 16)
	idx := startIdx
	count := 0
	inSingle := false
	inDouble := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		switch {
		case inSingle:
			if d.BackslashEscapes && ch == '\\' {
				b.WriteByte(ch)
				if i+1 < len(sql) {
					b.WriteByte(sql[i+1])
					i++
				}
			} else if ch == '\'' {
				b.WriteByte(ch)
				if i+1 < len(sql) && sql[i+1] == '\'' {
					b.WriteByte(sql[i+1])
					i++
				} else {
					inSingle = false
				}
			} else {
				b.WriteByte(ch)
			}
		case inDouble:
			if d.BackslashEscapes && ch == '\\' {
				b.WriteByte(ch)
				if i+1 < len(sql) {
					b.WriteByte(sql[i+1])
					i++
				}
			} else if ch == '"' {
				b.WriteByte(ch)
				if i+1 < len(sql) && sql[i+1] == '"' {
					b.WriteByte(sql[i+1])
					i++
				} else {
					inDouble = false
				}
			} else {
				b.WriteByte(ch)
			}
		case ch == '\'':
			b.WriteByte(ch)
			inSingle = true
		case ch == '"':
			b.WriteByte(ch)
			inDouble = true
		case ch == '?':
			b.WriteString(d.placeholder(idx))
			idx++
			count++
		default:
			b.WriteByte(ch)
		}
	}

	return b.String(), count
}

// countUnquoted counts '?' characters that are NOT inside quoted strings.
func countUnquoted(sql string, backslashEscapes bool) int {
	count := 0
	inSingle := false
	inDouble := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch {
		case inSingle:
			if backslashEscapes && ch == '\\' {
				if i+1 < len(sql) {
					i++
				}
			} else if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++
				} else {
					inSingle = false
				}
			}
		case inDouble:
			if backslashEscapes && ch == '\\' {
				if i+1 < len(sql) {
					i++
				}
			} else if ch == '"' {
				if i+1 < len(sql) && sql[i+1] == '"' {
					i++
				} else {
					inDouble = false
				}
			}
		case ch == '\'':
			inSingle = true
		case ch == '"':
			inDouble = true
		case ch == '?':
			count++
		}
	}
	return count
}

// buildClauses renders a slice of whereClauses into a joined string, tracking
// the placeholder index and collecting args. InValues args are expanded.
// Supports WhereGroup nesting via groupConnector.
func (d *BaseDialect) buildClauses(clauses []whereClause, paramIdx int, args []interface{}) (string, int, []interface{}, error) {
	parts := make([]string, len(clauses))
	for i, w := range clauses {
		connector := "AND"
		if w.connector == orConnector {
			connector = "OR"
		}

		if w.connector == groupConnector || w.connector == groupOrConnector {
			if w.connector == groupOrConnector {
				connector = "OR"
			}
			groupStr, newIdx, newArgs, err := d.buildClauses(w.children, paramIdx, args)
			if err != nil {
				return "", 0, nil, err
			}
			paramIdx = newIdx
			args = newArgs
			parts[i] = connector + " (" + groupStr + ")"
			continue
		}

		condition := w.condition
		var expandedArgs []interface{}

		for _, arg := range w.args {
			if in, ok := arg.(InValues); ok {
				if len(in) == 0 {
					return "", 0, nil, ErrEmptyInValues
				}
				condition = expandInPlaceholders(condition, len(in))
				expandedArgs = append(expandedArgs, []interface{}(in)...)
			} else {
				expandedArgs = append(expandedArgs, arg)
			}
		}

		replaced, n := d.replaceAndCount(condition, paramIdx)
		paramIdx += n
		args = append(args, expandedArgs...)
		parts[i] = connector + " " + replaced
	}

	result := ""
	if len(parts) > 0 {
		result = strings.TrimPrefix(parts[0], "AND ")
		for _, p := range parts[1:] {
			result += " " + p
		}
	}

	return result, paramIdx, args, nil
}

// expandInPlaceholders replaces the first unquoted "?" with N "?" placeholders for IN expansion.
func expandInPlaceholders(condition string, count int) string {
	idx := findFirstUnquotedPlaceholder(condition)
	if idx == -1 {
		return condition
	}

	var b strings.Builder
	b.Grow(len(condition) + count*3)
	b.WriteString(condition[:idx])
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("?")
	}
	b.WriteString(condition[idx+1:])
	return b.String()
}

// findFirstUnquotedPlaceholder finds the index of the first ? not inside quotes.
func findFirstUnquotedPlaceholder(s string) int {
	inSingle := false
	inDouble := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case inSingle:
			if ch == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
				} else {
					inSingle = false
				}
			}
		case inDouble:
			if ch == '"' {
				if i+1 < len(s) && s[i+1] == '"' {
					i++
				} else {
					inDouble = false
				}
			}
		case ch == '\'':
			inSingle = true
		case ch == '"':
			inDouble = true
		case ch == '?':
			return i
		}
	}
	return -1
}

// QuoteIdentifier wraps an identifier with the dialect's quoting style.
// Skips quoting for expressions containing "(", "AS", spaces (aliases), or "*".
func (d *BaseDialect) QuoteIdentifier(name string) string {
	if name == "" || name == "*" {
		return name
	}

	if strings.Contains(name, "(") ||
		strings.Contains(name, " AS ") ||
		strings.Contains(name, " as ") ||
		strings.Contains(name, " ") ||
		strings.Contains(name, ",") {
		return name
	}

	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		quoted := make([]string, len(parts))
		for i, p := range parts {
			quoted[i] = d.quoteSingle(p)
		}
		return strings.Join(quoted, ".")
	}

	return d.quoteSingle(name)
}

func (d *BaseDialect) quoteSingle(name string) string {
	switch d.QuoteStyle {
	case BacktickQuoting:
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case DoubleQuoteQuoting:
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case BracketQuoting:
		return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
	default:
		return name
	}
}

// quoteColumns joins and optionally quotes column names.
func (d *BaseDialect) quoteColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	return strings.Join(quoted, ", ")
}

func (d *BaseDialect) BuildSelect(q *SelectQuery) (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, ErrEmptyTable
	}

	var b strings.Builder
	cols := "*"
	if len(q.columns) > 0 {
		cols = d.quoteColumns(q.columns)
	}

	b.WriteString("SELECT ")
	if q.distinct {
		b.WriteString("DISTINCT ")
	}
	b.WriteString(cols)
	b.WriteString(" FROM ")
	b.WriteString(d.QuoteIdentifier(q.table))

	var args []interface{}
	paramIdx := 1

	if len(q.joins) > 0 {
		b.WriteString(" ")
		b.WriteString(strings.Join(q.joins, " "))
	}

	if len(q.wheres) > 0 {
		whereStr, newIdx, newArgs, err := d.buildClauses(q.wheres, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		paramIdx = newIdx
		args = newArgs
		b.WriteString(" WHERE ")
		b.WriteString(whereStr)
	}

	if len(q.groupBy) > 0 {
		quoted := make([]string, len(q.groupBy))
		for i, c := range q.groupBy {
			quoted[i] = d.QuoteIdentifier(c)
		}
		b.WriteString(" GROUP BY ")
		b.WriteString(strings.Join(quoted, ", "))
	}

	if len(q.having) > 0 {
		havingStr, newIdx, newArgs, err := d.buildClauses(q.having, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		paramIdx = newIdx
		args = newArgs
		b.WriteString(" HAVING ")
		b.WriteString(havingStr)
	}

	if len(q.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(strings.Join(q.orderBy, ", "))
	}

	if d.LimitStyle == FetchNextStyle && (q.limit != nil || q.offset != nil) {
		if len(q.orderBy) == 0 {
			b.WriteString(" ORDER BY (SELECT NULL)")
		}
		if q.offset != nil {
			fmt.Fprintf(&b, " OFFSET %d ROWS", *q.offset)
		} else {
			b.WriteString(" OFFSET 0 ROWS")
		}
		if q.limit != nil {
			fmt.Fprintf(&b, " FETCH NEXT %d ROWS ONLY", *q.limit)
		}
	} else {
		if q.limit != nil {
			fmt.Fprintf(&b, " LIMIT %d", *q.limit)
		}
		if q.offset != nil {
			fmt.Fprintf(&b, " OFFSET %d", *q.offset)
		}
	}

	return b.String(), args, nil
}

func (d *BaseDialect) BuildInsert(q *InsertQuery) (string, []interface{}, error) {
	if len(q.keys) == 0 && len(q.rows) == 0 {
		return "", nil, ErrEmptyColumns
	}

	if len(q.returning) > 0 && !d.SupportsReturning() {
		return "", nil, ErrReturningNotSupported
	}

	var b strings.Builder

	if len(q.rows) == 0 {
		placeholders := make([]string, len(q.keys))
		for i := range placeholders {
			placeholders[i] = d.placeholder(i + 1)
		}

		quoted := d.quoteColumns(q.keys)
		b.WriteString("INSERT INTO ")
		b.WriteString(d.QuoteIdentifier(q.table))
		b.WriteString(" (")
		b.WriteString(quoted)
		b.WriteString(")")

		if len(q.returning) > 0 && d.PlaceholderStyle == AtPNumber {
			b.WriteString(" ")
			b.WriteString(d.buildOutput(q.returning))
		}

		b.WriteString(" VALUES (")
		b.WriteString(strings.Join(placeholders, ", "))
		b.WriteString(")")

		if len(q.returning) > 0 && d.PlaceholderStyle != AtPNumber {
			b.WriteString(d.buildReturning(q.returning))
		}

		return b.String(), q.values, nil
	}

	for _, row := range q.rows {
		if len(row) != len(q.keys) {
			return "", nil, ErrBatchRowLength
		}
	}

	quoted := d.quoteColumns(q.keys)

	var allPlaceholders []string
	var allArgs []interface{}
	for rowIdx, row := range q.rows {
		rowPhs := make([]string, len(q.keys))
		for i := range q.keys {
			rowPhs[i] = d.placeholder(rowIdx*len(q.keys) + i + 1)
		}
		allPlaceholders = append(allPlaceholders, "("+strings.Join(rowPhs, ", ")+")")
		allArgs = append(allArgs, row...)
	}

	b.WriteString("INSERT INTO ")
	b.WriteString(d.QuoteIdentifier(q.table))
	b.WriteString(" (")
	b.WriteString(quoted)
	b.WriteString(")")

	if len(q.returning) > 0 && d.PlaceholderStyle == AtPNumber {
		b.WriteString(" ")
		b.WriteString(d.buildOutput(q.returning))
	}

	b.WriteString(" VALUES ")
	b.WriteString(strings.Join(allPlaceholders, ", "))

	if len(q.returning) > 0 && d.PlaceholderStyle != AtPNumber {
		b.WriteString(d.buildReturning(q.returning))
	}

	return b.String(), allArgs, nil
}

func (d *BaseDialect) BuildUpdate(q *UpdateQuery) (string, []interface{}, error) {
	if len(q.keys) == 0 {
		return "", nil, ErrEmptyColumns
	}

	if len(q.returning) > 0 && !d.SupportsReturning() {
		return "", nil, ErrReturningNotSupported
	}

	var b strings.Builder
	setClauses := make([]string, len(q.keys))
	for i, key := range q.keys {
		setClauses[i] = fmt.Sprintf("%s = %s", d.QuoteIdentifier(key), d.placeholder(i+1))
	}

	b.WriteString("UPDATE ")
	b.WriteString(d.QuoteIdentifier(q.table))
	b.WriteString(" SET ")
	b.WriteString(strings.Join(setClauses, ", "))

	args := make([]interface{}, len(q.values))
	copy(args, q.values)
	paramIdx := len(q.keys) + 1

	if len(q.wheres) > 0 {
		whereStr, newIdx, newArgs, err := d.buildClauses(q.wheres, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		paramIdx = newIdx
		args = newArgs
		b.WriteString(" WHERE ")
		b.WriteString(whereStr)
	}

	if len(q.returning) > 0 {
		if d.PlaceholderStyle == AtPNumber {
			b.WriteString(" ")
			b.WriteString(d.buildOutput(q.returning))
		} else {
			b.WriteString(d.buildReturning(q.returning))
		}
	}

	return b.String(), args, nil
}

func (d *BaseDialect) BuildDelete(q *DeleteQuery) (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, ErrEmptyTable
	}

	if len(q.returning) > 0 && !d.SupportsReturning() {
		return "", nil, ErrReturningNotSupported
	}

	var b strings.Builder
	b.WriteString("DELETE FROM ")
	b.WriteString(d.QuoteIdentifier(q.table))

	var args []interface{}
	paramIdx := 1

	if len(q.wheres) > 0 {
		whereStr, newIdx, newArgs, err := d.buildClauses(q.wheres, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		paramIdx = newIdx
		args = newArgs
		b.WriteString(" WHERE ")
		b.WriteString(whereStr)
	}

	if len(q.returning) > 0 {
		if d.PlaceholderStyle == AtPNumber {
			b.WriteString(" ")
			b.WriteString(d.buildOutput(q.returning))
		} else {
			b.WriteString(d.buildReturning(q.returning))
		}
	}

	return b.String(), args, nil
}

func (d *BaseDialect) buildReturning(columns []string) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	return " RETURNING " + strings.Join(quoted, ", ")
}

func (d *BaseDialect) buildOutput(columns []string) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = "INSERTED." + d.QuoteIdentifier(c)
	}
	return "OUTPUT " + strings.Join(quoted, ", ")
}

// SafeIdentifier validates that a name is a safe SQL identifier (letters, digits, underscores, dots).
func SafeIdentifier(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("dal: invalid identifier %q", name)
	}
	for i, ch := range name {
		if ch == '.' {
			continue
		}
		if i == 0 {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
				return fmt.Errorf("dal: invalid identifier %q", name)
			}
		} else {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return fmt.Errorf("dal: invalid identifier %q", name)
			}
		}
	}
	return nil
}

// JoinIdentifiers safely quotes and joins identifiers with a separator,
// validating each one first.
func JoinIdentifiers(dialect Dialect, sep string, names ...string) (string, error) {
	for _, n := range names {
		if err := SafeIdentifier(n); err != nil {
			return "", err
		}
	}
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = dialect.QuoteIdentifier(n)
	}
	return strings.Join(quoted, sep), nil
}

// ErrNotImplemented is returned for features not yet implemented.
var ErrNotImplemented = errors.New("dal: feature not implemented")
