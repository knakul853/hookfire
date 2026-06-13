package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/spf13/cobra"
)

var verbosity int

// NewRootCmd builds the hookfire command tree. main.go calls Run, which calls this.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hookfire",
		Short:         "Fire correctly-signed synthetic webhook events at any URL",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			configureLogging(verbosity)
		},
	}
	root.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Verbose diagnostics to stderr (-v info, -vv debug)")
	root.AddCommand(newVersionCmd(), newTriggerCmd(), newShowCmd(), newListCmd(), newReplayCmd(), newVerifyCmd())
	return root
}

// Run executes the root command and returns the documented process exit code.
func Run(args []string) int {
	root := NewRootCmd()
	root.SetArgs(args)
	err := root.Execute()
	if err != nil {
		slog.Error(err.Error())
	}
	return exitCodeFor(err)
}

// configureLogging sends slog to stderr at a level gated by -v count: warn by
// default, info at -v, debug at -vv. stdout is reserved for command output.
func configureLogging(v int) {
	level := slog.LevelWarn
	switch {
	case v >= 2:
		level = slog.LevelDebug
	case v == 1:
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

// loadConfig loads the config from the default path, degrading to an empty
// config (with a warning) on any error — config is optional.
func loadConfig() config.Config {
	path, err := config.DefaultPath()
	if err != nil {
		slog.Debug("cannot resolve config path", "err", err)
		return config.Config{Targets: map[string]config.Target{}}
	}
	cfg, err := config.Load(path)
	if err != nil {
		slog.Warn("failed to load config", "path", path, "err", err)
		return config.Config{Targets: map[string]config.Target{}}
	}
	return cfg
}

// suggestErr enriches an unknown-provider/event error with a near-match hint.
func suggestErr(cat provider.Catalog, providerName, event string, err error) error {
	switch {
	case errors.Is(err, provider.ErrUnknownProvider):
		if s, ok := nearest(providerName, cat.List()); ok {
			return fmt.Errorf("%w (did you mean %q?)", err, s)
		}
	case errors.Is(err, provider.ErrUnknownEvent):
		if evs, eerr := cat.Events(providerName); eerr == nil {
			if s, ok := nearest(event, evs); ok {
				return fmt.Errorf("%w (did you mean %q?)", err, s)
			}
		}
	}
	return err
}
