package mssql

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for SQL Server.
func NewDialect() dal.Dialect {
	return &dal.BaseDialect{
		PlaceholderStyle: dal.AtPNumber,
		LimitStyle:       dal.FetchNextStyle,
	}
}
