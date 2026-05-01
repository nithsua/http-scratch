package errors

import "errors"

var ErrNoCRLF = errors.New("Unable to find CRLF")

var ErrInvalidFieldLine = errors.New("Invalid FieldLine")
var ErrNoFieldName = errors.New("Invalid FieldName provided")
var ErrInvalidFieldName = errors.New("Invalid FieldName provided")
var ErrIncomplteRequestBody = errors.New("Incomplete request body")
