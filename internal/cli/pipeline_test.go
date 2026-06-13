package cli

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/fire"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/knakul853/hookfire/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const githubManifestYAML = `
name: github
transport:
  method: POST
  content_type: application/json
  headers:
    User-Agent: "GitHub-Hookshot/hookfire"
signing:
  scheme: hmac
  algorithm: sha256
  encoding: hex
  basestring: "{{body}}"
  output: "sha256={{sig}}"
  header: "X-Hub-Signature-256"
events:
  push:
    headers: { X-GitHub-Event: push }
`

func oneProviderFS(t *testing.T) fs.FS {
	t.Helper()
	return fstest.MapFS{
		"github/manifest.yaml":    &fstest.MapFile{Data: []byte(githubManifestYAML)},
		"github/events/push.json": &fstest.MapFile{Data: []byte(`{"ref":"refs/heads/main"}`)},
	}
}

func githubEvent(t *testing.T) (provider.Manifest, provider.Event) {
	t.Helper()
	c, err := provider.NewCatalog(provider.Sources{Embedded: oneProviderFS(t)})
	require.NoError(t, err)
	mm, ev, err := c.Lookup("github", "push")
	require.NoError(t, err)
	return mm, ev
}

func TestPipelineTriggerSignsAndFires(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)

	sender := mocks.NewMockSender(t)
	sender.EXPECT().Send(mock.Anything, mock.Anything).
		Return(fire.Result{Status: 200, Body: []byte(`{"ok":true}`), Bytes: 11}, nil)

	view, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook",
		Secret: config.Secret("hookfire_test_secret"), Timestamp: 1700000000,
	})
	require.NoError(t, err)
	require.NotNil(t, view.Response)
	require.Equal(t, 200, view.Response.Status)
	require.True(t, view.Signed)
	require.Equal(t, "X-Hub-Signature-256", view.Signature.Header)
	require.Equal(t, "hmac-sha256", view.Signature.Scheme)
}

func TestPipelineDryRunDoesNotSend(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)
	sender := mocks.NewMockSender(t) // no Send expectation → fails if called

	view, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook",
		Secret: config.Secret("s"), Timestamp: 1700000000, DryRun: true,
	})
	require.NoError(t, err)
	require.Nil(t, view.Response)
	require.True(t, view.DryRun)
}

func TestPipelineNoSignSkipsSignature(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)
	sender := mocks.NewMockSender(t)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(fire.Result{Status: 200}, nil)

	view, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook",
		Timestamp: 1700000000, NoSign: true,
	})
	require.NoError(t, err)
	require.False(t, view.Signed)
	require.True(t, view.NoSign)
}

func TestPipelineMissingSecretErrors(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)
	sender := mocks.NewMockSender(t)

	_, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook", Timestamp: 1700000000,
	})
	require.ErrorIs(t, err, sign.ErrMissingSecret)
}

func TestPipelineFailFlagOnNon2xx(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)
	sender := mocks.NewMockSender(t)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(fire.Result{Status: 500, Body: []byte("err"), Bytes: 3}, nil)

	view, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook",
		Secret: config.Secret("s"), Timestamp: 1700000000, Fail: true,
	})
	require.Error(t, err)
	require.Equal(t, 22, exitCodeFor(err))
	require.NotNil(t, view.Response) // response still captured for rendering
}

func TestPipelineTransportError(t *testing.T) {
	cat := mocks.NewMockCatalog(t)
	m, ev := githubEvent(t)
	cat.EXPECT().Lookup("github", "push").Return(m, ev, nil)
	sender := mocks.NewMockSender(t)
	sender.EXPECT().Send(mock.Anything, mock.Anything).Return(fire.Result{}, errors.New("dial tcp: refused"))

	view, err := runPipeline(pipelineDeps{Catalog: cat, Sender: sender}, pipelineInput{
		Provider: "github", Event: "push", URL: "http://x/webhook",
		Secret: config.Secret("s"), Timestamp: 1700000000,
	})
	require.Equal(t, 4, exitCodeFor(err))
	require.NotEmpty(t, view.Err)
}
