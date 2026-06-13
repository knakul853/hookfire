package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModeSelection(t *testing.T) {
	require.Equal(t, ModeJSON, SelectMode(Env{IsTTY: true, NoColor: false, JSON: true}))
	require.Equal(t, ModePlain, SelectMode(Env{IsTTY: false, JSON: false}))
	require.Equal(t, ModePlain, SelectMode(Env{IsTTY: true, NoColor: true, JSON: false}))
	require.Equal(t, ModeHUD, SelectMode(Env{IsTTY: true, NoColor: false, JSON: false}))
}
