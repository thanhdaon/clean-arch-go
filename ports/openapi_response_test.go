package ports_test

import (
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/ports"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapErrorToStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "authorization error returns 401",
			err:      errors.E(errors.Op("test"), errkind.Authorization, "unauthorized"),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "permission error returns 403",
			err:      errors.E(errors.Op("test"), errkind.Permission, "forbidden"),
			expected: http.StatusForbidden,
		},
		{
			name:     "not exist error returns 404",
			err:      errors.E(errors.Op("test"), errkind.NotExist, "not found"),
			expected: http.StatusNotFound,
		},
		{
			name:     "exist error returns 409",
			err:      errors.E(errors.Op("test"), errkind.Exist, "already exists"),
			expected: http.StatusConflict,
		},
		{
			name:     "connection error returns 503",
			err:      errors.E(errors.Op("test"), errkind.Connection, "db down"),
			expected: http.StatusServiceUnavailable,
		},
		{
			name:     "internal error returns 500",
			err:      errors.E(errors.Op("test"), errkind.Internal, "internal"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "other error returns 500",
			err:      errors.E(errors.Op("test"), errkind.Other, "other"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "unknown error returns 500",
			err:      errors.E(errors.Op("test"), "unknown error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ports.MapErrorToStatus(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseErrorWithMapper(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "not exist returns 404",
			err:            errors.E(errors.Op("test"), errkind.NotExist, "item not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "authorization returns 401",
			err:            errors.E(errors.Op("test"), errkind.Authorization, "unauthorized"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			ctx := context.Background()

			ports.ResponseError(ctx, tt.err, w, r)

			require.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
