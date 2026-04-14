package mysql

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for MySQL.
func NewDialect() dal.Dialect {
	return &dal.BaseDialect{
		PlaceholderStyle: dal.QuestionMark,
		LimitStyle:       dal.LimitOffsetStyle,
	}
}
