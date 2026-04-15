package mysql

import "github.com/martinsuchenak/xdal/pkg/xdal"

// NewDialect returns a Dialect configured for MySQL.
// MySQL does not support RETURNING — neither hook is set.
func NewDialect() xdal.Dialect {
	return &xdal.BaseDialect{
		Placeholder:      xdal.QuestionMarkPlaceholder,
		AppendLimit:      xdal.LimitOffset,
		QuoteStyle:       xdal.BacktickQuoting,
		BackslashEscapes: true,
	}
}
