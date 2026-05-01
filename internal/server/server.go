package server

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/nithsua/http-scratch/internal/request"
)

type ServerState int

const (
	Initialized ServerState = iota
	Listening
	Accepting
	Closed
)

type Server struct {
	listener    net.Listener
	serverState ServerState
}

func Serve(port int) (*Server, error) {
	server := &Server{}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		if listener != nil {
			listener.Close()
		}
	} else {
		server.listener = listener
		go server.listen()
	}

	return server, err
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			log.Fatal("Error while accepting connection")
		}
		go func() {
			request, _ := request.RequestFromReader(connection)
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
			s.handle(connection)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	responseString := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!"
	conn.Write([]byte(responseString))
	conn.Close()
	fmt.Println("Connection closed")
}
