package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/nithsua/http-scratch/internal/request"
	respone "github.com/nithsua/http-scratch/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	request, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := HandlerError{StatusCode: respone.InternalServerError, Message: err.Error()}
		writeError(conn, handlerError)
	}
	printRequest(request)

	responseBodyWriter := new(bytes.Buffer)
	handlerError := s.handler(responseBodyWriter, request)
	if handlerError != nil {
		writeError(conn, *handlerError)
		return
	}

	respone.WriteStatusLine(conn, respone.Ok)
	headers := respone.GetDefaultHeaders(responseBodyWriter.Len())
	respone.WriteHeaders(conn, headers)
	fmt.Fprint(conn, responseBodyWriter)
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

func closeConnection(conn net.Conn) {
	conn.Close()
	fmt.Println("Connection closed")
}

type HandlerError struct {
	StatusCode respone.StatusCode
	Message    string
}

func (h HandlerError) writeError(w io.Writer) {
	respone.WriteStatusLine(w, h.StatusCode)
	headers := respone.GetDefaultHeaders(len(h.Error()))
	respone.WriteHeaders(w, headers)
	fmt.Fprint(w, h.Error())
}

func (h HandlerError) Error() string {
	return h.Message
}
