package postgres

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for PostgreSQL.
func NewDialect() dal.Dialect {
	return &dal.BaseDialect{
		PlaceholderStyle: dal.DollarNumber,
		LimitStyle:       dal.LimitOffsetStyle,
	}
}
