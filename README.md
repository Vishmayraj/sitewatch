# sitewatch

A lightweight, zero-dependency CLI tool written in Go that continuously monitors a website and reports uptime, latency, and failures in real time — with structured file logging and graceful shutdown.

Built to demonstrate idiomatic Go: goroutines, channels, context cancellation, mutexes, and a clean internal package architecture.

---

## Demo

```
$ sitewatch --interval 3s --timeout 5s https://example.com

Monitoring https://example.com every 3s (timeout: 5s)

[18:03:27] 200 OK        553ms
[18:03:30] 200 OK        215ms
[18:03:33] 200 OK        185ms
[18:03:36] 200 OK        155ms
[18:03:39] ERROR         Get "https://example.com": context deadline exceeded (Client.Timeout exceeded)
[18:03:42] 200 OK        141ms
^C
Shutting down...

--- Monitoring stopped ---
Total checks : 6
Successes    : 5
Failures     : 1
Uptime       : 83.33%
Latency min  : 141ms
Latency avg  : 250ms
Latency max  : 553ms
```

---

## Features

- **Continuous monitoring** — checks your target URL at a fixed interval forever, until you stop it
- **Real-time output** — every check prints a timestamped result to stdout immediately
- **Latency tracking** — measures actual round-trip time per request including DNS, TCP handshake, and response headers
- **Timeout handling** — configurable per-request timeout so hanging servers don't stall the monitor
- **Running statistics** — tracks total checks, successes, failures, uptime percentage, and min/avg/max latency across the entire session
- **Graceful shutdown** — CTRL+C triggers a clean exit with a full session summary printed before the program stops
- **File logging** — optional `--log` flag writes structured RFC3339-timestamped logs to disk, appending across multiple runs
- **Concurrent by design** — each HTTP check runs in its own goroutine so the monitor loop never freezes, even on slow or unresponsive servers

---

## Installation

### Download the binary (Windows)

Download the latest `sitewatch.exe` from the [Releases](../../releases) page. No installation required — it is a single self-contained executable with no dependencies.

Place it anywhere on your system. To use it from any directory, add its location to your `PATH`:

1. Move `sitewatch.exe` to a permanent location, e.g. `C:\Tools\`
2. Open **System Properties** → **Environment Variables**
3. Under **System variables**, find `Path` and click **Edit**
4. Click **New** and add `C:\Tools\`
5. Click OK and restart your terminal

Then run it from anywhere:

```
sitewatch --interval 5s https://example.com
```

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/Vishmayraj/sitewatch
cd sitewatch
go build -o sitewatch.exe ./cmd/sitewatch
```

---

## Usage

```
sitewatch [flags] <url>
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--interval` | `5s` | How often to check the URL. Accepts Go duration strings: `1s`, `30s`, `1m`, `5m` |
| `--timeout` | `10s` | Maximum time to wait for a response before marking the check as failed |
| `--log` | _(disabled)_ | Path to a log file. Created if it does not exist. Appends across runs |

### Examples

Monitor every 5 seconds with default timeout:
```
sitewatch https://example.com
```

Monitor every 10 seconds with a 3 second timeout:
```
sitewatch --interval 10s --timeout 3s https://example.com
```

Monitor and write logs to a file:
```
sitewatch --interval 5s --timeout 5s --log monitor.log https://example.com
```

Monitor a local development server:
```
sitewatch --interval 1s --timeout 2s http://localhost:8080
```

---

## Log File Format

When `--log` is specified, every check is written to the file in this format:

```
2026-03-14T18:03:27+05:30 200 553ms
2026-03-14T18:03:30+05:30 200 215ms
2026-03-14T18:03:33+05:30 200 185ms
2026-03-14T18:03:39+05:30 ERROR Get "https://example.com": context deadline exceeded
```

On shutdown, the session summary is also appended:

```
--- Monitoring stopped ---
Total checks : 6
Successes    : 5
Failures     : 1
Uptime       : 83.33%
Latency min  : 141ms
Latency avg  : 250ms
Latency max  : 553ms
```

Timestamps use [RFC3339](https://www.rfc-editor.org/rfc/rfc3339) format with local timezone offset. Log files are safe to keep across multiple runs — each new session appends to the existing file without overwriting previous data.

---

## Project Structure

```
sitewatch/
├── cmd/
│   └── sitewatch/
│       └── main.go                 # Entry point, CLI flags, graceful shutdown wiring
├── internal/
│   ├── checker/
│   │   ├── http_checker.go         # HTTP GET with timeout, latency measurement
│   │   └── http_checker_test.go
│   ├── monitor/
│   │   ├── monitor.go              # Ticker loop, goroutines, channels, output, logging
│   │   └── monitor_test.go
│   └── stats/
│       ├── stats.go                # Thread-safe running statistics
│       └── stats_test.go
├── pkg/
│   ├── types.go                    # Shared Result struct
│   └── types_test.go
├── go.mod
└── README.md
```

All packages under `internal/` are private to this module by Go's enforced visibility rules. No external library can import them.

---

## How It Works

```
main
 └── creates Monitor
      └── starts Run() loop
           ├── time.Ticker fires every interval
           │    └── goroutine spawned → HTTP GET → result sent to channel
           ├── result received from channel → Stats.Record() → printed to stdout + log
           └── context cancelled (CTRL+C) → final stats printed → clean exit
```

Each HTTP check runs in its own goroutine so the main loop stays fully responsive. A slow or hanging server cannot block the ticker, delay output, or prevent CTRL+C from working. The stats engine uses a `sync.Mutex` to protect shared state across concurrent goroutines.

---

## Running the Tests

```bash
go test -v ./...
```

Tests use Go's `net/http/httptest` package to spin up real local HTTP servers — no mocking libraries, no internet connection required. The test suite covers success, timeout, non-2xx responses, HEAD fallback, stats calculation edge cases, and monitor lifecycle.

To bypass Go's test cache and force a full rerun:

```bash
go test -count=1 -v ./...
```

---

## Known Limitations

- Monitors a single URL per invocation. Run multiple instances in separate terminals to monitor multiple targets.
- Uses HTTP GET for all checks. Some monitoring tools use HEAD to reduce bandwidth — GET was chosen here for maximum server compatibility since not all servers implement HEAD correctly.
- No retry logic. A single failed check counts as a failure immediately. This is intentional for accurate failure detection — retries would mask real downtime.
- No alert notifications. Output is stdout and log file only. Pipe to a notification tool if needed.

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Author

Built by [Vishmayraj](https://github.com/Vishmayraj).
