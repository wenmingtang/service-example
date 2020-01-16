package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type ctxKey int

const Key ctxKey = 1

type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

func (c Claims) HasRole(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}

func NewClaims(subject string, roles []string, now time.Time, expires time.Duration) Claims {
	c := Claims{
		Roles: roles,
		StandardClaims: jwt.StandardClaims{
			Subject:   subject,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(expires).Unix(),
		},
	}
	return c
}
