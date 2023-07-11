package graph

import (
	"context"
	"time"

	"github.com/romshark/taskhub/api/broadcast"
	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/graph/model"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DataProvider   dataprovider.DataProvider
	JWTGenerator   JWTGenerator
	PasswordHasher PasswordHasher
	TimeProvider   TimeProvider

	broadcastTaskUpsert    *broadcast.Broadcast[*model.Task]
	broadcastProjectUpsert *broadcast.Broadcast[*model.Project]
}

func NewResolver(
	dataProvider dataprovider.DataProvider,
	jWTGenerator JWTGenerator,
	passwordHasher PasswordHasher,
	timeProvider TimeProvider,
) *Resolver {
	return &Resolver{
		DataProvider:           dataProvider,
		JWTGenerator:           jWTGenerator,
		PasswordHasher:         passwordHasher,
		TimeProvider:           timeProvider,
		broadcastTaskUpsert:    broadcast.New[*model.Task](),
		broadcastProjectUpsert: broadcast.New[*model.Project](),
	}
}

type Subscriber[T any] interface {
	Notify(context.Context, T) error
	Subscribe(context.Context) (<-chan T, error)
}

type JWTGenerator interface {
	GenerateJWT(userID string, expiration time.Duration) (string, error)
}
type PasswordHasher interface {
	HashPassword(plainText []byte) (hash string, err error)
	ComparePassword(plainText, hash []byte) (ok bool, err error)
}

type TimeProvider interface {
	Now() time.Time
}
