package graph

import (
	"time"

	"github.com/romshark/taskhub/api/graph/model"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Users          []*model.User
	Tasks          []*model.Task
	Projects       []*model.Project
	JWTGenerator   JWTGenerator
	PasswordHasher PasswordHasher
	TimeProvider   TimeProvider
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
