package server

import (
	"bufio"
	"fmt"
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
	listener    net.Listener
	serverState ServerState
}

func Serve(port int) (*Server, error) {
	server := &Server{serverState: Closed}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		if listener != nil {
			listener.Close()
		}
	} else {
		server.listener = listener
		server.serverState = Initialized
		go server.listen()
	}

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
		s.serverState = Accepting
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
	defer closeConnection(conn)
	w := bufio.NewWriter(conn)

	if err := respone.WriteStatusLine(w, respone.Ok); err != nil {
		log.Println("Error writing statusLine to response", err)
		return
	}

	headers := respone.GetDefaultHeaders(0)
	if err := respone.WriteHeaders(w, headers); err != nil {
		log.Println("Error writing headers to response", err)
		return
	}

	if err := w.Flush(); err != nil {
		log.Println("Error while headers to response", err)
	}
}

func closeConnection(conn net.Conn) {
	conn.Close()
	fmt.Println("Connection closed")
}
