package main

import (
	"flag"
	"fmt"
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

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, err := resolveContext(cfg, *contextName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := ui.Run(lk.NewClient(ctx.URL, ctx.APIKey, ctx.APISecret), ctx.Name, buildVersion()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
