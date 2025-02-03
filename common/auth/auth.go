package auth

import (
	"clean-arch-go/common/errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	secret []byte
}

func NewAuth(secret string) Auth {
	return Auth{secret: []byte(secret)}
}

func (a Auth) VerifyIDToken(tokenString string) (map[string]any, error) {
	op := errors.Op("auth.VerifyIDToken")

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", errors.E(op, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"]))
		}

		return a.secret, nil
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, errors.E(op, fmt.Errorf("cannot extract data from claims"))
}

func (a Auth) CreateIDToken(data map[string]any) (string, error) {
	op := errors.Op("auth.CreateIDToken")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(data))

	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", errors.E(op, err)
	}

	return tokenString, nil
}
