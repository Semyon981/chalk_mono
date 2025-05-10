package repo

import (
	"chalk/internal/repo/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountsRepo interface {
}

func NewAccountsRepo(db *pgx.Conn) AccountsRepo {
	return &accountsRepo{db: db}
}

type accountsRepo struct {
	db *pgx.Conn
}

func (r *accountsRepo) GetAccountByID(ctx context.Context, id int64) (models.Account, error) {
	const query = `SELECT id, name FROM accounts WHERE id = $1`
	acc := models.Account{}
	err := r.db.QueryRow(ctx, query, id).Scan(&acc.ID, acc.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, ErrAccountNotFound
		}
		return models.Account{}, fmt.Errorf("select account: %w", err)
	}
	return acc, nil
}

func (r *accountsRepo) GetAccountByName(ctx context.Context, name string) (models.Account, error) {
	const query = `SELECT id, name FROM accounts WHERE name = $1`
	acc := models.Account{}
	err := r.db.QueryRow(ctx, query, name).Scan(&acc.ID, acc.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, ErrAccountNotFound
		}
		return models.Account{}, fmt.Errorf("select account: %w", err)
	}
	return acc, nil
}

type CreateAccountParams struct {
	OwnerUserID int64
	AccountName string
}

func (r *accountsRepo) CreateAccount(ctx context.Context, params CreateAccountParams) (models.Account, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Account{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var accountID int64
	err = tx.QueryRow(ctx, "INSERT INTO accounts (name) VALUES ($1) RETURNING id", params.AccountName).Scan(&accountID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return models.Account{}, ErrAccountNameAlreadyTaken
			}
		}
		return models.Account{}, fmt.Errorf("insert account: %w", err)
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO account_members (user_id, account_id, role) VALUES ($1, $2, $3)",
		params.OwnerUserID,
		accountID,
		models.AccountMembersRoleOwner,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return models.Account{}, ErrAccountOwnerNotFound
			}
		}
		return models.Account{}, fmt.Errorf("insert account_members: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Account{}, fmt.Errorf("commit: %w", err)
	}

	return models.Account{ID: accountID, Name: params.AccountName}, nil
}

type GetAccountMembersParams struct {
	AccountID int64
}

func (r *accountsRepo) GetAccountMembers(ctx context.Context, params GetAccountMembersParams) ([]models.AccountMember, error) {

	_, err := r.GetAccountByID(ctx, params.AccountID)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("check acc exists: %w", err)
	}

	const query = `
	SELECT am.id, am.user_id, am.account_id, am.role, u.name, u.email
	FROM account_members am
	INNER JOIN users u ON u.id = am.user_id
	WHERE am.account_id = $1`

	rows, err := r.db.Query(ctx, query, params.AccountID)
	if err != nil {
		return nil, fmt.Errorf("query account members: %w", err)
	}
	defer rows.Close()

	var members []models.AccountMember
	for rows.Next() {
		var m models.AccountMember
		err := rows.Scan(
			&m.UserID,
			&m.AccountID,
			&m.Role,
			&m.Name,
			&m.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("scan account member: %w", err)
		}
		members = append(members, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return members, nil
}

type GetUserAccountsParams struct {
	UserID int64
}

func (r *accountsRepo) GetUserAccounts(ctx context.Context, params GetUserAccountsParams) ([]models.Account, error) {
	const checkQuery = `SELECT 1 FROM users WHERE id = $1`
	var exists int
	err := r.db.QueryRow(ctx, checkQuery, params.UserID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("checking user existence: %w", err)
	}

	const query = `
	SELECT a.id, a.name
	FROM accounts a
	INNER JOIN account_members am ON a.id = am.account_id
	WHERE am.user_id = $1
`
	rows, err := r.db.Query(ctx, query, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("query user accounts: %w", err)
	}
	defer rows.Close()

	var result []models.Account
	for rows.Next() {
		var m models.Account
		err := rows.Scan(
			&m.ID,
			&m.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user account: %w", err)
		}
		result = append(result, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}

type AddAccountMemberParams struct {
	UserID    int64
	AccountID int64
	Role      models.AccountMembersRole
}

func (r *accountsRepo) AddAccountMember(ctx context.Context, params AddAccountMemberParams) error {
	const query = `INSERT INTO account_members (user_id, account_id, role) VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, query, params.UserID, params.AccountID, params.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if pgErr.ColumnName == "user_id" {
					err = ErrUserNotFound
				} else if pgErr.ColumnName == "account_id" {
					err = ErrAccountNotFound
				}
				return err
			}
		}
		return fmt.Errorf("insert account_members: %w", err)
	}
	return nil
}

type RemoveAccountMemberParams struct {
	AccountID int64
	UserID    int64
}

func (r *accountsRepo) RemoveAccountMember(ctx context.Context, params RemoveAccountMemberParams) error {
	const query = `DELETE FROM account_members WHERE user_id = $1 AND account_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, params.UserID, params.AccountID)
	if err != nil {
		return fmt.Errorf("delete account member: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrAccountMemberNotFound
	}
	return nil
}

type UpdateAccountMemberRoleParams struct {
	AccountID int64
	UserID    int64
	Role      models.AccountMembersRole
}

func (r *accountsRepo) UpdateAccountMemberRole(ctx context.Context, params UpdateAccountMemberRoleParams) error {
	const query = `UPDATE account_members SET role = $1 WHERE user_id = $2 AND account_id = $3`
	cmdTag, err := r.db.Exec(ctx, query, params.Role, params.UserID, params.AccountID)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrAccountMemberNotFound
	}
	return nil
}
