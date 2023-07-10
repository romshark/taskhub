package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/romshark/taskhub/api"
	"github.com/romshark/taskhub/api/graph"
)

const defaultPort = "8080"

func main() {
	envPort := os.Getenv("PORT")
	if envPort == "" {
		envPort = defaultPort
	}
	envJWTSecret := os.Getenv("JWT_SECRET")
	if envJWTSecret == "" {
		log.Print("no JWT secret provided")
		return
	}

	apiServer := api.NewServer([]byte(envJWTSecret), func(r *graph.Resolver) {
		makeData(r)
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", apiServer)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", envPort)
	log.Fatal(http.ListenAndServe(":"+envPort, nil))
}
