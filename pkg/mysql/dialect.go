package mysql

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for MySQL.
// MySQL does not support RETURNING — neither hook is set.
func NewDialect() dal.Dialect {
	return &dal.BaseDialect{
		Placeholder:      dal.QuestionMarkPlaceholder,
		AppendLimit:      dal.LimitOffset,
		QuoteStyle:       dal.BacktickQuoting,
		BackslashEscapes: true,
	}
}
