package mysql

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for MySQL.
// MySQL uses backslash escapes by default (e.g., \' inside strings).
func NewDialect() dal.Dialect {
	return &dal.BaseDialect{
		PlaceholderStyle: dal.QuestionMark,
		LimitStyle:       dal.LimitOffsetStyle,
		QuoteStyle:       dal.BacktickQuoting,
		BackslashEscapes: true,
	}
}
