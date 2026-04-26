package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nithsua/tcp-scratch/internal/request"
)

// // RFC 9110
// // Sementics for all HTTP versions - 1.0, 1.1, 2.0

// // RFC 9112
// // Specifically focuses on HTTP 1.1
// //
// // Fieldline by RFC basically denoted Headers
func main() {
	port := ":42069"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error while creating a listener")
	}
	fmt.Println("Accepting connection at {}", port)

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal("Error while accepting connection")
		}

		request, err := request.RequestFromReader(connection)
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Connection closed")
	}
}
