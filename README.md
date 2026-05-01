# http-scratch

A from-scratch HTTP/1.1 server built on raw TCP, written in Go. This project explores low-level HTTP protocol handling at the socket level, following [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110) (HTTP Semantics) and [RFC 9112](https://www.rfc-editor.org/rfc/rfc9112) (HTTP/1.1).

## What It Does

- Listens on raw TCP and parses incoming HTTP/1.1 requests
- Handles chunked/fragmented reads as data arrives over the wire
- Validates request lines, methods, and HTTP version strings
- Parses and validates headers per RFC 9110 field-name rules
- Normalizes header keys to lowercase and merges duplicate headers
- Reads request bodies framed by `Content-Length`
- Dispatches parsed requests to user-supplied handlers and writes back a response

## Project Structure

```
.
├── cmd/
│   ├── httpserver/        # HTTP server with routed handlers (port 42069)
│   ├── tcplistener/       # Bare TCP request inspector
│   └── udpsender/         # UDP client for sending test data
└── internal/
    ├── errors/            # Custom error types
    ├── headers/           # HTTP header parsing & validation
    ├── request/           # Request line / headers / body state machine
    ├── response/          # Status line and header writing
    └── server/            # Listener, connection lifecycle, handler dispatch
```

## Getting Started

**Prerequisites:** Go 1.26+

### Run the HTTP Server

```bash
go run ./cmd/httpserver/
# Server started on port 42069
```

### Try the Routes

```bash
curl -i http://localhost:42069/                  # index of routes
curl -i -H 'X-Demo: hi' http://localhost:42069/headers
curl -i -X POST --data 'ping' http://localhost:42069/echo
curl -i http://localhost:42069/status/200
curl -i http://localhost:42069/status/400
curl -i http://localhost:42069/status/500
curl -i http://localhost:42069/whatever          # falls through to 400 "not found"
```

### Run the Inspector (raw request dump)

```bash
go run ./cmd/tcplistener/
```

Useful for seeing exactly what the parser produces from a request — method, target, version, headers, body.

### Run Tests

```bash
go test ./internal/...
```

## How It Works

### Parser State Machine

Request parsing flows through four states:

```
Initialized → ParsingHeaders → ParsingBody → Done
```

1. **Initialized** — reads and validates the request line (`METHOD /target HTTP/1.1\r\n`)
2. **ParsingHeaders** — reads headers one at a time until the blank `\r\n` terminator
3. **ParsingBody** — if `Content-Length` is set, reads exactly that many bytes; otherwise skipped
4. **Done** — returns the fully parsed `Request` struct

The parser uses a growable buffer so TCP fragmentation doesn't matter — data can arrive in any size chunks and parsing continues correctly.

### Supported Methods

`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTION`

### Header Parsing

- Keys are normalized to lowercase
- Whitespace around values is trimmed
- Duplicate headers are merged with `", "` separator
- Field names are validated against RFC 9110 `token` rules

### Body Framing

Per RFC 9112 §6.3, request body length is determined by `Content-Length` alone (chunked transfer encoding isn't implemented yet). If neither header is present, the body is treated as zero-length — any extra bytes on the wire belong to the next request.

### Handler API

`server.Serve` takes a handler with this signature:

```go
type Handler func(w io.Writer, req *request.Request) *HandlerError
```

- Write to `w` to set the response body. The server emits `200 OK` and adds `Content-Length` automatically.
- Return a non-nil `*HandlerError` to send an error status (`400`, `500`) with a message body.

See [`cmd/httpserver/main.go`](cmd/httpserver/main.go) for a worked example covering routing, headers inspection, body echo, and dynamic status responses.

## Dependencies

- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) — test assertions only

Everything else uses the Go standard library (`net`, `io`, `bufio`, `strings`, `slices`, `strconv`).
