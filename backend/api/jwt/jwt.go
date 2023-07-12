package jwt

import (
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

func GetUserID(secret []byte, r *http.Request, timeNow time.Time) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		// Unauthenticated client
		return "", nil
	}
	p := strings.Split(h, " ")
	if len(p) != 2 || p[0] != "Bearer" {
		return "", ErrTokenInvalid
	}
	token, _ := jwt.Parse(p[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(
				"unexpected signing method: %v", token.Header["alg"],
			)
		}
		return secret, nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrTokenInvalid
	}
	if !token.Valid {
		if exp := claims["exp"]; exp != nil &&
			timeNow.Unix() > int64(claims["exp"].(float64)) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}
	return claims["iss"].(string), nil
}

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)
