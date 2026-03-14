package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sitewatch/internal/monitor"
)

func main() {
	// --- CLI Flags ---
	interval := flag.Duration("interval", 5*time.Second, "check interval (e.g. 5s, 1m)")
	timeout := flag.Duration("timeout", 10*time.Second, "request timeout (e.g. 3s, 10s)")
	flag.Parse()

	// --- Validate positional argument (the URL) ---
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: sitewatch <url> [--interval 5s] [--timeout 10s]")
		os.Exit(1)
	}
	url := flag.Arg(0)

	// --- Setup ---
	fmt.Printf("Monitoring %s every %s (timeout: %s)\n\n", url, *interval, *timeout)

	m := monitor.New(url, *interval, *timeout)

	// --- Graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// --- Run ---
	m.Run(ctx)
}