package respone

import (
	"fmt"
	"io"
	"strconv"

	"github.com/nithsua/http-scratch/internal/headers"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error = nil
	if statusCode == Ok {
		_, err = w.Write([]byte("HTTP/1.1 200 OK "))
	} else if statusCode == BadRequest {
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request "))
	} else {
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error "))
	}
	_, err = w.Write([]byte("\r\n"))
	return err
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	headers := headers.Headers{}
	headers["content-length"] = strconv.Itoa(contentLength)
	headers["content-type"] = "text/plain"
	headers["connection"] = "close"
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	var err error = nil
	for fieldName, fieldValue := range headers {
		_, err = w.Write([]byte(fmt.Sprintf("%s: %s\r\n", fieldName, fieldValue)))
	}
	w.Write([]byte("\r\n"))
	return err
}
