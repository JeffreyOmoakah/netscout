package result

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"
)

// Status represents the state of a scanned port
type Status string

const (
	StatusOpen     Status = "open"
	StatusClosed   Status = "closed"
	StatusFiltered Status = "filtered"
	StatusError    Status = "error"
)

// Result represents a single scan result
type Result struct {
	IP        string        `json:"ip"`
	Port      int           `json:"port"`
	Status    Status        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// Summary contains aggregated scan statistics
type Summary struct {
	TotalScanned int
	OpenPorts    int
	ClosedPorts  int
	Filtered     int
	Errors       int
	Duration     time.Duration
	StartTime    time.Time
	EndTime      time.Time
}

// Collector collects and manages scan results
type Collector struct {
	mu         sync.Mutex
	results    []*Result
	summary    Summary
	resultChan chan *Result
	doneChan   chan struct{}
	writer     io.Writer
	format     string
	verbose    bool
}

// NewCollector creates a new result collector
func NewCollector(outputFile, format string, verbose bool) (*Collector, error) {
	var writer io.Writer = os.Stdout

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		writer = f
	}

	c := &Collector{
		results:    make([]*Result, 0),
		resultChan: make(chan *Result, 1000), // Buffered to prevent blocking
		doneChan:   make(chan struct{}),
		writer:     writer,
		format:     format,
		verbose:    verbose,
		summary: Summary{
			StartTime: time.Now(),
		},
	}

	// Start the collector goroutine
	go c.collect()

	return c, nil
}

// Submit submits a result to the collector
func (c *Collector) Submit(r *Result) {
	c.resultChan <- r
}

// collect runs in a goroutine and processes results
func (c *Collector) collect() {
	for r := range c.resultChan {
		c.mu.Lock()
		c.results = append(c.results, r)
		c.updateSummary(r)
		c.mu.Unlock()

		// Print result immediately in verbose mode (text format only)
		if c.verbose && c.format == "text" {
			c.printResult(r)
		}
	}
	close(c.doneChan)
}

// updateSummary updates the summary statistics
func (c *Collector) updateSummary(r *Result) {
	c.summary.TotalScanned++
	switch r.Status {
	case StatusOpen:
		c.summary.OpenPorts++
	case StatusClosed:
		c.summary.ClosedPorts++
	case StatusFiltered:
		c.summary.Filtered++
	case StatusError:
		c.summary.Errors++
	}
}

// printResult prints a single result (for verbose mode)
func (c *Collector) printResult(r *Result) {
	if r.Status == StatusOpen {
		fmt.Fprintf(c.writer, "[+] %s:%d - %s\n", r.IP, r.Port, r.Status)
	}
}

// Close closes the collector and waits for all results to be processed
func (c *Collector) Close() {
	close(c.resultChan)
	<-c.doneChan
	c.summary.EndTime = time.Now()
	c.summary.Duration = c.summary.EndTime.Sub(c.summary.StartTime)
}

// WriteResults writes all collected results in the specified format
func (c *Collector) WriteResults() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Sort results by IP and port
	sort.Slice(c.results, func(i, j int) bool {
		if c.results[i].IP != c.results[j].IP {
			return c.results[i].IP < c.results[j].IP
		}
		return c.results[i].Port < c.results[j].Port
	})

	switch c.format {
	case "json":
		return c.writeJSON()
	case "csv":
		return c.writeCSV()
	case "text":
		return c.writeText()
	default:
		return fmt.Errorf("unsupported format: %s", c.format)
	}
}

// writeJSON writes results in JSON format
func (c *Collector) writeJSON() error {
	encoder := json.NewEncoder(c.writer)
	encoder.SetIndent("", "  ")

	output := map[string]interface{}{
		"summary": c.summary,
		"results": c.results,
	}

	return encoder.Encode(output)
}

// writeCSV writes results in CSV format
func (c *Collector) writeCSV() error {
	writer := csv.NewWriter(c.writer)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"IP", "Port", "Status", "Timestamp", "Duration", "Error"}); err != nil {
		return err
	}

	// Write results
	for _, r := range c.results {
		record := []string{
			r.IP,
			fmt.Sprintf("%d", r.Port),
			string(r.Status),
			r.Timestamp.Format(time.RFC3339),
			r.Duration.String(),
			r.Error,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// writeText writes results in human-readable text format
func (c *Collector) writeText() error {
	// If verbose mode was on, results were already printed
	if c.verbose {
		return nil
	}

	// Otherwise, print all results now
	for _, r := range c.results {
		if r.Status == StatusOpen {
			fmt.Fprintf(c.writer, "%s:%d - %s\n", r.IP, r.Port, r.Status)
		}
	}

	return nil
}

// GetSummary returns the current summary statistics
func (c *Collector) GetSummary() Summary {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.summary
}

// GetResults returns all collected results
func (c *Collector) GetResults() []*Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	resultsCopy := make([]*Result, len(c.results))
	copy(resultsCopy, c.results)
	return resultsCopy
}