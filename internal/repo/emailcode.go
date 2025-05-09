package repo

import (
	"chalk/internal/repo/models"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const authCodeRepoPrefix string = "authcode"

type emailCodeRepo struct {
	rdb *redis.Client
}

type EmailCodeRepo interface {
	Set(ctx context.Context, codeID string, code models.EmailCode, ttl time.Duration) error
	Get(ctx context.Context, codeID string) (models.EmailCode, error)
	Delete(ctx context.Context, codeID string) (bool, error)
}

func NewAuthCodeRepo(rdb *redis.Client) EmailCodeRepo {
	return &emailCodeRepo{rdb: rdb}
}

func (r *emailCodeRepo) Set(ctx context.Context, codeID string, code models.EmailCode, ttl time.Duration) error {
	err := r.rdb.Set(ctx, fmt.Sprintf("%s:%s", authCodeRepoPrefix, codeID), code, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

func (r *emailCodeRepo) Get(ctx context.Context, codeID string) (models.EmailCode, error) {
	code := models.EmailCode{}
	err := r.rdb.Get(ctx, fmt.Sprintf("%s:%s", authCodeRepoPrefix, codeID)).Scan(&code)
	if err != nil {
		if err == redis.Nil {
			return models.EmailCode{}, ErrRecordNotFound
		}
		return models.EmailCode{}, fmt.Errorf("failed to get value: %w", err)
	}
	return code, nil
}

func (r *emailCodeRepo) Delete(ctx context.Context, codeID string) (bool, error) {
	res, err := r.rdb.Del(ctx, fmt.Sprintf("%s:%s", authCodeRepoPrefix, codeID)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete value: %w", err)
	}
	return res > 0, nil
}
