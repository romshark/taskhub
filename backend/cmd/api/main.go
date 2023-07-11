package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/romshark/taskhub/api"
	"github.com/romshark/taskhub/api/dataprovider/inmem"
)

const defaultHost = "localhost:8080"

func main() {
	envHost := os.Getenv("HOST")
	if envHost == "" {
		envHost = defaultHost
	}

	envJWTSecret := os.Getenv("JWT_SECRET")
	if envJWTSecret == "" {
		log.Print("missing JWT secret")
		return
	}

	envGQLWhitelistPath := os.Getenv("GRAPHQL_WHITELIST_PATH")
	if envGQLWhitelistPath == "" {
		log.Print("missing GraphQL whitelist path")
		return
	}

	mode := api.ModeDebug
	envMode := strings.ToLower(os.Getenv("MODE"))
	switch {
	case strings.EqualFold(envMode, "PRODUCTION"):
		mode = api.ModeProduction
	case strings.EqualFold(envMode, "DEBUG"):
	default:
		log.Printf("invalid mode %q; use either DEBUG or PRODUCTION", envMode)
		return
	}

	gqlWhitelist, err := api.ParseWhitelist(os.DirFS("."), envGQLWhitelistPath)
	if err != nil {
		log.Printf("parsing whitelist: %v", err)
		return
	}
	for name := range gqlWhitelist {
		log.Printf("parsed whitelisted query %q", name)
	}
	log.Printf("total parsed whitelisted queries: %d", len(gqlWhitelist))

	inmemDataProvider := inmem.NewFake()

	apiServer, err := api.NewServer(
		mode,
		[]byte(envJWTSecret),
		inmemDataProvider,
		gqlWhitelist,
	)
	if err != nil {
		log.Printf("initializing api server: %v", err)
		return
	}

	log.Printf("connect to %s for GraphQL playground", envHost)
	err = http.ListenAndServe(envHost, apiServer)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("serving: %v", err)
		return
	}
}
