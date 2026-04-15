package postgres

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for PostgreSQL.
func NewDialect() dal.Dialect {
	d := &dal.BaseDialect{
		Placeholder: dal.DollarPlaceholder,
		AppendLimit: dal.LimitOffset,
		QuoteStyle:  dal.DoubleQuoteQuoting,
	}
	d.AppendReturning = d.WriteReturning
	return d
}
