package ports

import (
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/errkind"
	"context"
	"net/http"
	"strings"
)

type JwtHttpMiddleware struct {
	jwtSecret []byte
}

func (a JwtHttpMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bearerToken := a.tokenFromHeader(r)
		if bearerToken == "" {
			Unauthorised("empty-bearer-token", nil, w, r)
			return
		}

		token, err := a.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			Unauthorised("unable-to-verify-jwt", err, w, r)
			return
		}

		ctx = context.WithValue(ctx, userContextKey, User{
			UUID: token.UID,
			Role: token.Claims["role"].(string),
		})
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a JwtHttpMiddleware) tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")

	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}

	return ""
}

type User struct {
	UUID string
	Role string
}

type ctxKey int

const (
	userContextKey ctxKey = iota
)

func UserFromCtx(ctx context.Context) (User, error) {
	u, ok := ctx.Value(userContextKey).(User)
	if ok {
		return u, nil
	}

	return User{}, errors.E(errors.Op("UserFromCtx"), errkind.Authorization, "no user in context")
}
