package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/beelis/lk9s/internal/ui"
)

func main() {
	url := flag.String("url", "", "LiveKit server URL")
	apiKey := flag.String("api-key", "", "LiveKit API key")
	apiSecret := flag.String("api-secret", "", "LiveKit API secret")
	flag.Parse()

	if *url == "" || *apiKey == "" || *apiSecret == "" {
		fmt.Fprintln(os.Stderr, "usage: lk9s -url <url> -api-key <key> -api-secret <secret>")
		os.Exit(1)
	}

	if err := ui.Run(lk.NewClient(*url, *apiKey, *apiSecret)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
