package dal

import (
	"errors"
	"fmt"
	"strings"
)

// QuoteStyle determines how identifiers are quoted.
type QuoteStyle int

const (
	NoQuoting          QuoteStyle = iota
	BacktickQuoting               // `name` — MySQL
	DoubleQuoteQuoting            // "name" — PostgreSQL, SQLite
	BracketQuoting                // [name] — MSSQL
)

// QuestionMarkPlaceholder returns "?" for all indices (MySQL, SQLite).
func QuestionMarkPlaceholder(int) string { return "?" }

// DollarPlaceholder returns "$1", "$2", ... for PostgreSQL.
func DollarPlaceholder(idx int) string { return fmt.Sprintf("$%d", idx) }

// AtPPlaceholder returns "@p1", "@p2", ... for SQL Server.
func AtPPlaceholder(idx int) string { return fmt.Sprintf("@p%d", idx) }

// LimitOffset appends "LIMIT X OFFSET Y" to the query (MySQL, PostgreSQL, SQLite).
func LimitOffset(b *strings.Builder, _ []string, limit, offset *int64) {
	if limit != nil {
		fmt.Fprintf(b, " LIMIT %d", *limit)
	}
	if offset != nil {
		fmt.Fprintf(b, " OFFSET %d", *offset)
	}
}

// FetchNextLimit appends "OFFSET X ROWS FETCH NEXT Y ROWS ONLY" (SQL Server).
func FetchNextLimit(b *strings.Builder, orderBy []string, limit, offset *int64) {
	if len(orderBy) == 0 {
		b.WriteString(" ORDER BY (SELECT NULL)")
	}
	if offset != nil {
		fmt.Fprintf(b, " OFFSET %d ROWS", *offset)
	} else {
		b.WriteString(" OFFSET 0 ROWS")
	}
	if limit != nil {
		fmt.Fprintf(b, " FETCH NEXT %d ROWS ONLY", *limit)
	}
}

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
// MySQL, PostgreSQL, SQLite, and SQL Server. Drivers configure behavior via
// function fields — no need to override Build methods unless introducing a
// genuinely novel SQL variant.
//
// Configure via struct literal, selecting the appropriate helper functions:
//
//	&dal.BaseDialect{
//	    Placeholder: dal.DollarPlaceholder,
//	    AppendLimit: dal.LimitOffset,
//	    QuoteStyle:  dal.DoubleQuoteQuoting,
//	}
//
// For RETURNING support, set hooks using the dialect's own formatting methods:
//
//	d.AppendReturning = d.WriteReturning  // PostgreSQL/SQLite
//	// or: d.PrependReturning = d.WriteOutput  // MSSQL
type BaseDialect struct {
	// Placeholder formats a parameter placeholder for the given 1-based index.
	// Use QuestionMarkPlaceholder, DollarPlaceholder, or AtPPlaceholder.
	Placeholder func(idx int) string

	// AppendLimit appends the LIMIT/OFFSET (or equivalent) clause.
	// Use LimitOffset or FetchNextLimit, or provide a custom implementation.
	AppendLimit func(b *strings.Builder, orderBy []string, limit, offset *int64)

	// QuoteStyle controls identifier quoting (backticks, double quotes, brackets).
	QuoteStyle QuoteStyle

	// BackslashEscapes enables handling of \' and \" inside string literals (MySQL).
	BackslashEscapes bool

	// AppendReturning appends a RETURNING clause after the VALUES clause (PostgreSQL, SQLite)
	// or after the WHERE clause in UPDATE/DELETE. Set to d.WriteReturning for PostgreSQL-style
	// dialects, d.WriteOutput for MSSQL, or nil for databases that don't support it (MySQL).
	AppendReturning func(b *strings.Builder, columns []string)

	// PrependReturning inserts a returning clause before the VALUES clause in INSERT.
	// Set to d.WriteOutput for MSSQL, or nil for other databases.
	PrependReturning func(b *strings.Builder, columns []string)
}

// SupportsReturning returns true when any returning hook is configured.
func (d *BaseDialect) SupportsReturning() bool {
	return d.AppendReturning != nil || d.PrependReturning != nil
}

// isQuestionMark returns true when the placeholder function produces bare "?".
func (d *BaseDialect) isQuestionMark() bool {
	return d.Placeholder(1) == "?"
}

// replaceAndCount replaces unquoted '?' characters with dialect-specific
// placeholders starting at startIdx. Returns the replaced string and the
// number of placeholders substituted.
func (d *BaseDialect) replaceAndCount(sql string, startIdx int) (string, int) {
	if d.isQuestionMark() {
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
			b.WriteString(d.Placeholder(idx))
			idx++
			count++
		default:
			b.WriteByte(ch)
		}
	}
	return b.String(), count
}

// countUnquoted counts '?' characters not inside quoted strings.
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

// buildClauses renders whereClauses into a joined string, expanding InValues.
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

// expandInPlaceholders replaces the first unquoted "?" with N placeholders.
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

func findFirstUnquotedPlaceholder(s string) int {
	inSingle, inDouble := false, false
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
// Skips quoting for expressions containing "(", "AS", spaces, commas, or "*".
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

func (d *BaseDialect) quoteColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	return strings.Join(quoted, ", ")
}

// --- RETURNING helpers ---

// WriteReturning appends " RETURNING col1, col2" using the dialect's identifier quoting.
// Use as: d.AppendReturning = d.WriteReturning
func (d *BaseDialect) WriteReturning(b *strings.Builder, columns []string) {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = d.QuoteIdentifier(c)
	}
	b.WriteString(" RETURNING ")
	b.WriteString(strings.Join(quoted, ", "))
}

// WriteOutput appends " OUTPUT INSERTED.col1, INSERTED.col2" using the dialect's quoting.
// Use as: d.PrependReturning = d.WriteOutput  (for INSERT)
//
//	or: d.AppendReturning = d.WriteOutput   (for UPDATE/DELETE)
func (d *BaseDialect) WriteOutput(b *strings.Builder, columns []string) {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = "INSERTED." + d.QuoteIdentifier(c)
	}
	b.WriteString(" OUTPUT ")
	b.WriteString(strings.Join(quoted, ", "))
}

// --- Build methods ---

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
		havingStr, _, newArgs, err := d.buildClauses(q.having, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		args = newArgs
		b.WriteString(" HAVING ")
		b.WriteString(havingStr)
	}

	if len(q.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(strings.Join(q.orderBy, ", "))
	}

	if q.limit != nil || q.offset != nil {
		d.AppendLimit(&b, q.orderBy, q.limit, q.offset)
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
	quoted := d.quoteColumns(q.keys)

	if len(q.rows) == 0 {
		placeholders := make([]string, len(q.keys))
		for i := range placeholders {
			placeholders[i] = d.Placeholder(i + 1)
		}

		b.WriteString("INSERT INTO ")
		b.WriteString(d.QuoteIdentifier(q.table))
		b.WriteString(" (")
		b.WriteString(quoted)
		b.WriteString(")")

		if len(q.returning) > 0 {
			if d.PrependReturning != nil {
				d.PrependReturning(&b, q.returning)
			}
		}

		b.WriteString(" VALUES (")
		b.WriteString(strings.Join(placeholders, ", "))
		b.WriteString(")")

		if len(q.returning) > 0 {
			if d.PrependReturning == nil && d.AppendReturning != nil {
				d.AppendReturning(&b, q.returning)
			}
		}

		return b.String(), q.values, nil
	}

	for _, row := range q.rows {
		if len(row) != len(q.keys) {
			return "", nil, ErrBatchRowLength
		}
	}

	var allPlaceholders []string
	var allArgs []interface{}
	for rowIdx, row := range q.rows {
		rowPhs := make([]string, len(q.keys))
		for i := range q.keys {
			rowPhs[i] = d.Placeholder(rowIdx*len(q.keys) + i + 1)
		}
		allPlaceholders = append(allPlaceholders, "("+strings.Join(rowPhs, ", ")+")")
		allArgs = append(allArgs, row...)
	}

	b.WriteString("INSERT INTO ")
	b.WriteString(d.QuoteIdentifier(q.table))
	b.WriteString(" (")
	b.WriteString(quoted)
	b.WriteString(")")

	if len(q.returning) > 0 {
		if d.PrependReturning != nil {
			d.PrependReturning(&b, q.returning)
		}
	}

	b.WriteString(" VALUES ")
	b.WriteString(strings.Join(allPlaceholders, ", "))

	if len(q.returning) > 0 {
		if d.PrependReturning == nil && d.AppendReturning != nil {
			d.AppendReturning(&b, q.returning)
		}
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
		setClauses[i] = fmt.Sprintf("%s = %s", d.QuoteIdentifier(key), d.Placeholder(i+1))
	}

	b.WriteString("UPDATE ")
	b.WriteString(d.QuoteIdentifier(q.table))
	b.WriteString(" SET ")
	b.WriteString(strings.Join(setClauses, ", "))

	args := make([]interface{}, len(q.values))
	copy(args, q.values)
	paramIdx := len(q.keys) + 1

	if len(q.wheres) > 0 {
		whereStr, _, newArgs, err := d.buildClauses(q.wheres, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		args = newArgs
		b.WriteString(" WHERE ")
		b.WriteString(whereStr)
	}

	if len(q.returning) > 0 {
		if len(q.returning) > 0 && d.AppendReturning != nil {
			d.AppendReturning(&b, q.returning)
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
		whereStr, _, newArgs, err := d.buildClauses(q.wheres, paramIdx, args)
		if err != nil {
			return "", nil, err
		}
		args = newArgs
		b.WriteString(" WHERE ")
		b.WriteString(whereStr)
	}

	if len(q.returning) > 0 {
		if len(q.returning) > 0 && d.AppendReturning != nil {
			d.AppendReturning(&b, q.returning)
		}
	}

	return b.String(), args, nil
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
			if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && ch != '_' {
				return fmt.Errorf("dal: invalid identifier %q", name)
			}
		} else {
			if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '_' {
				return fmt.Errorf("dal: invalid identifier %q", name)
			}
		}
	}
	return nil
}

// ErrNotImplemented is returned for features not yet implemented.
var ErrNotImplemented = errors.New("dal: feature not implemented")
