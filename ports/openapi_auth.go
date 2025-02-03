package ports

import (
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/errkind"
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type FirebaseAuthHttpMiddleware struct {
	AuthClient *auth.Client
}

func (a FirebaseAuthHttpMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bearerToken := a.tokenFromHeader(r)
		if bearerToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		token, err := a.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			unauthorised("unable-to-verify-jwt", err, w, r)
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

func (a FirebaseAuthHttpMiddleware) tokenFromHeader(r *http.Request) string {
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

func UserFromCtx(ctx context.Context) (User, error) {
	u, ok := ctx.Value(userContextKey).(User)
	if ok {
		return u, nil
	}

	return User{}, errors.E(errors.Op("UserFromCtx"), errkind.Authorization, "no user in context")
}
