package repo

import (
	"chalk/internal/repo/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type SessionsRepo interface {
	CreateSession(ctx context.Context, params CreateSessionParams) (int64, error)
	GetSessionByAccessToken(ctx context.Context, accessToken string) (models.Session, error)
	UpdateSession(ctx context.Context, params UpdateSessionParams) error
	UpdateSessionWithRefreshToken(ctx context.Context, params UpdateSessionWithRefreshTokenParams) (int64, error)
}

func NewSessionsRepo(db *pgx.Conn) SessionsRepo {
	return &sessionsRepo{db: db}
}

type sessionsRepo struct {
	db *pgx.Conn
}

type CreateSessionParams struct {
	UserID         int64
	AccessToken    string
	RefreshToken   string
	AccessExpires  time.Time
	RefreshExpires time.Time
	Issued         time.Time
}

func (r *sessionsRepo) CreateSession(ctx context.Context, params CreateSessionParams) (int64, error) {
	var sessionID int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO sessions (user_id, access_token, refresh_token, access_expires, refresh_expires, issued) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		params.UserID, params.AccessToken, params.RefreshToken, params.AccessExpires.UTC(), params.RefreshExpires.UTC(), params.Issued.UTC(),
	).Scan(&sessionID)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}
	return sessionID, nil
}

func (r *sessionsRepo) GetSessionByAccessToken(ctx context.Context, accessToken string) (models.Session, error) {
	session := models.Session{}
	err := r.db.QueryRow(ctx,
		"SELECT id, user_id, access_token, refresh_token, access_expires, refresh_expires, issued FROM sessions WHERE access_token = $1",
		accessToken).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.AccessExpires,
		&session.RefreshExpires,
		&session.Issued)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, ErrRecordNotFound
		}
		return models.Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

type UpdateSessionParams struct {
	ID                int64
	NewAccessToken    string
	NewRefreshToken   string
	NewAccessExpires  time.Time
	NewRefreshExpires time.Time
}

func (r *sessionsRepo) UpdateSession(ctx context.Context, params UpdateSessionParams) error {
	_, err := r.db.Exec(ctx,
		"UPDATE sessions SET (access_token, refresh_token, access_expires, refresh_expires) = ($1, $2, $3, $4) WHERE id = $5",
		params.NewAccessToken, params.NewRefreshToken, params.NewAccessExpires, params.NewRefreshExpires, params.ID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

type UpdateSessionWithRefreshTokenParams struct {
	RefreshToken      string
	NewAccessToken    string
	NewRefreshToken   string
	NewAccessExpires  time.Time
	NewRefreshExpires time.Time
}

func (r *sessionsRepo) UpdateSessionWithRefreshToken(ctx context.Context, params UpdateSessionWithRefreshTokenParams) (int64, error) {
	var sessionID int64
	err := r.db.QueryRow(ctx,
		`UPDATE sessions SET 
		(access_token, refresh_token, access_expires, refresh_expires) = ($1, $2, $3, $4) 
		WHERE refresh_token = $5 AND refresh_expires > $6 RETURNING id`,
		params.NewAccessToken,
		params.NewRefreshToken,
		params.NewAccessExpires.UTC(),
		params.NewRefreshExpires.UTC(),
		params.RefreshToken,
		time.Now().UTC(),
	).Scan(&sessionID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrRecordNotFound
		}
		return 0, fmt.Errorf("failed to update session: %w", err)
	}
	return sessionID, nil
}
