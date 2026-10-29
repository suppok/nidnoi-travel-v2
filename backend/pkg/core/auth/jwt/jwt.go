package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"backend/pkg/core/utils"
)

const issuer = "nidnoi-api"

// Service issues and verifies short-lived access tokens. The token carries only the user id:
// roles, status and the password-change flag are loaded from the database on every request,
// so role changes and deactivation take effect immediately.
type Service interface {
	Issue(userID string) (token string, expiresIn time.Duration, err error)
	Parse(token string) (userID string, err error)
}

type service struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) Service {
	return &service{secret: []byte(secret), ttl: ttl}
}

func (s *service) Issue(userID string) (string, time.Duration, error) {
	now := time.Now()
	claims := gojwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   userID,
		ID:        utils.NewID(),
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(s.ttl)),
	}
	token, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}
	return token, s.ttl, nil
}

func (s *service) Parse(token string) (string, error) {
	claims := &gojwt.RegisteredClaims{}
	_, err := gojwt.ParseWithClaims(token, claims, func(*gojwt.Token) (any, error) { return s.secret, nil },
		gojwt.WithValidMethods([]string{gojwt.SigningMethodHS256.Alg()}),
		gojwt.WithIssuer(issuer),
		gojwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", errors.New("token has no subject")
	}
	return claims.Subject, nil
}
