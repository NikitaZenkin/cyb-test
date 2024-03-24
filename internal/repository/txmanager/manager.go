package txmanager

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type txKey struct{}

type Manager struct {
	db  *sqlx.DB
	log *zap.Logger
}

type Executor interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type TxBody = func(ctx context.Context) error

func New(db *sqlx.DB, log *zap.Logger) *Manager {
	return &Manager{
		db:  db,
		log: log,
	}
}

func (m *Manager) Executor(ctx context.Context) Executor {
	tx, ok := m.getTX(ctx)
	if ok {
		return tx
	}

	return m.db
}

func (m *Manager) WithTx(ctx context.Context, exec TxBody) error {
	_, alreadyWithTx := m.getTX(ctx)

	if alreadyWithTx {
		return exec(ctx)
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		m.log.Error("старт транзакции", zap.Error(err))
		return err
	}

	ctx = m.setTX(ctx, tx)

	var committed bool
	defer func(committed *bool) {
		if *committed {
			return
		}

		if rErr := tx.Rollback(); rErr != nil {
			m.log.Error("откат транзакции", zap.Error(rErr))
		}
	}(&committed)

	if err = exec(ctx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		m.log.Error("завершение транзакции", zap.Error(err))
		return err
	}

	committed = true
	return nil
}

func (m *Manager) getTX(ctx context.Context) (*sqlx.Tx, bool) {
	txCtx := ctx.Value(txKey{})
	if txCtx == nil {
		return nil, false
	}

	tx, ok := txCtx.(*sqlx.Tx)
	if !ok {
		return nil, false
	}

	return tx, true
}

func (m *Manager) setTX(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
