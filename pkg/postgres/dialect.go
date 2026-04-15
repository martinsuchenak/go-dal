package postgres

import (
	"fmt"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

type postgresDialect struct {
	*dal.BaseDialect
}

func (d *postgresDialect) StringAggExpr(col, sep string) string {
	return fmt.Sprintf("STRING_AGG(%s, %s)", col, sep)
}

func (d *postgresDialect) RandExpr() string { return "RANDOM()" }

func NewDialect() dal.Dialect {
	b := &dal.BaseDialect{
		Placeholder: dal.DollarPlaceholder,
		AppendLimit: dal.LimitOffset,
		QuoteStyle:  dal.DoubleQuoteQuoting,
	}
	b.AppendReturning = b.WriteReturning
	return &postgresDialect{BaseDialect: b}
}
