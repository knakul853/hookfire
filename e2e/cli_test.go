//go:build e2e

package e2e

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func build(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "hookfire")
	out, err := exec.Command("go", "build", "-o", bin, "../cmd/hookfire").CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}

func TestE2EDryRunJSON(t *testing.T) {
	bin := build(t)
	cmd := exec.Command(bin, "trigger", "github", "push", "--dry-run", "--json", "--url", "http://x/webhook", "--secret", "s")
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	require.NoError(t, cmd.Run())
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout.String()), &got), "stdout must be valid JSON: %q (stderr=%q)", stdout.String(), stderr.String())
	require.Equal(t, "github", got["provider"])
	require.Equal(t, true, got["dry_run"])
}

func TestE2EUnknownProviderExit2(t *testing.T) {
	bin := build(t)
	cmd := exec.Command(bin, "trigger", "gitlab", "push", "--url", "http://x", "--secret", "s")
	err := cmd.Run()
	var ee *exec.ExitError
	require.ErrorAs(t, err, &ee)
	require.Equal(t, 2, ee.ExitCode())
}

func TestE2EListOnStdout(t *testing.T) {
	bin := build(t)
	cmd := exec.Command(bin, "list")
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	require.NoError(t, cmd.Run())
	require.Contains(t, stdout.String(), "github")
	require.Empty(t, stderr.String(), "list must not write to stderr")
}

func TestE2EVersionOnStdout(t *testing.T) {
	bin := build(t)
	cmd := exec.Command(bin, "version")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	require.NoError(t, cmd.Run())
	require.Contains(t, stdout.String(), "hookfire")
}
