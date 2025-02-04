package ports

import (
	"clean-arch-go/core/auth"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/user"
	"context"
	"net/http"
	"strings"
)

type OpenapiAuthMiddleware struct {
	auth auth.Auth
}

func newOpenapiAuthMiddleware(a auth.Auth) OpenapiAuthMiddleware {
	return OpenapiAuthMiddleware{auth: a}
}

func (a OpenapiAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bearerToken := a.tokenFromHeader(r)
		if bearerToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := a.auth.VerifyIDToken(bearerToken)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(r.Context(), userContextKey, User{
			UUID: claims["user_uuid"].(string),
			Role: claims["user_role"].(string),
		})
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (a OpenapiAuthMiddleware) tokenFromHeader(r *http.Request) string {
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

type ctxKey string

const (
	userContextKey ctxKey = "user-context-key"
)

func userFromCtx(ctx context.Context) (user.User, error) {
	op := errors.Op("userFromCtx")

	u, ok := ctx.Value(userContextKey).(User)
	if !ok {
		return nil, errors.E(op, "no user in context")
	}

	domainUser, err := user.From(u.UUID, u.Role)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainUser, nil
}
