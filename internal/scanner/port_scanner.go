package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JeffreyOmoakah/netscout.git/internal/config"
	"github.com/JeffreyOmoakah/netscout.git/internal/parser"
	"github.com/JeffreyOmoakah/netscout.git/internal/result"
	"github.com/JeffreyOmoakah/netscout.git/internal/worker"
)

// Scanner orchestrates the network scanning process
type Scanner struct {
	config      *config.Config
	collector   *result.Collector
	pool        *worker.Pool
	targets     []string
	ports       []int
	resultChan  chan *result.Result
	rateLimiter *time.Ticker
}

// New creates a new Scanner instance
func New(cfg *config.Config) (*Scanner, error) {
	// Parse targets
	targets, err := parser.ParseTargets(cfg.Targets)
	if err != nil {
		return nil, fmt.Errorf("failed to parse targets: %w", err)
	}

	// Parse ports
	ports, err := parser.ParsePorts(cfg.Ports)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ports: %w", err)
	}

	// Create result collector
	collector, err := result.NewCollector(cfg.OutputFile, cfg.OutputFormat, cfg.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create collector: %w", err)
	}

	// Create result channel
	resultChan := make(chan *result.Result, 1000)

	// Create worker pool
	pool := worker.NewPool(cfg.Workers, resultChan, cfg.Timeout)

	// Create rate limiter if needed
	var rateLimiter *time.Ticker
	if cfg.RateLimit > 0 {
		interval := time.Second / time.Duration(cfg.RateLimit)
		rateLimiter = time.NewTicker(interval)
	}

	s := &Scanner{
		config:      cfg,
		collector:   collector,
		pool:        pool,
		targets:     targets,
		ports:       ports,
		resultChan:  resultChan,
		rateLimiter: rateLimiter,
	}

	return s, nil
}

// Scan executes the network scan
func (s *Scanner) Scan(ctx context.Context) error {
	// Start the worker pool
	s.pool.Start(ctx)

	// Start result collection goroutine
	go s.collectResults()

	// Start progress reporting if verbose
	var progressWg sync.WaitGroup
	if s.config.Verbose {
		progressWg.Add(1)
		go s.reportProgress(ctx, &progressWg)
	}

	// Generate and submit tasks
	totalTasks := len(s.targets) * len(s.ports)
	if s.config.Verbose {
		fmt.Printf("Scanning %d hosts across %d ports (%d total probes)\n\n",
			len(s.targets), len(s.ports), totalTasks)
	}

	err := s.generateTasks(ctx)
	
	// Close the worker pool task channel
	s.pool.Close()

	// Wait for progress reporter to finish
	if s.config.Verbose {
		progressWg.Wait()
	}

	// Close the result channel and wait for collection to complete
	close(s.resultChan)
	s.collector.Close()

	// Write final results
	if err := s.collector.WriteResults(); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	return err
}

// generateTasks creates and submits scanning tasks to the worker pool
func (s *Scanner) generateTasks(ctx context.Context) error {
	for _, target := range s.targets {
		for _, port := range s.ports {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Apply rate limiting if configured
			if s.rateLimiter != nil {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-s.rateLimiter.C:
					// Continue after rate limit tick
				}
			}

			// Submit task to worker pool
			task := worker.Task{
				IP:   target,
				Port: port,
			}

			if err := s.pool.Submit(ctx, task); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectResults receives results from workers and submits them to the collector
func (s *Scanner) collectResults() {
	for r := range s.resultChan {
		s.collector.Submit(r)
	}
}

// reportProgress periodically reports scan progress
func (s *Scanner) reportProgress(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	totalTasks := len(s.targets) * len(s.ports)
	lastCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			summary := s.collector.GetSummary()
			currentCount := summary.TotalScanned
			rate := float64(currentCount-lastCount) / 5.0 // scans per second

			progress := float64(currentCount) / float64(totalTasks) * 100
			
			fmt.Printf("\rProgress: %d/%d (%.1f%%) | Open: %d | Rate: %.0f scans/sec",
				currentCount, totalTasks, progress, summary.OpenPorts, rate)

			lastCount = currentCount

			// Stop reporting when done
			if currentCount >= totalTasks {
				fmt.Println() // New line after progress
				return
			}
		}
	}
}

// Stop gracefully stops the scanner
func (s *Scanner) Stop() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
}

// GetSummary returns the scan summary
func (s *Scanner) GetSummary() result.Summary {
	return s.collector.GetSummary()
}

// GetResults returns all scan results
func (s *Scanner) GetResults() []*result.Result {
	return s.collector.GetResults()
}