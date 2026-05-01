package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/nithsua/http-scratch/internal/request"
	respone "github.com/nithsua/http-scratch/internal/response"
)

type ServerState int

const (
	Initialized ServerState = iota
	Accepting
	Closed
)

type Server struct {
	handler     Handler
	listener    net.Listener
	serverState ServerState
}

func Serve(port int, handler Handler) (*Server, error) {
	server := &Server{serverState: Closed, handler: handler}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		if listener != nil {
			listener.Close()
		}
		return nil, err
	}
	server.listener = listener
	server.serverState = Initialized

	go server.listen()
	return server, err
}

func (s *Server) Close() error {
	err := s.listener.Close()
	s.serverState = Closed
	return err
}

func (s *Server) listen() {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		s.serverState = Accepting
		go s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer closeConnection(conn)
	responseWriter := bufio.NewWriter(conn)

	request, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := HandlerError{StatusCode: respone.InternalServerError, Message: err.Error()}
		writeError(responseWriter, handlerError)
	}
	printRequest(request)

	responseBodyWriter := new(bytes.Buffer)
	handlerError := s.handler(responseBodyWriter, request)
	if handlerError != nil {
		writeError(responseWriter, *handlerError)
	} else {
		if err := respone.WriteStatusLine(responseWriter, respone.Ok); err != nil {
			log.Println("Error writing statusLine to response", err)
			return
		}

		headers := respone.GetDefaultHeaders(responseBodyWriter.Len())
		if err := respone.WriteHeaders(responseWriter, headers); err != nil {
			log.Println("Error writing headers to response", err)
			return
		}

		if _, err := fmt.Fprint(responseWriter, responseBodyWriter); err != nil {
			log.Println("Error writing body to response")
		}
	}

	if err := responseWriter.Flush(); err != nil {
		log.Println("Error while headers to response", err)
	}
}

func printRequest(request *request.Request) {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", request.RequestLine.Method)
	fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range request.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
	fmt.Println("Body:")
	fmt.Println(string(request.Body))
}

func writeError(w io.Writer, error HandlerError) {
	if err := respone.WriteStatusLine(w, error.StatusCode); err != nil {
		log.Println("Error writing statusLine to response", err)
		return
	}

	headers := respone.GetDefaultHeaders(len(error.Error()))
	if err := respone.WriteHeaders(w, headers); err != nil {
		log.Println("Error writing headers to response", err)
		return
	}

	if _, err := fmt.Fprint(w, error.Error()); err != nil {
		log.Println("Error writing body to response")
	}
}

func closeConnection(conn net.Conn) {
	conn.Close()
	fmt.Println("Connection closed")
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode respone.StatusCode
	Message    string
}

func (h HandlerError) Error() string {
	return h.Message
}
