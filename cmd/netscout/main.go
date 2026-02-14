package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JeffreyOmoakah/netscout.git/internal/config"
	"github.com/JeffreyOmoakah/netscout.git/internal/scanner"
)

var (
	version = "0.1.0"
)

func main() {
	// Define CLI flags
	var (
		targets     = flag.String("t", "", "Target IP or CIDR range (e.g., 192.168.1.0/24)")
		ports       = flag.String("p", "80,443", "Ports to scan (e.g., 80,443 or 1-1024)")
		workers     = flag.Int("w", 100, "Number of concurrent workers")
		timeout     = flag.Duration("timeout", 2*time.Second, "Connection timeout")
		rateLimit   = flag.Int("rate", 0, "Rate limit (requests per second, 0 = unlimited)")
		outputFile  = flag.String("o", "", "Output file (default: stdout)")
		outputFmt   = flag.String("f", "text", "Output format (text, json, csv)")
		showVersion = flag.Bool("version", false, "Show version")
		verbose     = flag.Bool("v", false, "Verbose output")
	)

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("NETscout v%s\n", version)
		os.Exit(0)
	}

	// Validate required arguments
	if *targets == "" {
		fmt.Fprintf(os.Stderr, "Error: target (-t) is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Build configuration
	cfg := &config.Config{
		Targets:      parseTargets(*targets),
		Ports:        *ports,
		Workers:      *workers,
		Timeout:      *timeout,
		RateLimit:    *rateLimit,
		OutputFile:   *outputFile,
		OutputFormat: *outputFmt,
		Verbose:      *verbose,
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Create scanner instance
	s, err := scanner.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create scanner: %v\n", err)
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Print scan info
	if *verbose {
		fmt.Fprintf(os.Stderr, "Starting NETscout v%s\n", version)
		fmt.Fprintf(os.Stderr, "Targets: %s\n", *targets)
		fmt.Fprintf(os.Stderr, "Ports: %s\n", *ports)
		fmt.Fprintf(os.Stderr, "Workers: %d\n", *workers)
		fmt.Fprintf(os.Stderr, "Timeout: %v\n", *timeout)
		fmt.Fprintln(os.Stderr, "")
	}

	// Run the scan
	if err := s.Scan(ctx); err != nil {
		if err == context.Canceled {
			fmt.Fprintln(os.Stderr, "Scan cancelled by user")
			os.Exit(130) // Standard exit code for SIGINT
		}
		fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
		os.Exit(1)
	}

	// Print summary if verbose
	if *verbose {
		summary := s.GetSummary()
		fmt.Fprintf(os.Stderr, "\nScan completed:\n")
		fmt.Fprintf(os.Stderr, "  Total scanned: %d\n", summary.TotalScanned)
		fmt.Fprintf(os.Stderr, "  Open ports: %d\n", summary.OpenPorts)
		fmt.Fprintf(os.Stderr, "  Closed ports: %d\n", summary.ClosedPorts)
		fmt.Fprintf(os.Stderr, "  Filtered: %d\n", summary.Filtered)
		fmt.Fprintf(os.Stderr, "  Duration: %v\n", summary.Duration)
	}
}

// parseTargets splits comma-separated targets
func parseTargets(targets string) []string {
	parts := strings.Split(targets, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}