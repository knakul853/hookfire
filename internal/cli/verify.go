package cli

import (
	"fmt"
	"io"

	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/spf13/cobra"
)

const (
	verifySecret = "hookfire_verify_secret"
	verifyTS     = int64(1700000000)
)

var verifyBody = []byte(`{"hookfire":"verify"}`)

func newVerifyCmd() *cobra.Command {
	var providersDir string
	cmd := &cobra.Command{
		Use:   "verify [providers-dir]",
		Short: "Check that every provider's signing scheme constructs and signs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := providersDir
			if len(args) == 1 {
				dir = args[0]
			}
			cat, err := buildCatalog(dir, func(string, string) {})
			if err != nil {
				return err
			}
			return runVerify(cmd.OutOrStdout(), cat.List(), func(name string) (string, error) {
				m, err := cat.Manifest(name)
				if err != nil {
					return "", err
				}
				return verifyProvider(m)
			})
		},
	}
	cmd.Flags().StringVar(&providersDir, "providers-dir", "", "Extra providers directory to verify")
	return cmd
}

// verifyProvider builds the signer for m and signs the golden body twice,
// returning the signature header value. It errors if construction fails, the
// signature is empty, or two signings disagree (non-deterministic).
func verifyProvider(m provider.Manifest) (string, error) {
	signer, err := sign.New(m.SigningConfig())
	if err != nil {
		return "", err
	}
	opts := sign.Options{Secret: verifySecret, Timestamp: verifyTS}
	h1, err := signer.Sign(verifyBody, opts)
	if err != nil {
		return "", err
	}
	h2, err := signer.Sign(verifyBody, opts)
	if err != nil {
		return "", err
	}
	header := m.Signing.Header
	v1, v2 := h1.Get(header), h2.Get(header)
	if m.Signing.Scheme != "none" && v1 == "" {
		return "", fmt.Errorf("signer produced an empty %q header", header)
	}
	if v1 != v2 {
		return "", fmt.Errorf("signature is non-deterministic")
	}
	return v1, nil
}

// runVerify prints a per-provider ✓/✗ table to w and returns an error if any
// provider failed.
func runVerify(w io.Writer, providers []string, check func(name string) (string, error)) error {
	failed := 0
	for _, name := range providers {
		sigVal, err := check(name)
		if err != nil {
			failed++
			if _, werr := fmt.Fprintf(w, "✗ %s: %v\n", name, err); werr != nil {
				return werr
			}
			continue
		}
		if _, werr := fmt.Fprintf(w, "✓ %s: %s\n", name, sigVal); werr != nil {
			return werr
		}
	}
	if failed > 0 {
		return fmt.Errorf("verify: %d provider(s) failed", failed)
	}
	return nil
}
