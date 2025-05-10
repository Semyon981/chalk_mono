package usecases

import (
	"chalk/internal/entities"
	"chalk/internal/repo"
	"context"
	"errors"
	"fmt"

	uerrors "chalk/internal/errors"
)

type UsersUseCase interface {
}

func NewUsersUseCase(ur repo.UsersRepo) UsersUseCase {
	return &usersUseCase{ur: ur}
}

type usersUseCase struct {
	ur repo.UsersRepo
}

func (uc *usersUseCase) GetUserById(ctx context.Context, id int64) (entities.User, error) {
	user, err := uc.ur.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrRecordNotFound) {
			return entities.User{}, uerrors.ErrUserNotFound
		}
		return entities.User{}, fmt.Errorf("failed to get user by id: %w", err)
	}
	return entities.User{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}
