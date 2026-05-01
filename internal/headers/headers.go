package headers

import (
	"fmt"
	"log"
	"slices"
	"strings"

	internalError "github.com/nithsua/http-scratch/internal/errors"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIndex := strings.Index(string(data), "\r\n")
	if crlfIndex == -1 {
		return 0, false, internalError.ErrNoCRLF
	}
	if crlfIndex == 0 {
		return 2, true, nil
	}

	seperatorIndex := strings.Index(string(data[:crlfIndex]), ":")
	if seperatorIndex == -1 || seperatorIndex == 0 {
		log.Println("No key or value provided")
		return 0, false, internalError.ErrInvalidFieldLine
	}

	fieldLine := strings.SplitN(strings.TrimSpace(string(data[:crlfIndex])), ":", 2)
	key := fieldLine[0]
	err = validateFieldName([]byte(key))
	if err != nil {
		return 0, false, err
	}

	value := strings.TrimSpace(string(fieldLine[1]))
	if _, ok := h[strings.ToLower(key)]; !ok {
		h[strings.ToLower(key)] = value
	} else {
		h[strings.ToLower(key)] += fmt.Sprintf(", %s", value)
	}

	return crlfIndex + 2, false, nil
}

var allowedSpecialAscii = [256]uint8{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func validateFieldName(fieldName []byte) error {
	key := ""
	i := 0
	ch := byte(0)

	for i, ch = range fieldName {
		if !(slices.Contains(allowedSpecialAscii[:], uint8(ch)) ||
			(ch >= 'a' && ch <= 'z' || // lowercase
				ch >= 'A' && ch <= 'Z' || // uppercase
				ch >= '0' && ch <= '9')) { // numerical
			return internalError.ErrInvalidFieldName
		}
		key += string(ch)
	}

	if i == 0 {
		return internalError.ErrNoFieldName
	}

	if strings.TrimSpace(key) != key {
		log.Println("Found whitespace between the key")
		return internalError.ErrInvalidFieldLine
	}

	return nil
}
