package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

type ParserState int

const (
	Initialized ParserState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	parserState ParserState
}

func (r *Request) parse(data []byte) (int, error) {
	r.parserState = Initialized

	requestLine := RequestLine{}
	n, err := parseRequestLine(data, &requestLine)
	if n != 0 && err == nil {
		r.RequestLine = requestLine
		r.parserState = Done
	}

	return n, err
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// Eg: "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}
	readSize := 0
	parsedSize := 0

	filled := 0
	requestBuffer := make([]byte, 8)
	for request.parserState != Done {
		bufferReadSize := 0
		var err error = nil
		bufferReadSize, err = reader.Read(requestBuffer[filled:])
		if err != nil && err != io.EOF {
			return nil, errors.Join(errors.New("Error while reading from reader"), err)
		}
		readSize += bufferReadSize
		filled += bufferReadSize

		bufferParsedSize, err := request.parse(requestBuffer)
		if bufferParsedSize == 0 && err.Error() == "Unable to find CRLF" {
			requestBuffer = slices.Grow(requestBuffer, filled)
			requestBuffer = requestBuffer[:len(requestBuffer)+filled]
			continue
		} else if err != nil {
			return nil, err
		}
		requestBuffer = make([]byte, 8)
		filled = 0
		parsedSize += bufferParsedSize
	}

	return request, nil
}

func parseRequestLine(buffer []byte, requestLine *RequestLine) (int, error) {
	stringval := string(buffer)
	fmt.Println(stringval)
	requestLineEOLIndex := strings.Index(string(buffer), "\r\n")
	if requestLineEOLIndex == -1 {
		return 0, errors.New("Unable to find CRLF")
	}
	requestLineByte := buffer[:requestLineEOLIndex]
	parts := strings.Split(string(requestLineByte), " ")
	if len(parts) != 3 {
		return 0, errors.New("Request line malformed")
	}

	method := parts[0]
	switch method {
	case "GET", "POST", "PATCH", "DELETE", "OPTION", "PUT":
	default:
		return requestLineEOLIndex, errors.New("Invalid HTTP method provided")
	}
	requestLine.Method = method

	requestTarget := parts[1]
	requestLine.RequestTarget = requestTarget

	httpVersion := parts[2]
	httpParts := strings.Split(httpVersion, "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return requestLineEOLIndex, errors.New("Invalid HTTP version provided in request line")
	}
	requestLine.HttpVersion = httpParts[1]

	return requestLineEOLIndex, nil
}
