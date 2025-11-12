package testlib

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
)

func AssertMap(t require.TestingT, expected map[string]any, actual map[string]any, msgAndArgs ...interface{}) {
	expJSON, _ := json.MarshalIndent(expected, "", "  ")
	actJSON, _ := json.MarshalIndent(actual, "", "  ")
	require.Equal(t, string(expJSON), string(actJSON), msgAndArgs...)
}
