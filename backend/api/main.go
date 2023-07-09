package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/romshark/taskhub/graph"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	gqlResolver := new(graph.Resolver)
	makeData(gqlResolver)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: gqlResolver}))
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

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
