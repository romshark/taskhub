// Package api provides the GraphQL API server.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/gqlpq"
	"github.com/romshark/taskhub/api/graph"
	"github.com/romshark/taskhub/api/jwt"
	"github.com/romshark/taskhub/api/passhash"
	"github.com/romshark/taskhub/api/reqctx"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"golang.org/x/exp/slog"
)

// TimeProviderLive provides live time.
type TimeProviderLive struct{}

func (p *TimeProviderLive) Now() time.Time { return time.Now() }

type ServerDebug struct {
	playgroundHandler http.Handler
	productionServer  *ServerProduction
}

func (s *ServerDebug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		if r.Method != http.MethodGet {
			httpNotFound(w)
			return
		}
		s.playgroundHandler.ServeHTTP(w, r)
		return
	case r.URL.Path == "/query":
		s.productionServer.gqlHandler.ServeHTTP(w, r)
	default:
		s.productionServer.ServeHTTP(w, r)
	}
}

type ServerProduction struct {
	log              *slog.Logger
	gqlHandler       http.Handler
	persistedQueries *gqlpq.PersistedQueries
}

func (s *ServerProduction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key, ok := strings.CutPrefix(r.URL.Path, "/e/")
	if !ok {
		httpNotFound(w)
		return
	}
	query := s.persistedQueries.GetQuery(key)
	if query == "" {
		httpNotFound(w)
		return
	}

	originalBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Info("reading request body", slog.Any("error", err))
		return
	}
	r.Body.Close()

	if len(originalBody) > 0 {
		if !json.Valid(originalBody) {
			http.Error(w, "invalid variables JSON", http.StatusBadRequest)
			return
		}

		// Query with variables
		originalBody = bytes.TrimLeft(originalBody, " \n\r\t")
		if len(originalBody) < 1 || originalBody[0] != '{' {
			http.Error(
				w, "body must contain object with variables", http.StatusBadRequest,
			)
			return
		}
		r.Body = makeQueryWithVars(query, originalBody)
	} else {
		// Query without variables
		r.Body = makeQuery(query)
	}

	s.gqlHandler.ServeHTTP(w, r)
}

type Mode int8

const (
	ModeDebug      Mode = 0
	ModeProduction Mode = 1
)

func NewServer(
	log *slog.Logger,
	mode Mode,
	jwtSecret []byte,
	dataProvider dataprovider.DataProvider,
	persistedQueries *gqlpq.PersistedQueries,
) (http.Handler, error) {
	gqlResolver := graph.NewResolver(
		dataProvider,
		jwt.NewJWTGenerator(jwtSecret),
		passhash.NewPasswordHasherBcrypt(0),
		new(TimeProviderLive),
	)
	conf := graph.Config{Resolvers: gqlResolver}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(conf))
	srv.AddTransport(&transport.Websocket{})
	srv.AroundResponses(newGQLMiddlewareLogResponses(log))

	withContextMiddleware := newMiddlewareSetRequestContext(srv, log, jwtSecret)

	prodSrv := &ServerProduction{
		log:              log,
		gqlHandler:       withContextMiddleware,
		persistedQueries: persistedQueries,
	}
	if mode == ModeDebug {
		play := playground.Handler("GraphQL Playground", "/query")
		return &ServerDebug{
			playgroundHandler: play,
			productionServer:  prodSrv,
		}, nil
	}
	return prodSrv, nil
}

func newGQLMiddlewareLogResponses(log *slog.Logger) func(
	ctx context.Context, next graphql.ResponseHandler,
) *graphql.Response {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		requestContext := graphql.GetOperationContext(ctx)
		reqCtx := reqctx.GetRequestContext(ctx)
		took := time.Since(reqCtx.Start)
		if requestContext.Operation != nil {
			if reqCtx.PersistedQueryName != "" {
				log.Info(
					"handled persisted operation",
					slog.String("requestID", string(reqCtx.RequestID)),
					slog.String("persistedQueryName", reqCtx.PersistedQueryName),
					slog.String("type", string(requestContext.Operation.Operation)),
					slog.String("took", took.String()),
				)
			} else {
				log.Info(
					"handled operation",
					slog.String("requestID", string(reqCtx.RequestID)),
					slog.String("type", string(requestContext.Operation.Operation)),
					slog.String("took", took.String()),
				)
			}
		} else {
			log.Info(
				"handled invalid request",
				slog.String("took", took.String()),
			)
		}
		return next(ctx)
	}
}

func newMiddlewareSetRequestContext(
	next http.Handler,
	log *slog.Logger,
	jwtSecret []byte,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var persistedQueryName string
		if s, ok := strings.CutPrefix(r.URL.Path, "/e/"); ok {
			persistedQueryName = s
		}

		userID, err := jwt.GetUserID(jwtSecret, r, time.Now())
		switch {
		case err == nil:
		case errors.Is(err, jwt.ErrTokenInvalid):
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		case errors.Is(err, jwt.ErrTokenExpired):
			http.Error(w, "expired bearer token", http.StatusUnauthorized)
			return
		default:
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		ctx := reqctx.WithRequestContext(
			r.Context(), log, userID, persistedQueryName, time.Now(),
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func httpNotFound(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func makeQuery(query string) io.ReadCloser {
	b := new(bytes.Buffer)
	b.Grow(len(`{"query":"`) + len(query) + len(`"}`))
	b.WriteString(`{"query":`)
	b.WriteString(query)
	b.WriteString(`}`)
	return io.NopCloser(b)
}

func makeQueryWithVars(query string, variablesObjJSON []byte) io.ReadCloser {
	b := new(bytes.Buffer)
	b.Grow(
		len(`{"query":"`) +
			len(query) +
			len(`","variables":`) +
			len(variablesObjJSON) +
			len(`}`),
	)
	b.WriteString(`{"query":`)
	b.WriteString(query)
	b.WriteString(`,"variables":`)
	b.Write(variablesObjJSON)
	b.WriteString(`}`)
	return io.NopCloser(b)
}
