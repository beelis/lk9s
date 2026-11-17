package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/mariusgassen/lk9s/internal/config"
	"github.com/mariusgassen/lk9s/internal/lk"
	"github.com/mariusgassen/lk9s/internal/ui"
)

// buildVersion reports the module version, or for local builds "dev" with
// the VCS revision stamped by the Go toolchain.
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}

	var rev string

	dirty := false

	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value
		case "vcs.modified":
			dirty = kv.Value == "true"
		}
	}

	if rev == "" {
		return "dev"
	}

	if len(rev) > 7 {
		rev = rev[:7]
	}

	if dirty {
		rev += "+dirty"
	}

	return "dev (" + rev + ")"
}

func main() {
	os.Exit(run())
}

func run() int {
	contextName := flag.String("context", "", "context name to use (default: interactive selection)")
	debugLog := flag.String(
		"debug-log", "", "path to write debug logs to (room names and API errors); empty disables logging",
	)

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	for _, w := range cfg.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}

	ctx, err := resolveContext(cfg, *contextName)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			return 0
		}

		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	logger, closeLog, err := newDebugLogger(*debugLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}
	defer closeLog()

	dial := func(c config.Context) ui.RoomLister {
		return lk.NewClient(c.URL, c.APIKey, c.APISecret, logger)
	}

	if err := ui.Run(dial, cfg.Contexts, ctx, buildVersion()); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	return 0
}

// newDebugLogger returns a discarding logger when path is empty. The TUI takes over the
// terminal, so logs go to a file, never stdout/stderr. The returned func closes the file.
func newDebugLogger(path string) (*slog.Logger, func(), error) {
	if path == "" {
		return slog.New(slog.DiscardHandler), func() {}, nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("open debug log: %w", err)
	}

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})

	return slog.New(handler), func() { _ = f.Close() }, nil
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
