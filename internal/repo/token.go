package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenInfo struct {
	UserID    int64
	SessionID int64
}

func (u TokenInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *TokenInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

const tokenRepoPrefix string = "token"

type TokenRepo interface {
}

func NewTokenRepo(rdb *redis.Client) TokenRepo {
	return &tokenRepo{rdb: rdb}
}

type tokenRepo struct {
	rdb *redis.Client
}

func (r *tokenRepo) Set(ctx context.Context, token string, info TokenInfo, TTL time.Duration) error {
	err := r.rdb.Set(ctx, fmt.Sprintf("%s:%s", tokenRepoPrefix, token), info, TTL).Err()
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func (r *tokenRepo) Get(ctx context.Context, token string) (TokenInfo, error) {
	info := TokenInfo{}
	err := r.rdb.Get(ctx, fmt.Sprintf("%s:%s", tokenRepoPrefix, token)).Scan(&info)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("failed to get token: %w", err)
	}

	return TokenInfo{}, nil
}

func (r *tokenRepo) Delete(ctx context.Context, token string) (bool, error) {
	res, err := r.rdb.Del(ctx, fmt.Sprintf("%s:%s", tokenRepoPrefix, token)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete token: %w", err)
	}
	return res > 0, nil
}
