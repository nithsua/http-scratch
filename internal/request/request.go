package request

import (
	"errors"
	"io"
	"slices"
	"strings"

	internalError "github.com/nithsua/tcp-scratch/internal/errors"
	"github.com/nithsua/tcp-scratch/internal/headers"
)

type ParserState int

const (
	Initialized ParserState = iota
	RequestStateParsingHeaders
	Done
)

const bufferLength int = 1024

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	parserState ParserState
}

func (r *Request) parse(data []byte) (int, error) {
	r.parserState = Initialized
	var n int = 0
	var err error = nil

	switch r.parserState {
	case Initialized:
		requestLine := RequestLine{}
		n, err = parseRequestLine(data, &requestLine)
		if n != 0 && err == nil {
			r.RequestLine = requestLine
			r.parserState = RequestStateParsingHeaders
		}
	case RequestStateParsingHeaders:
		headers := headers.NewHeaders()
		n, err = parseHeaders(data, &headers)
		if n == 0 && err == nil {
			r.Headers = headers
			r.parserState = Done
		}
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
	requestBuffer := make([]byte, bufferLength)
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
		if bufferParsedSize == 0 && err == internalError.ErrNoCRLF {
			requestBuffer = slices.Grow(requestBuffer, filled)
			requestBuffer = requestBuffer[:len(requestBuffer)+filled]
			continue
		} else if err != nil {
			return nil, err
		}
		requestBuffer = make([]byte, bufferLength)
		filled = 0
		parsedSize += bufferParsedSize
	}

	return request, nil
}

func parseRequestLine(buffer []byte, requestLine *RequestLine) (int, error) {
	requestLineEOLIndex := strings.Index(string(buffer), "\r\n")
	if requestLineEOLIndex == -1 {
		return 0, internalError.ErrNoCRLF
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

func parseHeaders(data []byte, headers *headers.Headers) (n int, err error) {
	n, done, err := headers.Parse(data)
	if err != nil {
		return n, internalError.ErrNoCRLF
	}

	if done == true {
		return 0, nil
	}

	return n, nil
}
