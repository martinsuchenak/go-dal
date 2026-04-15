package postgres

import (
	"fmt"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

type postgresDialect struct {
	*xdal.BaseDialect
}

func (d *postgresDialect) StringAggExpr(col, sep string) string {
	return fmt.Sprintf("STRING_AGG(%s, %s)", col, sep)
}

func (d *postgresDialect) RandExpr() string { return "RANDOM()" }

func NewDialect() xdal.Dialect {
	b := &xdal.BaseDialect{
		Placeholder: xdal.DollarPlaceholder,
		AppendLimit: xdal.LimitOffset,
		QuoteStyle:  xdal.DoubleQuoteQuoting,
	}
	b.AppendReturning = b.WriteReturning
	return &postgresDialect{BaseDialect: b}
}
