package errors_test

import (
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorCreation(t *testing.T) {
	err := errors.E(errors.Op("test/operation"), "an error occurred", errkind.Permission)
	require.Error(t, err)
	if e, ok := err.(*errors.Error); ok {
		require.Equal(t, errors.Op("test/operation"), e.Op)
		require.Equal(t, errkind.Permission, e.Kind)
		require.Equal(t, "an error occurred", e.Err.Error())
	} else {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
}

func TestNestedErrorCreation(t *testing.T) {
	err1 := errors.E(errors.Op("operation1"), errkind.Permission, "error1 message")
	err2 := errors.E(errors.Op("operation2"), err1, errkind.Permission)

	require.Error(t, err2)
	require.Equal(t, "operation2: permission denied:: operation1: error1 message", err2.Error())
}

func TestStrFunction(t *testing.T) {
	err := errors.Str("simple error")
	require.Error(t, err)
	require.Equal(t, "simple error", err.Error())
}

func TestErrorFormatting(t *testing.T) {
	err := errors.E(errors.Op("format/operation"), "error message", errkind.Internal)
	require.Equal(t, "format/operation: internal error: error message", err.Error())
}

func TestIsFunction(t *testing.T) {
	err := errors.E(errors.Op("check/operation"), "check error", errkind.Permission)

	require.True(t, errors.Is(errkind.Permission, err))
	require.False(t, errors.Is(errkind.Other, err))
}
