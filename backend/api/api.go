// Package api provides the GraphQL API server.
package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/graph"
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

func NewServer(jwtSecret []byte, dataProvider dataprovider.DataProvider) http.Handler {
	gqlResolver := graph.NewResolver(
		dataProvider,
		auth.NewJWTGenerator(jwtSecret),
		NewPasswordHasherBcrypt(0),
		new(TimeProviderLive),
	)
	conf := graph.Config{Resolvers: gqlResolver}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(conf))

	srv.AddTransport(&transport.Websocket{})

	// Define middleware to log incoming operations.
	srv.AroundOperations(func(
		ctx context.Context, next graphql.OperationHandler,
	) graphql.ResponseHandler {
		start := time.Now()
		requestContext := graphql.GetOperationContext(ctx)
		defer func() {
			took := time.Since(start)
			log.Print("handled ", requestContext.Operation.Operation, " in ", took)
		}()
		return next(ctx)
	})

	return auth.NewJWTMiddleware(jwtSecret)(srv)
}
