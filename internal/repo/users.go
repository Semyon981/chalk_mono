package repo

import (
	"chalk/internal/repo/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, params CreateUserParams) (int64, error)
	GetUserById(ctx context.Context, id int64) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
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
	var userID int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, hashpass) VALUES ($1, $2, $3) RETURNING id`,
		params.Name, params.Email, params.PasswordHashed,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrUniqueViolation
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return userID, nil
}

func (r *usersRepo) GetUserById(ctx context.Context, id int64) (models.User, error) {
	u := models.User{}
	err := r.db.QueryRow(ctx, "SELECT id, name, email, hashpass FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrRecordNotFound
		}
		return models.User{}, fmt.Errorf("failed to get pass: %w", err)
	}
	return u, nil
}

func (r *usersRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	u := models.User{}
	err := r.db.QueryRow(ctx, "SELECT id, name, email, hashpass FROM users WHERE email = $1", email).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrRecordNotFound
		}
		return models.User{}, fmt.Errorf("failed to get pass: %w", err)
	}
	return u, nil
}
