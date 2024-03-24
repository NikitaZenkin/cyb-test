package repository

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"cyb-test/internal/repository/txmanager"
)

type Repository struct {
	*txmanager.Manager
	log *zap.Logger
}

func New(db *sqlx.DB, log *zap.Logger) *Repository {
	return &Repository{
		Manager: txmanager.New(db, log),
		log:     log,
	}
}
