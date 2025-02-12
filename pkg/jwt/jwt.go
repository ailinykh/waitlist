package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func Encode(secret string, claims map[string]interface{}) (string, error) {
	clms := jwt.MapClaims{}
	for k, v := range claims {
		clms[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clms)
	return token.SignedString([]byte(secret))
}

func Decode(hash, secret string) (map[string]interface{}, error) {
	token, err := jwt.Parse(hash, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("failed to parse jwt claims")
}
