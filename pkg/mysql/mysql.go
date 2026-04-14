package mysql

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

var _ dal.DBInterface = (*MySQLDB)(nil)

type MySQLDB struct {
	*dal.BaseDB
}

func NewMySQLDB(db *sql.DB, log ...dal.Logger) *MySQLDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &MySQLDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (m *MySQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.BaseDB.Exec(ctx, query, args...)
}

func (m *MySQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.BaseDB.Query(ctx, query, args...)
}

func (m *MySQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.BaseDB.QueryRow(ctx, query, args...)
}

func (m *MySQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.BaseDB.BeginTx(ctx, opts)
}

func (m *MySQLDB) Close() error {
	return m.BaseDB.Close()
}

func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder()
}
