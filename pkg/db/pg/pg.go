package pg

import (
	"context"
	"fmt"
	"log"

	"github.com/Dnlbb/platform_common/pkg/db"
	"github.com/Dnlbb/platform_common/pkg/db/prettier"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type key string

// TxKey ключ транзакции.
const (
	TxKey key = "tx"
)

type pg struct {
	dbc *pgxpool.Pool
}

// NewDB конструктор для базы
func NewDB(dbc *pgxpool.Pool) db.DB {
	return &pg{
		dbc: dbc,
	}
}

func (p *pg) ScanOneContext(ctx context.Context, dest interface{}, query db.Query, args ...interface{}) error {
	logQuery(ctx, query, args...)

	rows, err := p.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, rows)
}

func (p *pg) QueryContext(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	logQuery(ctx, q, args...)

	tx, err := ctx.Value(TxKey).(pgx.Tx)
	if err {
		return tx.Query(ctx, q.QueryRow, args...)
	}

	return p.dbc.Query(ctx, q.QueryRow, args...)
}

func (p *pg) ScanAllContext(ctx context.Context, dest interface{}, query db.Query, args ...interface{}) error {
	logQuery(ctx, query, args...)

	rows, err := p.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (p *pg) ExecContext(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	logQuery(ctx, q, args...)

	tx, err := ctx.Value(TxKey).(pgx.Tx)
	if err {
		return tx.Exec(ctx, q.QueryRow, args...)
	}

	return p.dbc.Exec(ctx, q.QueryRow, args...)
}

func (p *pg) QueryRowContext(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	logQuery(ctx, q, args...)

	tx, err := ctx.Value(TxKey).(pgx.Tx)
	if err {
		return tx.QueryRow(ctx, q.QueryRow, args...)
	}

	return p.dbc.QueryRow(ctx, q.QueryRow, args...)
}

func logQuery(_ context.Context, q db.Query, args ...interface{}) {
	prettyQuery := prettier.Pretty(q.QueryRow, "$", args...)

	log.Printf(fmt.Sprintf("sql: %s", q.Name),
		fmt.Sprintf("query: %s", prettyQuery),
	)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.dbc.Ping(ctx)
}

func (p *pg) Close() {
	p.dbc.Close()
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.dbc.BeginTx(ctx, txOptions)
}

// MakeContextTx добавляем в контекст ключ транзакций
func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}
