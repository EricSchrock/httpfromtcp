package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	header, _, found := strings.Cut(string(data), "\r\n")
	if !found {
		return 0, false, nil
	} else if len(header) == 0 { // headers section ends with blank line
		return 2, true, nil // account for \r\n
	}

	name, value, found := strings.Cut(header, ":")
	if !found {
		return 0, false, fmt.Errorf("Expected header to be a name:value pair): header='%v'", header)
	}

	if name != strings.TrimSpace(name) {
		return 0, false, fmt.Errorf("Header field name has leading or trailing whitespace: name='%v'", name)
	}

	if _, exists := h[name]; exists {
		return 0, false, fmt.Errorf("Duplicate header field name: name='%v'", name)
	}

	h[name] = strings.TrimSpace(value)

	return len(header) + 2, false, nil // acount for \r\n
}
