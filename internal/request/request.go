package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// Eg: "GET / HTTP/1.1
func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(errors.New("Error while reading from reader"), err)
	}

	index := strings.Index(string(buffer), "\r\n")
	if index == -1 {
		return nil, errors.New("Unable to find CRLF")
	}

	buffer = buffer[:index]
	parts := strings.Split(string(buffer), " ")
	if len(parts) != 3 {
		return nil, errors.New("Request line malformed")
	}

	method := parts[0]
	switch method {
	case "GET", "POST", "PATCH", "DELETE", "OPTION", "PUT":
	default:
		return nil, errors.New("Invalid HTTP method provided")
	}

	requestTarget := parts[1]

	httpVersion := parts[2]
	httpParts := strings.Split(httpVersion, "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, errors.New("Invalid HTTP version provided in request line")
	}

	requestLine := RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpParts[1],
	}
	request := &Request{
		RequestLine: requestLine,
	}

	return request, nil
}
