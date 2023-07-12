package auth

import (
	"context"
	"errors"

	"github.com/romshark/taskhub/api/reqctx"
)

// RequireAuthenticated returns nil if the client is authenticated,
// otherwise returns ErrUnauthenticated.
func RequireAuthenticated(ctx context.Context) error {
	c := reqctx.GetRequestContext(ctx)
	if c.UserID == "" {
		return ErrUnauthenticated
	}
	return nil
}

// RequireOwner returns nil if the client is authenticated and
// is also the owner of the resource, otherwise returns ErrUnauthenticated.
func RequireOwner(ctx context.Context, ownerID string) error {
	c := reqctx.GetRequestContext(ctx)
	if c.UserID == "" {
		return ErrUnauthenticated
	}
	if c.UserID != ownerID {
		return ErrUnauthorized
	}
	return nil
}

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")
)
