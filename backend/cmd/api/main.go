package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/romshark/taskhub/api"
	"github.com/romshark/taskhub/api/dataprovider/inmem"
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

	inmemDataProvider := inmem.NewFake()

	apiServer := api.NewServer([]byte(envJWTSecret), inmemDataProvider)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", apiServer)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", envPort)
	log.Fatal(http.ListenAndServe(":"+envPort, nil))
}
