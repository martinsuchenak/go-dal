package sqlite

import (
	"fmt"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

type sqliteDialect struct {
	*dal.BaseDialect
}

func (d *sqliteDialect) CurrentTimestamp() string { return "datetime('now')" }

func (d *sqliteDialect) BoolLiteral(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func (d *sqliteDialect) StringAggExpr(col, sep string) string {
	return fmt.Sprintf("GROUP_CONCAT(%s, %s)", col, sep)
}

func (d *sqliteDialect) RandExpr() string { return "RANDOM()" }

func NewDialect() dal.Dialect {
	b := &dal.BaseDialect{
		Placeholder: dal.QuestionMarkPlaceholder,
		AppendLimit: dal.LimitOffset,
		QuoteStyle:  dal.DoubleQuoteQuoting,
	}
	b.AppendReturning = b.WriteReturning
	return &sqliteDialect{BaseDialect: b}
}
