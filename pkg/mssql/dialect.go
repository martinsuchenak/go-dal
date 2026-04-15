package mssql

import (
	"fmt"
	"strings"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

type mssqlDialect struct {
	*xdal.BaseDialect
}

func (d *mssqlDialect) ConcatExpr(parts ...string) string {
	return strings.Join(parts, " + ")
}

func (d *mssqlDialect) LengthExpr(col string) string {
	return "LEN(" + col + ")"
}

func (d *mssqlDialect) CurrentTimestamp() string { return "GETDATE()" }

func (d *mssqlDialect) BoolLiteral(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func (d *mssqlDialect) StringAggExpr(col, sep string) string {
	return fmt.Sprintf("STRING_AGG(%s, %s)", col, sep)
}

func (d *mssqlDialect) RandExpr() string { return "RAND()" }

func NewDialect() xdal.Dialect {
	b := &xdal.BaseDialect{
		Placeholder: xdal.AtPPlaceholder,
		AppendLimit: xdal.FetchNextLimit,
		QuoteStyle:  xdal.BracketQuoting,
	}
	b.PrependReturning = b.WriteOutput
	b.AppendReturning = b.WriteOutput
	b.AppendDeletedReturning = b.WriteDeletedOutput
	return &mssqlDialect{BaseDialect: b}
}
