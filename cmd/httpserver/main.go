package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nithsua/http-scratch/internal/request"
	respone "github.com/nithsua/http-scratch/internal/response"
	"github.com/nithsua/http-scratch/internal/server"
)

const port = 42069

const indexBody = `Routes:
  GET  /                       this index
  GET  /headers                dumps request headers
  POST /echo                   echoes the request body
  GET  /status/{200|400|500}   returns the given status
`

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	target := req.RequestLine.RequestTarget
	method := req.RequestLine.Method

	switch {
	case method == "GET" && target == "/":
		_, _ = io.WriteString(w, indexBody)
		return nil

	case method == "GET" && target == "/headers":
		for k, v := range req.Headers {
			fmt.Fprintf(w, "%s: %s\n", k, v)
		}
		return nil

	case method == "POST" && target == "/echo":
		_, _ = w.Write(req.Body)
		return nil

	case method == "GET" && strings.HasPrefix(target, "/status/"):
		switch strings.TrimPrefix(target, "/status/") {
		case "200":
			_, _ = io.WriteString(w, "OK\n")
			return nil
		case "400":
			return &server.HandlerError{
				StatusCode: respone.BadRequest,
				Message:    "Bad Request\n",
			}
		case "500":
			return &server.HandlerError{
				StatusCode: respone.InternalServerError,
				Message:    "Internal Server Error\n",
			}
		default:
			return &server.HandlerError{
				StatusCode: respone.BadRequest,
				Message:    "unsupported status code\n",
			}
		}

	default:
		return &server.HandlerError{
			StatusCode: respone.BadRequest,
			Message:    "not found\n",
		}
	}
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
