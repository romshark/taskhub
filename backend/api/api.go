// Package api provides the GraphQL API server.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/romshark/taskhub/api/auth"
	"github.com/romshark/taskhub/api/dataprovider"
	"github.com/romshark/taskhub/api/graph"
	"github.com/romshark/taskhub/api/passhash"
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
		if r.Method != http.MethodPost {
			httpNotFound(w)
			return
		}
		s.productionServer.gqlHandler.ServeHTTP(w, r)
	default:
		s.productionServer.ServeHTTP(w, r)
	}
}

type ServerProduction struct {
	gqlHandler http.Handler
	whitelist  GraphQLWhitelist
}

func (s *ServerProduction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := strings.CutPrefix(r.URL.Path, "/e/")
	if !ok {
		httpNotFound(w)
		return
	}
	query, ok := s.whitelist[id]
	if !ok {
		httpNotFound(w)
		return
	}

	originalBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("reading request body: ", err)
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

type GraphQLWhitelist map[string]string

func NewServer(
	mode Mode,
	jwtSecret []byte,
	dataProvider dataprovider.DataProvider,
	whitelist GraphQLWhitelist,
) (http.Handler, error) {
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

	prodSrv := &ServerProduction{
		gqlHandler: withStartTime,
		whitelist:  whitelist,
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

func gqlMiddlewareLogResponses(
	ctx context.Context, next graphql.ResponseHandler,
) *graphql.Response {
	requestContext := graphql.GetOperationContext(ctx)
	start := ctx.Value(ctxKeyStartTime).(time.Time)
	took := time.Since(start)
	if requestContext.Operation != nil {
		log.Print(requestContext.Operation.Operation, "; took ", took)
	} else {
		log.Print("invalid request; took: ", took)
	}
	return next(ctx)
}

func newMiddlewareSetStartTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyStartTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ctxKey int8

var ctxKeyStartTime ctxKey = 1

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

const gqlFileExtension = ".graphql"

func ParseWhitelist(filesystem fs.FS, path string) (GraphQLWhitelist, error) {
	m := make(GraphQLWhitelist)
	dir, err := fs.ReadDir(filesystem, path)
	if err != nil {
		return nil, fmt.Errorf("reading query directory path: %w", err)
	}
	for _, o := range dir {
		if o.IsDir() {
			continue
		}
		n := o.Name()
		if !strings.HasSuffix(n, gqlFileExtension) {
			continue
		}
		if n != url.PathEscape(n) {
			return nil, fmt.Errorf("invalid file name (not URL safe): %q", n)
		}
		p := filepath.Join(path, n)
		query, err := fs.ReadFile(filesystem, p)
		if err != nil {
			return nil, fmt.Errorf("reading file query %q: %w", p, err)
		}
		n, _ = strings.CutSuffix(n, gqlFileExtension)
		encodedQuery, err := json.Marshal(string(query))
		if err != nil {
			return nil, fmt.Errorf("encoding query to JSON string: %w", err)
		}
		m[n] = string(encodedQuery)
	}

	return m, nil
}
