package service

import (
	"context"
	"errors"
)

type UserChecker interface {
	UserExists(ctx context.Context, userID uint64) (bool, error)
}

var ErrUserNotFound = errors.New("ユーザーが見つかりませんでした。s")
