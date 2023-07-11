// Package api provides the GraphQL API server.
package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/graph"
	"github.com/romshark/taskhub/api/passhash"
)

// TimeProviderLive provides live time.
type TimeProviderLive struct{}

func (p *TimeProviderLive) Now() time.Time { return time.Now() }

func NewServer(
	jwtSecret []byte,
	dataProvider dataprovider.DataProvider,
) http.Handler {
	gqlResolver := graph.NewResolver(
		dataProvider,
		auth.NewJWTGenerator(jwtSecret),
		passhash.NewPasswordHasherBcrypt(0),
		new(TimeProviderLive),
	)
	conf := graph.Config{Resolvers: gqlResolver}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(conf))
	srv.AddTransport(&transport.Websocket{})
	srv.AroundResponses(gqlMiddlewareLogResponses)

	withJWT := auth.NewJWTMiddleware(jwtSecret)(srv)
	withStartTime := newMiddlewareSetStartTime(withJWT)
	return withStartTime
}

func gqlMiddlewareLogResponses(
	ctx context.Context, next graphql.ResponseHandler,
) *graphql.Response {
	requestContext := graphql.GetOperationContext(ctx)
	start := ctx.Value(ctxKeyStartTime).(time.Time)
	took := time.Since(start)
	log.Print(
		"handled ",
		requestContext.Operation.Operation,
		" in ",
		took,
	)
	return next(ctx)
}

func newMiddlewareSetStartTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyStartTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ctxKey int8

var ctxKeyStartTime = 1
