package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/beelis/lk9s/internal/config"
	"github.com/beelis/lk9s/internal/lk"
	"github.com/beelis/lk9s/internal/ui"
)

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
		return info.Main.Version
	}

	return "dev"
}

func main() {
	contextName := flag.String("context", "", "context name to use (default: interactive selection)")
	debugLog := flag.String("debug-log", "", "path to write debug logs to (API requests/responses); empty disables logging")

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, err := resolveContext(cfg, *contextName)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			os.Exit(0)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger, err := newDebugLogger(*debugLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dial := func(c config.Context) ui.RoomLister {
		return lk.NewClient(c.URL, c.APIKey, c.APISecret, logger)
	}

	if err := ui.Run(dial, cfg.Contexts, ctx, buildVersion()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newDebugLogger returns nil (lk.NewClient then discards logs) when path is
// empty. The TUI takes over the terminal, so logs go to a file, never
// stdout/stderr.
func newDebugLogger(path string) (*slog.Logger, error) {
	if path == "" {
		return nil, nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open debug log: %w", err)
	}

	return slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})), nil
}

func resolveContext(cfg *config.Config, name string) (config.Context, error) {
	if name != "" {
		for _, ctx := range cfg.Contexts {
			if ctx.Name == name {
				return ctx, nil
			}
		}

		return config.Context{}, fmt.Errorf("context %q not found", name)
	}

	if len(cfg.Contexts) == 1 {
		return cfg.Contexts[0], nil
	}

	return ui.SelectContext(cfg.Contexts)
}
