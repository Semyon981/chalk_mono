package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, params CreateUserParams) (int64, error)
}

func NewUsersRepo(db *pgx.Conn) UsersRepo {
	return &usersRepo{db: db}
}

type usersRepo struct {
	db *pgx.Conn
}

type CreateUserParams struct {
	Email          string
	PasswordHashed string
	Name           string
}

func (r *usersRepo) CreateUser(ctx context.Context, params CreateUserParams) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Вставка в users
	var userID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO users (name) VALUES ($1) RETURNING id`,
		params.Name,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}

	// 2. Вставка в credentials
	_, err = tx.Exec(ctx,
		`INSERT INTO credentials (user_id, email, hashpass) VALUES ($1, $2, $3)`,
		userID, params.Email, params.PasswordHashed,
	)
	if err != nil {
		return 0, fmt.Errorf("insert credentials: %w", err)
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return userID, nil
}
