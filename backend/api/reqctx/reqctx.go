package reqctx

import (
	"context"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	"golang.org/x/exp/slog"
)

type ctxKey int8

var ctxKeyRequestContext ctxKey = 1

func WithRequestContext(
	ctx context.Context,
	log *slog.Logger,
	userID string,
	persistedQueryName string,
	start time.Time,
) context.Context {
	now := time.Now()
	entropy := rand.New(rand.NewSource(now.UnixNano()))
	ms := ulid.Timestamp(now)
	return context.WithValue(ctx, ctxKeyRequestContext, &RequestContext{
		RequestID:          ulid.MustNew(ms, entropy).String(),
		UserID:             userID,
		Log:                log,
		PersistedQueryName: persistedQueryName,
		Start:              start,
	})
}

func GetRequestContext(ctx context.Context) *RequestContext {
	c := ctx.Value(ctxKeyRequestContext)
	if c != nil {
		return c.(*RequestContext)
	}
	return nil
}

type RequestContext struct {
	RequestID          string
	UserID             string
	Log                *slog.Logger
	PersistedQueryName string
	Start              time.Time
}
