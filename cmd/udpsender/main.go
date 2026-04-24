package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

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
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost"+port)
	if err != nil {
		log.Fatal("Error while resolving UDP for address")
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Error while dialing up UDP at addr")
	}
	defer udpConn.Close()
	buffer := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		line, err := buffer.ReadSlice('\n')
		if err != nil {
			log.Println("Error while reading from Stdio")
		}
		_, err = udpConn.Write(line)
		if err != nil {
			log.Println("Err while writing to the conn")
		}
	}
}
