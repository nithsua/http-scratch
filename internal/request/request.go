package request

import (
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	internalError "github.com/nithsua/http-scratch/internal/errors"
	"github.com/nithsua/http-scratch/internal/headers"
)

type ParserState int

const (
	Initialized ParserState = iota
	ParsingHeaders
	ParsingBody
	Done
)

const bufferLength int = 1024

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	parserState ParserState
}

func (r *Request) Get(key string) string {
	return r.Headers[strings.ToLower(key)]
}

func (r *Request) parse(data []byte) (int, error) {
	r.parserState = Initialized
	var n int = 0
	var err error = nil

	if r.parserState == Initialized {
		requestLine := RequestLine{}
		n, err = parseRequestLine(data, &requestLine)
		if n != 0 && err == nil {
			r.RequestLine = requestLine
			r.parserState = ParsingHeaders
		}
	}
	if r.parserState == ParsingHeaders {
		headers := headers.NewHeaders()
		unparsedData := data[n:]
		m := 0
		m, err = parseHeaders(unparsedData, headers)
		if err == nil {
			r.Headers = headers
			r.parserState = ParsingBody
			if r.Get("content-length") != "" {
				r.parserState = ParsingBody
			} else {
				r.parserState = Done
			}
		}
		n += m
	}
	if r.parserState == ParsingBody {
		o := 0
		o, err = r.parseBody(data[n:])
		if err == nil {
			r.parserState = Done
		}
		n += o
	}
	return n, err
}

func (r *Request) parseBody(data []byte) (int, error) {
	bodyLength := len(data)
	contentLength, err := strconv.Atoi(r.Get("content-length"))
	if err != nil {
		return 0, err
	}
	if (bodyLength + len(r.Body)) < contentLength {
		return 0, internalError.ErrIncomplteRequestBody
	}
	contentLengthToRead := contentLength - len(r.Body)
	r.Body = append(r.Body, data[:contentLengthToRead]...)
	return contentLengthToRead, nil
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
	var err error = nil
	for request.parserState != Done {
		bufferReadSize := 0
		if err == io.EOF && bufferReadSize == parsedSize && request.parserState != Done {
			return nil, errors.Join(errors.New("EOF reached, Request has not been successfully parsed"), err)
		}
		err = nil
		bufferReadSize, err = reader.Read(requestBuffer[filled:])
		if err != nil && err != io.EOF {
			return nil, errors.Join(errors.New("Error while reading from reader"), err)
		}
		readSize += bufferReadSize
		filled += bufferReadSize

		bufferParsedSize, err := request.parse(requestBuffer[:filled])
		if err == internalError.ErrNoCRLF {
			requestBuffer = slices.Grow(requestBuffer, filled)
			requestBuffer = requestBuffer[:len(requestBuffer)+filled]
			continue
		} else if err == internalError.ErrIncomplteRequestBody {
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

	return requestLineEOLIndex + 2, nil
}

func parseHeaders(data []byte, headers headers.Headers) (int, error) {
	dataLength := len(data)
	parsedData := 0
	for dataLength != parsedData {
		n, done, err := headers.Parse(data[parsedData:])
		parsedData += n
		if err != nil {
			return parsedData, err
		}
		if done == true {
			return parsedData, nil
		}
	}
	// This happens only when out of loop and done is false
	return parsedData, internalError.ErrNoCRLF
}
