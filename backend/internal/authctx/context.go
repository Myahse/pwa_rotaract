package authctx

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
)

type contextKey string

const userKey contextKey = "auth_user"

var ErrUnauthorized = errors.New("unauthorized")

func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userKey).(*domain.User)
	return user, ok && user != nil
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}
	return user.ID, nil
}
