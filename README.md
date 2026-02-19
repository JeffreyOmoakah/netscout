# netscout

A fast, concurrent TCP port scanner built in Go. Scan single hosts or entire CIDR ranges with configurable worker pools, rate limiting, and multiple output formats.

## Features

- **Concurrent scanning** — configurable worker pool (up to 10,000 goroutines)
- **CIDR support** — expand `192.168.1.0/24` into individual host scans
- **Flexible port specs** — single ports (`80`), lists (`80,443`), and ranges (`8000-9000`)
- **Rate limiting** — cap requests per second to avoid flooding
- **Multiple output formats** — text, JSON, CSV
- **Graceful shutdown** — handles `SIGINT`/`SIGTERM` cleanly
- **Progress reporting** — real-time scan rate and completion percentage
- **Cross-platform** — builds for Linux, macOS (Intel + Apple Silicon), and Windows

## Installation

### From source

```sh path=null start=null
git clone https://github.com/JeffreyOmoakah/netscout.git
cd netscout
make build
```

The binary will be at `bin/netscout`.

### Go install

```sh path=null start=null
go install github.com/JeffreyOmoakah/netscout.git/cmd/netscout@latest
```

## Usage

```sh path=null start=null
netscout -t <target> [options]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t` | *(required)* | Target IP or CIDR range (comma-separated) |
| `-p` | `80,443` | Ports to scan (e.g. `80,443` or `1-1024`) |
| `-w` | `100` | Number of concurrent workers |
| `-timeout` | `2s` | Connection timeout |
| `-rate` | `0` | Rate limit in requests/sec (`0` = unlimited) |
| `-o` | stdout | Output file path |
| `-f` | `text` | Output format: `text`, `json`, `csv` |
| `-v` | `false` | Verbose output with progress |
| `-version` | | Print version and exit |

### Examples

Scan common ports on a single host:

```sh path=null start=null
netscout -t 192.168.1.1 -p 22,80,443,8080
```

Scan an entire subnet with verbose output:

```sh path=null start=null
netscout -t 10.0.0.0/24 -p 80,443 -v
```

Scan a port range, rate-limited, output to JSON:

```sh path=null start=null
netscout -t 192.168.1.1 -p 1-1024 -rate 500 -f json -o results.json
```

Scan with more workers and a longer timeout:

```sh path=null start=null
netscout -t 10.0.0.0/24 -p 22,80,443,3306,5432,8080 -w 500 -timeout 5s -v
```

## Project Structure

```
netscout/
├── cmd/netscout/          # CLI entrypoint
│   └── main.go
├── internal/
│   ├── config/            # Configuration and validation
│   ├── parser/            # IP/CIDR and port parsing
│   ├── result/            # Result collection and output formatting
│   ├── scanner/           # Scan orchestration and progress reporting
│   └── worker/            # Worker pool and TCP connect scanning
├── Makefile
└── go.mod
```

## Development

```sh path=null start=null
make build       # Build the binary
make run         # Build and run with example args
make dev         # Build and run in dev mode (verbose, more ports)
make test        # Run tests
make fmt         # Format code
make vet         # Run go vet
make build-all   # Cross-compile for linux/darwin/windows
make clean       # Remove build artifacts
```

## License

This project is open source. See the repository for license details.
