package sqlite

import (
	"fmt"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

type sqliteDialect struct {
	*xdal.BaseDialect
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

func NewDialect() xdal.Dialect {
	b := &xdal.BaseDialect{
		Placeholder: xdal.QuestionMarkPlaceholder,
		AppendLimit: xdal.LimitOffset,
		QuoteStyle:  xdal.DoubleQuoteQuoting,
	}
	b.AppendReturning = b.WriteReturning
	return &sqliteDialect{BaseDialect: b}
}
