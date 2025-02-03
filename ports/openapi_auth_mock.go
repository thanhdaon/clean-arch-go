package ports

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/request"
)

func HttpMockMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claims jwt.MapClaims
		token, err := request.ParseFromRequest(
			r,
			request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (i interface{}, e error) {
				return []byte("mock_secret"), nil
			},
			request.WithClaims(&claims),
		)
		if err != nil {
			badRequest(err, w, r)
			return
		}

		if !token.Valid {
			badRequest(fmt.Errorf("invalid-jwt"), w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, User{
			UUID: claims["user_uuid"].(string),
			Role: claims["role"].(string),
		})
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
