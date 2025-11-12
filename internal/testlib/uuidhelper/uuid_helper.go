package uuidhelper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewUUID(t testing.TB) uuid.UUID {
	id, err := uuid.NewUUID()
	require.NoError(t, err, "Failed to generate UUID")
	return id
}
