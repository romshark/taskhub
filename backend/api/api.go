package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/romshark/taskhub/api/graph"
	"github.com/romshark/taskhub/api/graph/auth"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/crypto/bcrypt"
)

func NewPasswordHasherBcrypt(cost int) *PasswordHasherBcrypt {
	switch {
	case cost == 0:
		cost = bcrypt.DefaultCost
	case cost < bcrypt.MinCost:
		cost = bcrypt.MinCost
	case cost > bcrypt.MaxCost:
		cost = bcrypt.MaxCost
	}
	return &PasswordHasherBcrypt{cost: cost}
}

// PasswordHasherBcrypt hashes passwords using scrypt
type PasswordHasherBcrypt struct{ cost int }

func (h *PasswordHasherBcrypt) HashPassword(plainText []byte) (hash string, err error) {
	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	b, err := bcrypt.GenerateFromPassword(plainText, bcrypt.DefaultCost)
	return string(b), err
}

func (h *PasswordHasherBcrypt) ComparePassword(
	plainText []byte, hash []byte,
) (ok bool, err error) {
	err = bcrypt.CompareHashAndPassword(hash, plainText)
	return err == nil, err
}

// TimeProviderLive provides live time.
type TimeProviderLive struct{}

func (p *TimeProviderLive) Now() time.Time { return time.Now() }

func NewServer(jwtSecret []byte, onResolver func(*graph.Resolver)) http.Handler {
	gqlResolver := &graph.Resolver{
		JWTGenerator:   auth.NewJWTGenerator(jwtSecret),
		PasswordHasher: NewPasswordHasherBcrypt(0),
		TimeProvider:   new(TimeProviderLive),
	}
	onResolver(gqlResolver)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(
		graph.Config{Resolvers: gqlResolver},
	))

	{
		// This middleware synchronizes concurrent access to the
		// in-memory state of the world of the resolver
		// and logs incoming operations.
		var lock sync.RWMutex
		srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			start := time.Now()
			requestContext := graphql.GetOperationContext(ctx)
			switch requestContext.Operation.Operation {
			case ast.Query:
				lock.RLock()
				defer lock.RUnlock()
			case ast.Mutation, ast.Subscription:
				lock.Lock()
				defer lock.Unlock()
			}
			defer func() {
				took := time.Since(start)
				log.Print("handled ", requestContext.Operation.Operation, " in ", took)
			}()
			return next(ctx)
		})
	}

	return auth.NewJWTMiddleware(jwtSecret)(srv)
}
