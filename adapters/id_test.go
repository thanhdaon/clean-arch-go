package adapters_test

import (
	"clean-arch-go/adapters"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID_GenerationAndValidation(t *testing.T) {
	t.Parallel()

	idGen := adapters.NewID()
	generatedID := idGen.New()
	require.NotEmpty(t, generatedID)
	require.True(t, idGen.IsValid(generatedID))
}

func TestID_InvalidID(t *testing.T) {
	t.Parallel()

	idGen := adapters.NewID()
	invalidID := "invalid-ulid"
	require.False(t, idGen.IsValid(invalidID))
}
