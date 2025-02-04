package auth_test

import (
	"clean-arch-go/core/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndVerifyIDToken(t *testing.T) {
	a := auth.NewAuth("testsecret")

	data := map[string]any{
		"user_id": "1234",
		"role":    "admin",
	}

	tokenString, err := a.CreateIDToken(data)
	require.NoError(t, err, "should create token without error")
	require.NotEmpty(t, tokenString, "token string should not be empty")

	verifiedData, err := a.VerifyIDToken(tokenString)
	require.NoError(t, err, "should verify token without error")
	assert.Equal(t, data, verifiedData, "verified data should match the original data")
}

func TestVerifyInvalidToken(t *testing.T) {
	a := auth.NewAuth("testsecret")

	invalidToken := "this.is.an.invalid.token"

	_, err := a.VerifyIDToken(invalidToken)
	assert.Error(t, err, "should return an error for an invalid token")
}
