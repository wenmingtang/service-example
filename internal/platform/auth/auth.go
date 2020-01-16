package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type KeyLookupFunc func(kid string) (*rsa.PublicKey, error)

func NewSimpleKeyLookupFunc(activeKID string, publicKey *rsa.PublicKey) KeyLookupFunc {
	f := func(kid string) (*rsa.PublicKey, error) {
		if activeKID != kid {
			return nil, fmt.Errorf("unrecogoized key id %q", kid)
		}
		return publicKey, nil
	}
	return f
}

type Authenticator struct {
	privateKey       *rsa.PrivateKey
	activeKID        string
	algorithm        string
	pubKeyLookupFunc KeyLookupFunc
	parser           *jwt.Parser
}

func NewAuthenticator(privateKey *rsa.PrivateKey, activeKID, algorithm string, publicKeyLookupFunc KeyLookupFunc) (*Authenticator, error) {
	if privateKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	if activeKID == "" {
		return nil, errors.New("active kid cannot be blank")
	}

	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, fmt.Errorf("unknown algorithm %q", algorithm)
	}

	if publicKeyLookupFunc == nil {
		return nil, errors.New("public key function cannot be nil")
	}

	parser := jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Authenticator{
		privateKey:       privateKey,
		activeKID:        activeKID,
		algorithm:        algorithm,
		pubKeyLookupFunc: publicKeyLookupFunc,
		parser:           &parser,
	}

	return &a, nil
}

func (a *Authenticator) GenerateToken(claims Claims) (string, error) {
	method := jwt.GetSigningMethod(a.algorithm)

	tkn := jwt.NewWithClaims(method, claims)
	tkn.Header["kid"] = a.activeKID

	str, err := tkn.SignedString(a.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token %w", err)
	}
	return str, nil
}

func (a *Authenticator) ParseClaims(tokenStr string) (Claims, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}

		userKID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}

		return a.pubKeyLookupFunc(userKID)
	}

	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, keyFunc)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}
