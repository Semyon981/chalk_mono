package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const prefix string = "authcode"

type AuthCodeRepo interface {
	Set(ctx context.Context, codeID string, code VerificationCode, ttl time.Duration) error
	Get(ctx context.Context, codeID string) (VerificationCode, error)
	Delete(ctx context.Context, codeID string) (bool, error)
}

func NewAuthCodeRepo(rdb *redis.Client) AuthCodeRepo {
	return &authCodeRepo{rdb: rdb}
}

type authCodeRepo struct {
	rdb *redis.Client
}

type VerificationCode struct {
	Email string
	Code  string
}

func (u VerificationCode) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *VerificationCode) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func (r *authCodeRepo) Set(ctx context.Context, codeID string, code VerificationCode, ttl time.Duration) error {
	err := r.rdb.Set(ctx, fmt.Sprintf("%s:%s", prefix, codeID), code, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

func (r *authCodeRepo) Get(ctx context.Context, codeID string) (VerificationCode, error) {
	code := VerificationCode{}
	err := r.rdb.Get(ctx, fmt.Sprintf("%s:%s", prefix, codeID)).Scan(&code)
	if err != nil {
		if err == redis.Nil {
			return VerificationCode{}, fmt.Errorf("record not found: %w", err)
		}
		return VerificationCode{}, fmt.Errorf("failed to get value: %w", err)
	}
	return code, nil
}

func (r *authCodeRepo) Delete(ctx context.Context, codeID string) (bool, error) {
	res, err := r.rdb.Del(ctx, fmt.Sprintf("%s:%s", prefix, codeID)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete value: %w", err)
	}
	return res > 0, nil
}
