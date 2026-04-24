package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// // RFC 9110
// // Sementics for all HTTP versions - 1.0, 1.1, 2.0

// // RFC 9112
// // Specifically focuses on HTTP 1.1
// //
// // Fieldline by RFC basically denoted Headers

func getLinesChannel(connection net.Conn) <-chan string {
	channel := make(chan string)

	go func() {
		defer close(channel)
		defer connection.Close()

		lineBuffer := ""
		for {
			byteBuffer := make([]byte, 8)
			n, err := connection.Read(byteBuffer)
			if n < 8 || err == io.EOF {
				channel <- string(lineBuffer + string(byteBuffer[:n]))
				break
			}
			if err != nil {
				log.Fatal("Error while reading file")
			}

			if index := strings.IndexByte(string(byteBuffer), '\n'); index != -1 {
				channel <- string(lineBuffer + string(byteBuffer[:index]))
				byteBuffer = byteBuffer[index+1:]
				lineBuffer = ""
			}
			lineBuffer += string(byteBuffer)
		}
	}()

	return channel
}

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

		lineChannel := getLinesChannel(connection)
		for line := range lineChannel {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("Connection closed")
	}
}
