package mssql

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

var _ dal.DBInterface = (*MSSQLDB)(nil)

type MSSQLDB struct {
	*dal.BaseDB
}

func NewMSSQLDB(db *sql.DB, log ...dal.Logger) *MSSQLDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &MSSQLDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (m *MSSQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.BaseDB.Exec(ctx, query, args...)
}

func (m *MSSQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.BaseDB.Query(ctx, query, args...)
}

func (m *MSSQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.BaseDB.QueryRow(ctx, query, args...)
}

func (m *MSSQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.BaseDB.BeginTx(ctx, opts)
}

func (m *MSSQLDB) Close() error {
	return m.BaseDB.Close()
}

func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilderWithStyle(dal.AtPNumber)
}
