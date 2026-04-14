package sqlite

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

var _ dal.DBInterface = (*SQLiteDB)(nil)

type SQLiteDB struct {
	*dal.BaseDB
}

func NewSQLiteDB(db *sql.DB, log ...dal.Logger) *SQLiteDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &SQLiteDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (s *SQLiteDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.BaseDB.Exec(ctx, query, args...)
}

func (s *SQLiteDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.BaseDB.Query(ctx, query, args...)
}

func (s *SQLiteDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.BaseDB.QueryRow(ctx, query, args...)
}

func (s *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.BaseDB.BeginTx(ctx, opts)
}

func (s *SQLiteDB) Close() error {
	return s.BaseDB.Close()
}

func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder()
}
