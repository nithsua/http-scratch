# http-scratch

A from-scratch HTTP/1.1 request parser built on raw TCP, written in Go. This project explores low-level HTTP protocol handling at the socket level, following [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110) (HTTP Semantics) and [RFC 9112](https://www.rfc-editor.org/rfc/rfc9112) (HTTP/1.1).

## What It Does

- Accepts raw TCP connections and parses incoming HTTP/1.1 requests
- Handles chunked/fragmented reads as data arrives over the wire
- Validates request lines, methods, and HTTP version strings
- Parses and validates headers per RFC 9110 field name rules
- Normalizes header keys to lowercase and merges duplicate headers

## Project Structure

```
.
├── cmd/
│   ├── tcplistener/       # TCP server (port 42069)
│   └── udpsender/         # UDP client for sending test data
└── internal/
    ├── errors/            # Custom error types
    ├── headers/           # HTTP header parsing & validation
    └── request/           # HTTP request line & state machine parser
```

## Getting Started

**Prerequisites:** Go 1.22+

### Run the TCP Listener

```bash
go run ./cmd/tcplistener/
# Listening on :42069
```

### Send a Request

```bash
curl -v http://localhost:42069/some/path
```

The listener will print the parsed request — method, target, HTTP version, and all headers.

### Run Tests

```bash
go test ./internal/...
```

## How It Works

Request parsing is implemented as a state machine with three states:

```
Initialized → ParsingHeaders → Done
```

1. **Initialized** — reads and validates the request line (`METHOD /target HTTP/1.1\r\n`)
2. **ParsingHeaders** — reads headers one at a time until the blank `\r\n` line
3. **Done** — returns the fully parsed `Request` struct

The parser uses a growable buffer to handle TCP fragmentation — data can arrive in any size chunks and parsing continues correctly regardless.

### Supported Methods

`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTION`

### Header Parsing

- Keys are normalized to lowercase
- Whitespace around values is trimmed
- Duplicate headers are merged with `", "` separator
- Field names are validated against RFC 9110 token rules

## Dependencies

- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) — test assertions only

Everything else uses the Go standard library (`net`, `io`, `strings`, `slices`).
