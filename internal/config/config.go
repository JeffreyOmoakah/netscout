package config

import (
	"fmt"
	"time"
)

type Config struct {
	// Targets is a list of IP addresses or CIDR ranges to scan
	Targets []string

	// Ports is a string representation of ports to scan (e.g., "80,443,8000-9000")
	Ports string

	// Workers is the number of concurrent scanner workers
	Workers int

	// Timeout is the connection timeout duration
	Timeout time.Duration

	// RateLimit is the maximum requests per second (0 = unlimited)
	RateLimit int

	// OutputFile is the path to write results (empty = stdout)
	OutputFile string

	// OutputFormat is the format for results (text, json, csv)
	OutputFormat string

	// Verbose enables detailed logging
	Verbose bool
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target must be specified")
	}

	if c.Ports == "" {
		return fmt.Errorf("ports must be specified")
	}

	if c.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}

	if c.Workers > 10000 {
		return fmt.Errorf("workers cannot exceed 10000 (too many goroutines)")
	}

	if c.Timeout < time.Millisecond {
		return fmt.Errorf("timeout must be at least 1ms")
	}

	if c.Timeout > 5*time.Minute {
		return fmt.Errorf("timeout cannot exceed 5 minutes")
	}

	if c.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	validFormats := map[string]bool{
		"text": true,
		"json": true,
		"csv":  true,
	}

	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("invalid output format: %s (valid: text, json, csv)", c.OutputFormat)
	}

	return nil
}

// GetWorkerCount returns the configured number of workers
func (c *Config) GetWorkerCount() int {
	return c.Workers
}

// GetTimeout returns the configured timeout
func (c *Config) GetTimeout() time.Duration {
	return c.Timeout
}

// IsVerbose returns whether verbose mode is enabled
func (c *Config) IsVerbose() bool {
	return c.Verbose
}