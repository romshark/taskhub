package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/romshark/taskhub/api"
	"github.com/romshark/taskhub/api/dataprovider/inmem"
	"github.com/romshark/taskhub/api/gqlpq"
	"golang.org/x/exp/slog"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config, err := loadConfig()
	if err != nil {
		log.Error("loading config", slog.Any("error", err))
		return
	}

	// Reset log level
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	persistedQueries, err := gqlpq.New(
		config.GQLSchemaPath,
	)
	if err != nil {
		log.Error("initializing persisted queries list", slog.Any("error", err))
		return
	}

	err = persistedQueries.Load(config.PersistedQueriesDirPath)
	if err != nil {
		log.Error("loading persisted queries", slog.Any("error", err))
		return
	}

	if config.PersistedQueriesHotReload {
		go func() {
			err := persistedQueries.Watch(
				ctx, config.GQLSchemaPath, config.PersistedQueriesDirPath,
				config.PersistedQueriesReloadDebounce,
				func() {
					log.Info(
						"reloaded persisted queries",
						slog.Int("totalQueries", persistedQueries.Len()),
					)
				},
			)
			if err != nil {
				panic(fmt.Errorf("watching persisted queries dir changes: %w", err))
			}
		}()
	}

	persistedQueries.ForEach(func(key, query string) {
		log.Debug(
			"added persisted query",
			slog.String("name", key),
			slog.String("query", query),
		)
	})
	log.Info(
		"added persisted queries",
		slog.Int("totalQueries", persistedQueries.Len()),
	)

	inmemDataProvider := inmem.NewFake()

	apiServer, err := api.NewServer(
		log,
		config.APIMode,
		[]byte(config.JWTSecret),
		inmemDataProvider,
		persistedQueries,
	)
	if err != nil {
		log.Error("initializing api server", slog.Any("error", err))
		return
	}

	if config.APIMode == api.ModeDebug {
		log.Info(
			"GraphiQL playground available",
			slog.String("playgroundAddr", config.Host),
			slog.String("queryAddr", path.Join(config.Host, "/query")),
		)
	}

	httpServer := &http.Server{
		Addr:    config.Host,
		Handler: apiServer,
	}
	go func() {
		log.Info("listening", slog.String("addr", config.Host))
		err = httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("listening HTTP(S): %w", err))
		}
	}()

	<-ctx.Done()
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		log.Error("shutting down http server", slog.Any("error", err))
	}
}

type Config struct {
	LogLevel                       slog.Level
	Host                           string
	JWTSecret                      string
	GQLSchemaPath                  string
	PersistedQueriesDirPath        string
	PersistedQueriesReloadDebounce time.Duration
	PersistedQueriesHotReload      bool
	APIMode                        api.Mode
}

func loadConfig() (*Config, error) {
	c := new(Config)

	switch v := os.Getenv("LOG_LEVEL"); {
	case v == "":
		c.LogLevel = slog.LevelInfo
	case strings.EqualFold(v, "INFO"):
		c.LogLevel = slog.LevelInfo
	case strings.EqualFold(v, "DEBUG"):
		c.LogLevel = slog.LevelDebug
	default:
		return nil, fmt.Errorf(
			"invalid LOG_LEVEL %q; use either DEBUG or INFO", v,
		)
	}

	c.Host = os.Getenv("HOST")
	if c.Host == "" {
		c.Host = "localhost:8080"
	}

	c.JWTSecret = os.Getenv("JWT_SECRET")
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("missing JWT_SECRET")
	}

	c.GQLSchemaPath = os.Getenv("GQL_SCHEMA_PATH")
	if c.GQLSchemaPath == "" {
		c.GQLSchemaPath = "api/graph"
	}

	c.PersistedQueriesDirPath = os.Getenv("GQL_PQ_PATH")
	if c.PersistedQueriesDirPath == "" {
		c.PersistedQueriesDirPath = "persisted_queries"
	}

	c.PersistedQueriesReloadDebounce = 2 * time.Second
	if v := os.Getenv("GQL_PQ_RELOAD_DEBOUNCE"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf(
				"parsing GQL_PQ_RELOAD_DEBOUNCE: %w", err,
			)
		}
		c.PersistedQueriesReloadDebounce = d
	}

	switch v := os.Getenv("GQL_PQ_MODE"); {
	case v == "":
		c.PersistedQueriesHotReload = true
	case strings.EqualFold(v, "ON_INIT"):
		c.PersistedQueriesHotReload = false
	case strings.EqualFold(v, "LIVE"):
		c.PersistedQueriesHotReload = true
	default:
		return nil, fmt.Errorf(
			"invalid GQL_PQ_MODE %q; use either ON_INIT or LIVE", v,
		)
	}

	switch v := strings.ToLower(os.Getenv("MODE")); {
	case strings.EqualFold(v, "PRODUCTION"):
		c.APIMode = api.ModeProduction
	case strings.EqualFold(v, "DEBUG"):
		c.APIMode = api.ModeDebug
	default:
		return nil, fmt.Errorf("invalid MODE %q; use either DEBUG or PRODUCTION", v)
	}
	return c, nil
}
