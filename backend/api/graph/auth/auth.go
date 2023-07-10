package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

func NewJWTGenerator(secret []byte) *JWTGenerator {
	return &JWTGenerator{
		secret: secret,
	}
}

// JWTGenerator generates JWT access tokens based on the given secret.
type JWTGenerator struct{ secret []byte }

func (g *JWTGenerator) GenerateJWT(
	userID string, expiration time.Duration,
) (string, error) {
	now := time.Now()
	claims := &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(expiration).Unix(),
		Issuer:    userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}

func NewJWTMiddleware(secret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				// Unauthenticated client
				next.ServeHTTP(w, r)
				return
			}
			p := strings.Split(h, " ")
			if len(p) != 2 || p[0] != "Bearer" {
				http.Error(w, "Invalid or missing token", http.StatusUnauthorized)
				return
			}

			token, _ := jwt.Parse(p[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(
						"unexpected signing method: %v", token.Header["alg"],
					)
				}
				return secret, nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := context.WithValue(
					r.Context(), contextKeyUserID, claims["iss"],
				)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(
				w,
				`{"errors":[{"message":"unauthorized"}]}`,
				http.StatusUnauthorized,
			)
		})
	}
}

type contextKey int

const contextKeyUserID contextKey = 1

// GetUserIDFromContext gets the user id from the context
// planted by the authentication middleware.
func GetUserIDFromContext(ctx context.Context) string {
	if v := ctx.Value(contextKeyUserID); v != nil {
		return v.(string)
	}
	return ""
}

// RequireAuthenticated returns nil if the client is authenticated,
// otherwise returns ErrUnauthenticated.
func RequireAuthenticated(ctx context.Context) error {
	client := GetUserIDFromContext(ctx)
	if client == "" {
		return ErrUnauthenticated
	}
	return nil
}

// RequireOwner returns nil if the client is authenticated and
// is also the owner of the resource, otherwise returns ErrUnauthenticated.
func RequireOwner(ctx context.Context, ownerID string) error {
	client := GetUserIDFromContext(ctx)
	if client == "" {
		return ErrUnauthenticated
	}
	if client != ownerID {
		return ErrUnauthorized
	}
	return nil
}

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")
)
