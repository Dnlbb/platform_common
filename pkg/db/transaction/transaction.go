package transaction

import (
	"context"
	"fmt"

	"github.com/Dnlbb/platform_common/pkg/db"
	"github.com/Dnlbb/platform_common/pkg/db/pg"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type manager struct {
	db db.Transactor
}

// NewTransactionManager конструктор менеджера транзакций.
func NewTransactionManager(db db.Transactor) db.TxManager {
	return &manager{db: db}
}

func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn db.Handler) (err error) {
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("can't start transaction: %w", err)
	}

	ctx = pg.MakeContextTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}
			return
		}

		if errors.Is(err, nil) {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "tx commit failed")
			}
		}
	}()

	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err

}

func (m *manager) ReadCommitted(ctx context.Context, f db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}

	return m.transaction(ctx, txOpts, f)
}
