package mssql

import "github.com/martinsuchenak/go-dal/pkg/dal"

// NewDialect returns a Dialect configured for SQL Server.
// MSSQL uses OUTPUT (before VALUES for INSERT, after WHERE for UPDATE/DELETE).
func NewDialect() dal.Dialect {
	d := &dal.BaseDialect{
		Placeholder: dal.AtPPlaceholder,
		AppendLimit: dal.FetchNextLimit,
		QuoteStyle:  dal.BracketQuoting,
	}
	d.PrependReturning = d.WriteOutput
	d.AppendReturning = d.WriteOutput
	return d
}
