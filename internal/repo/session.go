package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type SessionsRepo interface {
}

func NewSessionsRepo(db *pgx.Conn) SessionsRepo {
	return &sessionsRepo{db: db}
}

type sessionsRepo struct {
	db *pgx.Conn
}

func (r *sessionsRepo) Create(ctx context.Context) {

}
