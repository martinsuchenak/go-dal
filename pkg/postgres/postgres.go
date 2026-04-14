package postgres

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

var _ dal.DBInterface = (*PostgresDB)(nil)

type PostgresDB struct {
	*dal.BaseDB
}

func NewPostgresDB(db *sql.DB, log ...dal.Logger) *PostgresDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &PostgresDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.BaseDB.Exec(ctx, query, args...)
}

func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.BaseDB.Query(ctx, query, args...)
}

func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.BaseDB.QueryRow(ctx, query, args...)
}

func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.BaseDB.BeginTx(ctx, opts)
}

func (p *PostgresDB) Close() error {
	return p.BaseDB.Close()
}

func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilderWithStyle(dal.DollarNumber)
}
