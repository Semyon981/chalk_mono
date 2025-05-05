package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, params CreateUserParams) (int64, error)
	GetHashedPassByEmail(ctx context.Context, email string) (string, error)
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

	var userID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO users (name) VALUES ($1) RETURNING id`,
		params.Name,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO credentials (user_id, email, hashpass) VALUES ($1, $2, $3)`,
		userID, params.Email, params.PasswordHashed,
	)
	if err != nil {
		return 0, fmt.Errorf("insert credentials: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return userID, nil
}

func (r *usersRepo) GetHashedPassByEmail(ctx context.Context, email string) (string, error) {
	hpass := ""
	err := r.db.QueryRow(ctx, "SELECT hashpass FROM Credentials WHERE email = $1", email).Scan(&hpass)
	if err != nil {
		return "", fmt.Errorf("failed to get pass: %w", err)
	}
	return hpass, err
}
