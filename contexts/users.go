package contexts

import (
	"context"
	"models"
)

type key string

const (
	userKey key = "user"
)

// store user into the context
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// retrieve a user from the context
func User(ctx context.Context) *models.User {
	if user, ok := ctx.Value(userKey).(*models.User); !ok {
		return nil
	} else {
		return user
	}
}
