package sqlite

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for SQLite.
func NewDialect() dal.Dialect {
	d := &dal.BaseDialect{
		Placeholder: dal.QuestionMarkPlaceholder,
		AppendLimit: dal.LimitOffset,
		QuoteStyle:  dal.DoubleQuoteQuoting,
	}
	d.AppendReturning = d.WriteReturning
	return d
}
